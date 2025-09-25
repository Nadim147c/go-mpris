package mpris

import (
	"fmt"
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

// Raise raises player priority.
func (i *Player) Raise() error {
	return i.obj.Call(BaseInterface+".Raise", 0).Err
}

// Quit closes the player.
func (i *Player) Quit() error {
	return i.obj.Call(BaseInterface+".Quit", 0).Err
}

// GetIdentity returns the player identity.
func (i *Player) GetIdentity() (string, error) {
	value, err := i.GetBaseProperty("Identity")
	if err != nil {
		return "", err
	}
	return cast.ToStringE(value.Value())
}

// CanPlay returns if player can play
func (i *Player) CanPlay() (bool, error) {
	return getPlayerPropertyCast(i, "CanPlay", cast.ToBoolE)
}

// CanPause returns if player can pause
func (i *Player) CanPause() (bool, error) {
	return getPlayerPropertyCast(i, "CanPause", cast.ToBoolE)
}

// CanSeek returns if player can seek
func (i *Player) CanSeek() (bool, error) {
	return getPlayerPropertyCast(i, "CanSeek", cast.ToBoolE)
}

// CanControl returns if player can control
func (i *Player) CanControl() (bool, error) {
	return getPlayerPropertyCast(i, "CanControl", cast.ToBoolE)
}

// CanGoNext returns if player can go next
func (i *Player) CanGoNext() (bool, error) {
	return getPlayerPropertyCast(i, "CanGoNext", cast.ToBoolE)
}

// CanGoPrevious returns if player can go previous
func (i *Player) CanGoPrevious() (bool, error) {
	return getPlayerPropertyCast(i, "CanGoPrevious", cast.ToBoolE)
}

// CanEditTracks returns if player can edit track list
func (i *Player) CanEditTracks() (bool, error) {
	return getTrackListPropertyCast(i, "CanEditTracks", cast.ToBoolE)
}

// Next skips to the next track in the tracklist.
func (i *Player) Next() error {
	return i.obj.Call(PlayerInterface+".Next", 0).Err
}

// Previous skips to the previous track in the tracklist.
func (i *Player) Previous() error {
	return i.obj.Call(PlayerInterface+".Previous", 0).Err
}

// Pause pauses the current track.
func (i *Player) Pause() error {
	return i.obj.Call(PlayerInterface+".Pause", 0).Err
}

// PlayPause resumes the current track if it's paused and pauses it if it's playing.
func (i *Player) PlayPause() error {
	return i.obj.Call(PlayerInterface+".PlayPause", 0).Err
}

// Stop stops the current track.
func (i *Player) Stop() error {
	return i.obj.Call(PlayerInterface+".Stop", 0).Err
}

// Play starts or resumes the current track.
func (i *Player) Play() error {
	return i.obj.Call(PlayerInterface+".Play", 0).Err
}

// Seek seeks the current track position by the offset.
// If the offset is negative it's seeked back.
func (i *Player) Seek(offset time.Duration) error {
	oms := offset.Microseconds()
	return i.obj.Call(PlayerInterface+".Seek", 0, oms).Err
}

// SetTrackPosition sets the position of a track.
func (i *Player) SetTrackPosition(trackId *dbus.ObjectPath, position time.Duration) error {
	oms := position.Microseconds()
	return i.obj.Call(PlayerInterface+".SetPosition", 0, trackId, oms).Err
}

// OpenUri opens and plays the uri if supported.
//
// Deprecated: Use OpenURI instead of OpenUri
func (i *Player) OpenUri(uri string) error {
	return i.OpenURI(uri)
}

// OpenURI opens and plays the uri if supported.
func (i *Player) OpenURI(uri string) error {
	return i.obj.Call(PlayerInterface+".OpenUri", 0, uri).Err
}

// PlaybackStatus the status of the playback. It can be "Playing", "Paused" or "Stopped".
type PlaybackStatus string

const (
	PlaybackPlaying PlaybackStatus = "Playing"
	PlaybackPaused  PlaybackStatus = "Paused"
	PlaybackStopped PlaybackStatus = "Stopped"
)

// GetPlaybackStatus gets the playback status.
func (i *Player) GetPlaybackStatus() (PlaybackStatus, error) {
	str, err := getPlayerPropertyCast(i, "PlaybackStatus", cast.ToStringE)
	return PlaybackStatus(str), err
}

// LoopStatus the status of the player loop. It can be "None", "Track" or "Playlist".
type LoopStatus string

const (
	LoopNone     LoopStatus = "None"
	LoopTrack    LoopStatus = "Track"
	LoopPlaylist LoopStatus = "Playlist"
)

// GetLoopStatus returns the loop status.
func (i *Player) GetLoopStatus() (LoopStatus, error) {
	str, err := getPlayerPropertyCast(i, "LoopStatus", cast.ToStringE)
	return LoopStatus(str), err
}

// SetLoopStatus sets the loop status to loopStatus.
func (i *Player) SetLoopStatus(loopStatus LoopStatus) error {
	return i.SetPlayerProperty("LoopStatus", loopStatus)
}

// Returns the current playback rate.
func (i *Player) GetRate() (float64, error) {
	return getPlayerPropertyCast(i, "Rate", cast.ToFloat64E)
}

// GetShuffle returns false if the player is going linearly through a playlist and true if it's
// in some other order.
func (i *Player) GetShuffle() (bool, error) {
	return getPlayerPropertyCast(i, "Shuffle", cast.ToBoolE)
}

// SetShuffle sets the shuffle playlist mode.
func (i *Player) SetShuffle(value bool) error {
	return i.SetPlayerProperty("Shuffle", value)
}

type Metadata map[string]dbus.Variant

func (m Metadata) Get(key string) (any, error) {
	v, ok := m[key]
	if !ok || v.Value() == nil {
		return v, fmt.Errorf("%s.Metadata missing or nil for key %q", PlayerInterface, key)
	}
	return v.Value(), nil
}

// GetMetadata returns the metadata.
func (i *Player) GetMetadata() (Metadata, error) {
	return getPlayerPropertyCast(i, "Metadata", func(a any) (Metadata, error) {
		v, ok := a.(map[string]dbus.Variant)
		if !ok {
			return Metadata{}, fmt.Errorf("failed to cast %s.Metadata value (%v) to map[string]dbus.Variant", PlayerInterface, a)
		}
		return Metadata(v), nil
	})
}

// GetVolume returns the volume.
func (i *Player) GetVolume() (float64, error) {
	return getPlayerPropertyCast(i, "Volume", cast.ToFloat64E)
}

// GetLength returns the current track length.
func (i *Player) GetLength() (time.Duration, error) {
	micro, err := getMetadataCast(i, "mpris:length", cast.ToInt64E)
	return time.Duration(micro) * time.Microsecond, err
}

// GetPosition returns the position of the current track.
func (i *Player) GetPosition() (time.Duration, error) {
	micro, err := getPlayerPropertyCast(i, "Position", cast.ToInt64E)
	return time.Duration(micro) * time.Microsecond, err
}

// GetTrackID returns track id for player as dbus.ObjectPath
func (i *Player) GetTrackID() (dbus.ObjectPath, error) {
	trackIdStr, err := getMetadataCast(i, "mpris:trackid", cast.ToStringE)
	return dbus.ObjectPath(trackIdStr), err
}

// SetPosition sets the position of the current track.
func (i *Player) SetPosition(position time.Duration) error {
	trackID, err := i.GetTrackID()
	if err != nil {
		return err
	}
	return i.SetTrackPosition(&trackID, position)
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

// SetVolume sets the volume.
func (i *Player) SetVolume(volume float64) error {
	return i.SetPlayerProperty("Volume", volume)
}

// New connects the the player with the name in the connection conn.
func New(conn *dbus.Conn, name string) *Player {
	obj := conn.Object(name, dbusObjectPath).(*dbus.Object)
	return &Player{conn, obj, name}
}

// OnSignal adds a handler to the player's properties change signal.
func (i *Player) OnSignal(ch chan<- *dbus.Signal) (err error) {
	err = i.conn.AddMatchSignal()
	if err == nil {
		i.conn.Signal(ch)
	}
	return
}
