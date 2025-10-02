package mpris

import (
	"context"
	"fmt"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/spf13/cast"
)

// Methods

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

// Play starts or resumes playback of the current track.
func (i *Player) Play() error {
	return i.obj.Call(PlayerInterface+".Play", 0).Err
}

// Seek changes the current track position by the given offset.
// If the offset is negative, the playback position moves backward.
func (i *Player) Seek(offset time.Duration) error {
	micro := offset.Microseconds()
	return i.obj.Call(PlayerInterface+".Seek", 0, micro).Err
}

// SetTrackPosition sets the playback position of a specific track.
func (i *Player) SetTrackPosition(trackId *dbus.ObjectPath, position time.Duration) error {
	oms := position.Microseconds()
	return i.obj.Call(PlayerInterface+".SetPosition", 0, trackId, oms).Err
}

// SetPosition sets the playback position of the current track.
func (i *Player) SetPosition(position time.Duration) error {
	trackID, err := i.GetTrackID()
	if err != nil {
		return err
	}
	return i.SetTrackPosition(&trackID, position)
}

// OpenUri opens and plays the given URI if supported.
//
// Deprecated: Use OpenURI instead.
func (i *Player) OpenUri(uri string) error {
	return i.OpenURI(uri)
}

// OpenURI opens and plays the given URI if supported.
func (i *Player) OpenURI(uri string) error {
	return i.obj.Call(PlayerInterface+".OpenUri", 0, uri).Err
}

// Signals

// OnSeeked listens for  "Seeked" signal and sends the new position as time.Duration to position
// until ctx is canceled.
func (i *Player) OnSeeked(ctx context.Context, position chan<- time.Duration) error {
	sigChan := make(chan *dbus.Signal, 10) // buffered to avoid blocking
	defer close(sigChan)

	err := i.conn.AddMatchSignal(
		dbus.WithMatchInterface(PlayerInterface),
		dbus.WithMatchMember("Seeked"),
	)
	if err != nil {
		return err
	}

	i.conn.Signal(sigChan)

	for {
		select {
		case <-ctx.Done():
			return nil
		case signal, ok := <-sigChan:
			if !ok {
				return nil
			}
			if len(signal.Body) != 1 {
				continue
			}

			if val, err := cast.ToInt64E(signal.Body[0]); err == nil {
				position <- time.Duration(val) * time.Microsecond
			}
		}
	}
}

// Properties

// PlaybackStatus represents the playback status. It can be "Playing", "Paused" or "Stopped".
type PlaybackStatus string

const (
	PlaybackPlaying PlaybackStatus = "Playing"
	PlaybackPaused  PlaybackStatus = "Paused"
	PlaybackStopped PlaybackStatus = "Stopped"
)

// GetPlaybackStatus returns the current playback status.
func (i *Player) GetPlaybackStatus() (PlaybackStatus, error) {
	str, err := getPlayerPropertyCast(i, "PlaybackStatus", cast.ToStringE)
	return PlaybackStatus(str), err
}

// LoopStatus represents the loop status of the player. It can be "None", "Track" or "Playlist".
type LoopStatus string

const (
	LoopNone     LoopStatus = "None"
	LoopTrack    LoopStatus = "Track"
	LoopPlaylist LoopStatus = "Playlist"
)

// GetLoopStatus returns the current loop status.
func (i *Player) GetLoopStatus() (LoopStatus, error) {
	str, err := getPlayerPropertyCast(i, "LoopStatus", cast.ToStringE)
	return LoopStatus(str), err
}

// SetLoopStatus sets the loop status.
func (i *Player) SetLoopStatus(loopStatus LoopStatus) error {
	return i.SetPlayerProperty("LoopStatus", loopStatus)
}

// GetRate returns the current playback rate.
func (i *Player) GetRate() (float64, error) {
	return getPlayerPropertyCast(i, "Rate", cast.ToFloat64E)
}

// SetRate sets the playback rate.
func (i *Player) SetRate(rate float64) error {
	return i.SetPlayerProperty("Rate", rate)
}

// GetShuffle returns true if shuffle mode is enabled, false if playing linearly through a playlist.
func (i *Player) GetShuffle() (bool, error) {
	return getPlayerPropertyCast(i, "Shuffle", cast.ToBoolE)
}

// SetShuffle sets the shuffle mode.
func (i *Player) SetShuffle(value bool) error {
	return i.SetPlayerProperty("Shuffle", value)
}

// Metadata represents the metadata of the current track.
type Metadata map[string]dbus.Variant

// Get returns the value for the given metadata key.
func (m Metadata) Get(key string) (any, error) {
	v, ok := m[key]
	if !ok || v.Value() == nil {
		return v, fmt.Errorf("%s.Metadata missing or nil for key %q", PlayerInterface, key)
	}
	return v.Value(), nil
}

// GetMetadata returns the current track metadata.
func (i *Player) GetMetadata() (Metadata, error) {
	return getPlayerPropertyCast(i, "Metadata", func(a any) (Metadata, error) {
		v, ok := a.(map[string]dbus.Variant)
		if !ok {
			return Metadata{}, fmt.Errorf(
				"failed to cast %s.Metadata value (%v) to map[string]dbus.Variant",
				PlayerInterface, a)
		}
		return Metadata(v), nil
	})
}

// GetVolume returns the current volume.
func (i *Player) GetVolume() (float64, error) {
	return getPlayerPropertyCast(i, "Volume", cast.ToFloat64E)
}

// SetVolume sets the current volume.
func (i *Player) SetVolume(volume float64) error {
	return i.SetPlayerProperty("Volume", volume)
}

// GetPosition returns the current playback position.
func (i *Player) GetPosition() (time.Duration, error) {
	micro, err := getPlayerPropertyCast(i, "Position", cast.ToInt64E)
	return time.Duration(micro) * time.Microsecond, err
}

// GetMinimumRate returns the minimum playback rate.
func (i *Player) GetMinimumRate() (float64, error) {
	return getPlayerPropertyCast(i, "MinimumRate", cast.ToFloat64E)
}

// GetMaximumRate returns the maximum playback rate.
func (i *Player) GetMaximumRate() (float64, error) {
	return getPlayerPropertyCast(i, "MaximumRate", cast.ToFloat64E)
}

// CanGoNext returns whether the player can skip to the next track.
func (i *Player) CanGoNext() (bool, error) {
	return getPlayerPropertyCast(i, "CanGoNext", cast.ToBoolE)
}

// CanGoPrevious returns whether the player can skip to the previous track.
func (i *Player) CanGoPrevious() (bool, error) {
	return getPlayerPropertyCast(i, "CanGoPrevious", cast.ToBoolE)
}

// CanPlay returns whether the player can start or resume playback.
func (i *Player) CanPlay() (bool, error) {
	return getPlayerPropertyCast(i, "CanPlay", cast.ToBoolE)
}

// CanPause returns whether the player can pause playback.
func (i *Player) CanPause() (bool, error) {
	return getPlayerPropertyCast(i, "CanPause", cast.ToBoolE)
}

// CanSeek returns whether the player can seek within the current track.
func (i *Player) CanSeek() (bool, error) {
	return getPlayerPropertyCast(i, "CanSeek", cast.ToBoolE)
}

// CanControl returns whether the player can be controlled.
func (i *Player) CanControl() (bool, error) {
	return getPlayerPropertyCast(i, "CanControl", cast.ToBoolE)
}
