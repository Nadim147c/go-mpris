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
	if err != nil {
		return "", err
	}

	return cast.ToStringE(value.Value())
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
		return "", fmt.Errorf("failed to get property %s.PlaybackStatus: %w", PlayerInterface, err)
	}
	if variant.Value() == nil {
		return "", fmt.Errorf("property %s.PlaybackStatus returned nil value", PlayerInterface)
	}
	str, err := cast.ToStringE(variant.Value())
	if err != nil {
		return "", fmt.Errorf("failed to cast PlaybackStatus value (%v) to string: %w", variant.Value(), err)
	}
	return PlaybackStatus(str), nil
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
		return "", fmt.Errorf("failed to get property %s.LoopStatus: %w", PlayerInterface, err)
	}
	if variant.Value() == nil {
		return "", fmt.Errorf("property %s.LoopStatus returned nil value", PlayerInterface)
	}
	str, err := cast.ToStringE(variant.Value())
	if err != nil {
		return "", fmt.Errorf("failed to cast LoopStatus value (%v) to string: %w", variant.Value(), err)
	}
	return LoopStatus(str), nil
}

// SetLoopStatus sets the loop status to loopStatus.
func (i *Player) SetLoopStatus(loopStatus LoopStatus) error {
	return i.SetPlayerProperty("LoopStatus", loopStatus)
}

// SetProperty sets the value of a propertyName in the targetInterface.
func (i *Player) SetProperty(targetInterface, propertyName string, value any) error {
	return setProperty(i.obj, targetInterface, propertyName, value)
}

// SetPlayerProperty sets the propertyName from the player interface.
func (i *Player) SetPlayerProperty(propertyName string, value any) error {
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
		return 0.0, fmt.Errorf("failed to get property %s.Rate: %w", PlayerInterface, err)
	}
	if variant.Value() == nil {
		return 0.0, fmt.Errorf("property %s.Rate returned nil value", PlayerInterface)
	}
	val, err := cast.ToFloat64E(variant.Value())
	if err != nil {
		return 0.0, fmt.Errorf("failed to cast Rate value (%v) to float64: %w", variant.Value(), err)
	}
	return val, nil
}

// GetShuffle returns false if the player is going linearly through a playlist and true if it's
// in some other order.
func (i *Player) GetShuffle() (bool, error) {
	variant, err := getProperty(i.obj, PlayerInterface, "Shuffle")
	if err != nil {
		return false, fmt.Errorf("failed to get property %s.Shuffle: %w", PlayerInterface, err)
	}
	if variant.Value() == nil {
		return false, fmt.Errorf("property %s.Shuffle returned nil value", PlayerInterface)
	}
	val, err := cast.ToBoolE(variant.Value())
	if err != nil {
		return false, fmt.Errorf("failed to cast Shuffle value (%v) to bool: %w", variant.Value(), err)
	}
	return val, nil
}

// SetShuffle sets the shuffle playlist mode.
func (i *Player) SetShuffle(value bool) error {
	if err := setProperty(i.obj, PlayerInterface, "Shuffle", value); err != nil {
		return fmt.Errorf("failed to set property %s.Shuffle to %v: %w", PlayerInterface, value, err)
	}
	return nil
}

// GetMetadata returns the metadata.
func (i *Player) GetMetadata() (map[string]dbus.Variant, error) {
	variant, err := getProperty(i.obj, PlayerInterface, "Metadata")
	if err != nil {
		return nil, fmt.Errorf("failed to get property %s.Metadata: %w", PlayerInterface, err)
	}
	if variant.Value() == nil {
		return nil, fmt.Errorf("property %s.Metadata returned nil value", PlayerInterface)
	}
	v, ok := variant.Value().(map[string]dbus.Variant)
	if !ok {
		return nil, fmt.Errorf("failed to cast Metadata value (%v) to map[string]dbus.Variant", variant.Value())
	}
	return v, nil
}

// GetVolume returns the volume.
func (i *Player) GetVolume() (float64, error) {
	variant, err := getProperty(i.obj, PlayerInterface, "Volume")
	if err != nil {
		return 0.0, fmt.Errorf("failed to get property %s.Volume: %w", PlayerInterface, err)
	}
	if variant.Value() == nil {
		return 0.0, fmt.Errorf("property %s.Volume returned nil value", PlayerInterface)
	}
	val, err := cast.ToFloat64E(variant.Value())
	if err != nil {
		return 0.0, fmt.Errorf("failed to cast Volume value (%v) to float64: %w", variant.Value(), err)
	}
	return val, nil
}

// GetLength returns the current track length.
func (i *Player) GetLength() (time.Duration, error) {
	metadata, err := i.GetMetadata()
	if err != nil {
		return 0, fmt.Errorf("failed to get metadata for length: %w", err)
	}
	if metadata == nil {
		return 0, fmt.Errorf("metadata is nil")
	}
	v, ok := metadata["mpris:length"]
	if !ok || v.Value() == nil {
		return 0, fmt.Errorf("metadata missing or nil for key 'mpris:length'")
	}
	micro, err := cast.ToInt64E(v.Value())
	if err != nil {
		return 0, fmt.Errorf("failed to cast 'mpris:length' value (%v) to int64: %w", v.Value(), err)
	}
	return time.Duration(micro) * time.Microsecond, nil
}

// GetPosition returns the position of the current track.
func (i *Player) GetPosition() (time.Duration, error) {
	variant, err := getProperty(i.obj, PlayerInterface, "Position")
	if err != nil {
		return 0, fmt.Errorf("failed to get property %s.Position: %w", PlayerInterface, err)
	}
	if variant.Value() == nil {
		return 0, fmt.Errorf("property %s.Position returned nil value", PlayerInterface)
	}
	micro, err := cast.ToInt64E(variant.Value())
	if err != nil {
		return 0, fmt.Errorf("failed to cast Position value (%v) to int64: %w", variant.Value(), err)
	}
	return time.Duration(micro) * time.Microsecond, nil
}

// SetPosition sets the position of the current track.
func (i *Player) SetPosition(position time.Duration) error {
	metadata, err := i.GetMetadata()
	if err != nil {
		return fmt.Errorf("failed to get metadata for SetPosition: %w", err)
	}
	if metadata == nil {
		return fmt.Errorf("metadata is nil")
	}
	v, ok := metadata["mpris:trackid"]
	if !ok || v.Value() == nil {
		return fmt.Errorf("metadata missing or nil for key 'mpris:trackid'")
	}
	trackId, ok := v.Value().(dbus.ObjectPath)
	if !ok {
		return fmt.Errorf("failed to cast 'mpris:trackid' value (%v) to dbus.ObjectPath", v.Value())
	}
	if err := i.SetTrackPosition(&trackId, position); err != nil {
		return fmt.Errorf("failed to set track position for trackId %s at %v: %w", trackId, position, err)
	}
	return nil
}

// SetVolume sets the volume.
func (i *Player) SetVolume(volume float64) error {
	return setProperty(i.obj, PlayerInterface, "Volume", volume)
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
