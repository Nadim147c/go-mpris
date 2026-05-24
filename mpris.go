package mpris

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/spf13/cast"
)

const (
	// DBusObjectPath is the root object path for MPRIS-compatible media
	// players. All MPRIS interfaces are exposed under this path on the D-Bus.
	DBusObjectPath = "/org/mpris/MediaPlayer2"
	// PropertiesChangedSignal is the D-Bus signal name emitted when a property
	// changes on an MPRIS interface.
	PropertiesChangedSignal = "org.freedesktop.DBus.Properties.PropertiesChanged"
	// BaseInterface is the main MPRIS interface that provides general
	// information and capabilities about the media player instance.
	BaseInterface = "org.mpris.MediaPlayer2"
	// PlayerInterface defines methods and properties for controlling playback,
	// such as play, pause, seek, and retrieving track metadata.
	PlayerInterface = "org.mpris.MediaPlayer2.Player"
	// TrackListInterface provides access to the list of tracks managed by the
	// player, allowing navigation, retrieval, and management of track items.
	TrackListInterface = "org.mpris.MediaPlayer2.TrackList"
	// PlaylistsInterface defines the MPRIS interface for managing and
	// activating playlists exposed by the player.
	PlaylistsInterface = "org.mpris.MediaPlayer2.Playlists"
	// PlaylistsInterface defines the MPRIS properties.
	DBusPropertyInterface = "org.freedesktop.DBus.Properties"
	// GetPropertyMethod is the standard D-Bus method used to retrieve the value
	// of a property from an interface that implements
	// org.freedesktop.DBus.Properties.
	GetPropertyMethod = "org.freedesktop.DBus.Properties.Get"
	// SetPropertyMethod is the standard D-Bus method used to change the value
	// of a writable property on an interface that implements
	// org.freedesktop.DBus.Properties.
	SetPropertyMethod = "org.freedesktop.DBus.Properties.Set"
)

const (
	// KeyAlbum is the album name.
	KeyAlbum = "xesam:album"
	// KeyAlbumArtist is the album artist(s).
	KeyAlbumArtist = "xesam:albumArtist"
	// KeyArtist is the track artist(s).
	KeyArtist = "xesam:artist"
	// KeyAsText is the text/lyrics of the track.
	KeyAsText = "xesam:asText"
	// KeyAudioBPM is the speed of the music in beats per minute.
	KeyAudioBPM = "xesam:audioBPM"
	// KeyAutoRating is an automatically generated rating (0.0 to 1.0).
	KeyAutoRating = "xesam:autoRating"
	// KeyComment is a social comment about the track.
	KeyComment = "xesam:comment"
	// KeyComposer is the track composer(s).
	KeyComposer = "xesam:composer"
	// KeyContentCreated is the date/time the content was created.
	KeyContentCreated = "xesam:contentCreated"
	// KeyDiscNumber is the disc number on the album.
	KeyDiscNumber = "xesam:discNumber"
	// KeyFirstUsed is the date/time the track was first played.
	KeyFirstUsed = "xesam:firstUsed"
	// KeyGenre is the genre(s) of the track.
	KeyGenre = "xesam:genre"
	// KeyLastUsed is the date/time the track was last played by any user.
	KeyLastUsed = "xesam:lastUsed"
	// KeyLastUsedByMe is the date/time the track was last played by the current
	// user.
	KeyLastUsedByMe = "xesam:lastUsedByMe"
	// KeyLyricist is the person who wrote the lyrics for the track.
	KeyLyricist = "xesam:lyricist"
	// KeyTitle is the item title.
	KeyTitle = "xesam:title"
	// KeyTrackNumber is the track number on the album.
	KeyTrackNumber = "xesam:trackNumber"
	// KeyURL is the location of the media file.
	KeyURL = "xesam:url"
	// KeyUseCount is the number of times the track has been played.
	KeyUseCount = "xesam:useCount"
	// KeyUserRating is the user's rating of the track (0.0 to 1.0).
	KeyUserRating = "xesam:userRating"
	// KeyTrackID is a unique identity for the track within the context of the playlist.
	KeyTrackID = "mpris:trackid"
	// KeyLength is the duration of the track in microseconds.
	KeyLength = "mpris:length"
	// KeyArtURL is a URI of some album art designed to represent the track/album.
	KeyArtURL = "mpris:artUrl"
)

