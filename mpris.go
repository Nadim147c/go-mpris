package mpris

import (
	"strings"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/spf13/cast"
)

const (
	dbusObjectPath          = "/org/mpris/MediaPlayer2"
	PropertiesChangedSignal = "org.freedesktop.DBus.Properties.PropertiesChanged"

	BaseInterface      = "org.mpris.MediaPlayer2"
	PlayerInterface    = "org.mpris.MediaPlayer2.Player"
	TrackListInterface = "org.mpris.MediaPlayer2.TrackList"
	PlaylistsInterface = "org.mpris.MediaPlayer2.Playlists"

	getPropertyMethod = "org.freedesktop.DBus.Properties.Get"
	setPropertyMethod = "org.freedesktop.DBus.Properties.Set"
)

// List lists the available players.
func List(conn *dbus.Conn) ([]string, error) {
	var names []string
	err := conn.BusObject().Call("org.freedesktop.DBus.ListNames", 0).Store(&names)
	if err != nil {
		return nil, err
	}

	var mprisNames []string
	for _, name := range names {
		if strings.HasPrefix(name, BaseInterface) {
			mprisNames = append(mprisNames, name)
		}
	}
	return mprisNames, nil
}

// Player represents a mpris player.
type Player struct {
	conn *dbus.Conn
	obj  *dbus.Object
	name string
}

// GetName gets the player full name.
func (i *Player) GetName() string {
	return i.name
}

// CanEditTracks returns if player can edit track list
func (i *Player) CanEditTracks() (bool, error) {
	return getTrackListPropertyCast(i, "CanEditTracks", cast.ToBoolE)
}

// GetLength returns the current track length.
func (i *Player) GetLength() (time.Duration, error) {
	micro, err := getMetadataCast(i, "mpris:length", cast.ToInt64E)
	return time.Duration(micro) * time.Microsecond, err
}

// GetTrackID returns track id for player as dbus.ObjectPath
func (i *Player) GetTrackID() (dbus.ObjectPath, error) {
	trackIdStr, err := getMetadataCast(i, "mpris:trackid", cast.ToStringE)
	return dbus.ObjectPath(trackIdStr), err
}

// GetTitle returns the current track title.
func (i *Player) GetTitle() (string, error) {
	return getMetadataCast(i, "xesam:title", cast.ToStringE)
}

// GetArtist returns the current track artist(s).
func (i *Player) GetArtist() ([]string, error) {
	return getMetadataCast(i, "xesam:artist", cast.ToStringSliceE)
}

// GetAlbum returns the current track album.
func (i *Player) GetAlbum() (string, error) {
	return getMetadataCast(i, "xesam:album", cast.ToStringE)
}

// GetURL returns the URL of the current track.
func (i *Player) GetURL() (string, error) {
	return getMetadataCast(i, "xesam:url", cast.ToStringE)
}

// GetCoverURL returns the cover art URL of the current track.
func (i *Player) GetCoverURL() (string, error) {
	return getMetadataCast(i, "mpris:artUrl", cast.ToStringE)
}

// New connects the the player with the name in the connection conn.
func New(conn *dbus.Conn, name string) *Player {
	obj := conn.Object(name, dbusObjectPath).(*dbus.Object)
	return &Player{conn, obj, name}
}

// OnSignal adds a handler to the player's properties change signal.
//
// Deprecated: Use OnSeeked
func (i *Player) OnSignal(ch chan<- *dbus.Signal) (err error) {
	err = i.conn.AddMatchSignal()
	if err == nil {
		i.conn.Signal(ch)
	}
	return
}
