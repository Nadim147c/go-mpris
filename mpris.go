package mpris

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/godbus/dbus/v5"
)

const (
	dbusObjectPath          = "/org/mpris/MediaPlayer2"
	propertiesChangedSignal = "org.freedesktop.DBus.Properties.PropertiesChanged"

	BaseInterface      = "org.mpris.MediaPlayer2"
	PlayerInterface    = "org.mpris.MediaPlayer2.Player"
	TrackListInterface = "org.mpris.MediaPlayer2.TrackList"
	PlaylistsInterface = "org.mpris.MediaPlayer2.Playlists"

	getPropertyMethod = "org.freedesktop.DBus.Properties.Get"
	setPropertyMethod = "org.freedesktop.DBus.Properties.Set"
)

func getProperty(obj *dbus.Object, iface string, prop string) (dbus.Variant, error) {
	result := dbus.Variant{}
	err := obj.Call(getPropertyMethod, 0, iface, prop).Store(&result)
	if err != nil {
		return dbus.Variant{}, err
	}
	return result, nil
}

func setProperty(obj *dbus.Object, iface string, prop string, val interface{}) error {
	call := obj.Call(setPropertyMethod, 0, iface, prop, dbus.MakeVariant(val))
	return call.Err
}

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
	value, err := getProperty(i.obj, BaseInterface, "Identity")

	return value.Value().(string), err
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
func (i *Player) OpenUri(uri string) error {
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
	variant, err := i.obj.GetProperty(PlayerInterface + ".PlaybackStatus")
	if err != nil {
		return "", err
	}
	if variant.Value() == nil {
		return "", fmt.Errorf("Variant value is nil")
	}
	return PlaybackStatus(variant.Value().(string)), nil
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
	variant, err := getProperty(i.obj, PlayerInterface, "LoopStatus")
	if err != nil {
		return LoopStatus(""), err
	}
	if variant.Value() == nil {
		return "", fmt.Errorf("Variant value is nil")
	}
	return LoopStatus(variant.Value().(string)), nil
}

// SetLoopStatus sets the loop status to loopStatus.
func (i *Player) SetLoopStatus(loopStatus LoopStatus) error {
	return i.SetPlayerProperty("LoopStatus", loopStatus)
}

// SetProperty sets the value of a propertyName in the targetInterface.
func (i *Player) SetProperty(targetInterface, propertyName string, value interface{}) error {
	return setProperty(i.obj, targetInterface, propertyName, value)
}

// SetPlayerProperty sets the propertyName from the player interface.
func (i *Player) SetPlayerProperty(propertyName string, value interface{}) error {
	return setProperty(i.obj, PlayerInterface, propertyName, value)
}

// GetProperty returns the properityName in the targetInterface.
func (i *Player) GetProperty(targetInterface, properityName string) (dbus.Variant, error) {
	return getProperty(i.obj, targetInterface, properityName)
}

// GetPlayerProperty returns the properityName from the player interface.
func (i *Player) GetPlayerProperty(properityName string) (dbus.Variant, error) {
	return getProperty(i.obj, PlayerInterface, properityName)
}

// Returns the current playback rate.
func (i *Player) GetRate() (float64, error) {
	variant, err := getProperty(i.obj, PlayerInterface, "Rate")
	if err != nil {
		return 0.0, err
	}
	if variant.Value() == nil {
		return 0.0, fmt.Errorf("Variant value is nil")
	}
	return variant.Value().(float64), nil
}

// GetShuffle returns false if the player is going linearly through a playlist and false if it's
// in some other order.
func (i *Player) GetShuffle() (bool, error) {
	variant, err := getProperty(i.obj, PlayerInterface, "Shuffle")
	if err != nil {
		return false, err
	}
	if variant.Value() == nil {
		return false, fmt.Errorf("Variant value is nil")
	}
	return variant.Value().(bool), nil
}

// SetShuffle sets the shuffle playlist mode.
func (i *Player) SetShuffle(value bool) error {
	return setProperty(i.obj, PlayerInterface, "Shuffle", value)
}

// GetMetadata returns the metadata.
func (i *Player) GetMetadata() (map[string]dbus.Variant, error) {
	variant, err := getProperty(i.obj, PlayerInterface, "Metadata")
	if err != nil {
		return nil, err
	}
	if variant.Value() == nil {
		return nil, fmt.Errorf("Variant value is nil")
	}
	return variant.Value().(map[string]dbus.Variant), nil
}

// GetVolume returns the volume.
func (i *Player) GetVolume() (float64, error) {
	variant, err := getProperty(i.obj, PlayerInterface, "Volume")
	if err != nil {
		return 0.0, err
	}
	if variant.Value() == nil {
		return 0.0, fmt.Errorf("Variant value is nil")
	}
	return variant.Value().(float64), nil
}

// SetVolume sets the volume.
func (i *Player) SetVolume(volume float64) error {
	return setProperty(i.obj, PlayerInterface, "Volume", volume)
}

// GetLength returns the current track length.
func (i *Player) GetLength() (time.Duration, error) {
	metadata, err := i.GetMetadata()
	if err != nil {
		return 0, err
	}
	if metadata == nil || metadata["mpris:length"].Value() == nil {
		return 0, fmt.Errorf("Variant value is nil")
	}

	length := metadata["mpris:length"].Value()

	v := reflect.ValueOf(length)

	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return time.Duration(v.Int()) * time.Microsecond, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return time.Duration(int64(v.Uint())) * time.Microsecond, nil
	default:
		return 0, nil
	}
}

// GetPosition returns the position of the current track.
func (i *Player) GetPosition() (time.Duration, error) {
	variant, err := getProperty(i.obj, PlayerInterface, "Position")
	if err != nil {
		return 0, err
	}
	if variant.Value() == nil {
		return 0, fmt.Errorf("Variant value is nil")
	}
	duration := time.Duration(variant.Value().(int64)) * time.Microsecond
	return duration, nil
}

// SetPosition sets the position of the current track.
func (i *Player) SetPosition(position time.Duration) error {
	metadata, err := i.GetMetadata()
	if err != nil {
		return err
	}
	if metadata == nil || metadata["mpris:trackid"].Value() == nil {
		return fmt.Errorf("Variant value is nil")
	}
	trackId := metadata["mpris:trackid"].Value().(dbus.ObjectPath)
	i.SetTrackPosition(&trackId, position)
	return nil
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