var (
	ErrInvalidType     = errors.New("invalid type")
	ErrNotPrimaryOwner = errors.New("not primary owner")
	ErrValueMissing    = errors.New("value missing or nil")
)

// List lists the available players.
func List(conn *dbus.Conn) ([]string, error) {
	var names []string
	err := conn.BusObject().
		Call("org.freedesktop.DBus.ListNames", 0).
		Store(&names)
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

// toDBusError converts a Go error to a DBus error.
func toDBusError(err error) *dbus.Error {
	if err == nil {
		return nil
	}
	return dbus.MakeFailedError(err)
}

// SERVER

// Server represents a mpris server.
type Server struct {
	conn   *dbus.Conn
	name   string
	cancel func()
}

// NewServer connects the the server with the name in the connection conn.
func NewServer(conn *dbus.Conn, name string) *Server {
	return &Server{conn: conn, name: name}
}

func (s *Server) Listen(ctx context.Context) error {
	ctx, s.cancel = context.WithCancel(ctx)
	defer s.cancel()

	reply, err := s.conn.RequestName(s.name, dbus.NameFlagReplaceExisting)
	if err != nil || reply != dbus.RequestNameReplyPrimaryOwner {
		return fmt.Errorf("unable to claim %s: %w", s.name, ErrNotPrimaryOwner)
	}

	<-ctx.Done()

	if _, err := s.conn.ReleaseName(s.name); err != nil {
		return err
	}

	return ctx.Err()
}

func (s *Server) Close() error {
	_, err := s.conn.ReleaseName(s.name)
	if err != nil {
		return err
	}
	s.cancel()
	return nil
}

// CLIENT

// Client represents a mpris player.
type Client struct {
	conn *dbus.Conn
	obj  *dbus.Object
	name string
}

// GetName gets the player full name.
func (i *Client) GetName() string {
	return i.name
}

// CanEditTracks returns if player can edit track list.
func (i *Client) CanEditTracks() (bool, error) {
	return getTrackListPropertyCast(i, "CanEditTracks", cast.ToBoolE)
}

// GetLength returns the current track length.
func (i *Client) GetLength() (time.Duration, error) {
	micro, err := getMetadataCast(i, KeyLength, cast.ToInt64E)
	return time.Duration(micro) * time.Microsecond, err
}

// GetTrackID returns track id for player as dbus.ObjectPath.
func (i *Client) GetTrackID() (dbus.ObjectPath, error) {
	trackIDStr, err := getMetadataCast(i, KeyTrackID, cast.ToStringE)
	return dbus.ObjectPath(trackIDStr), err
}

// GetTitle returns the current track title.
func (i *Client) GetTitle() (string, error) {
	return getMetadataCast(i, KeyTitle, cast.ToStringE)
}

// GetArtist returns the current track artist(s).
func (i *Client) GetArtist() ([]string, error) {
	return getMetadataCast(i, KeyArtist, cast.ToStringSliceE)
}

// GetAlbum returns the current track album.
func (i *Client) GetAlbum() (string, error) {
	return getMetadataCast(i, KeyAlbum, cast.ToStringE)
}

// GetURL returns the URL of the current track.
func (i *Client) GetURL() (string, error) {
	return getMetadataCast(i, KeyURL, cast.ToStringE)
}

// GetArtURL returns the cover art URL of the current track.
func (i *Client) GetArtURL() (string, error) {
	return getMetadataCast(i, KeyArtURL, cast.ToStringE)
}

// NewClient connects the the player with the name in the connection conn.
func NewClient(conn *dbus.Conn, name string) *Client {
	obj := conn.Object(name, DBusObjectPath).(*dbus.Object)
	return &Client{conn, obj, name}
}

// OnSignal adds a handler to the player's properties change signal.
func OnSignal(conn *dbus.Conn, ch chan<- *dbus.Signal) error {
	// receive all MPRIS signal
	err := conn.AddMatchSignal(dbus.WithMatchObjectPath(DBusObjectPath))
	if err == nil {
		conn.Signal(ch)
	}
	return err
}
