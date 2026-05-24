package mpris

import (
	"context"
	"fmt"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/spf13/cast"
)

// SERVER

// PlayerHandler defines the methods for the org.mpris.MediaPlayer2.Player interface.
type PlayerHandler interface {
	Next() error
	Previous() error
	Pause() error
	PlayPause() error
	Stop() error
	Play() error
	Seek(offset time.Duration) error
	SetPosition(trackID dbus.ObjectPath, position time.Duration) error
	OpenURI(uri string) error
}

// NoOpPlayerHandler implement All method of PlayerHandler.
type NoOpPlayerHandler struct{}

func (NoOpPlayerHandler) Next() error              { return nil }
func (NoOpPlayerHandler) Previous() error          { return nil }
func (NoOpPlayerHandler) Pause() error             { return nil }
func (NoOpPlayerHandler) PlayPause() error         { return nil }
func (NoOpPlayerHandler) Stop() error              { return nil }
func (NoOpPlayerHandler) Play() error              { return nil }
func (NoOpPlayerHandler) Seek(time.Duration) error { return nil }
func (NoOpPlayerHandler) OpenURI(string) error     { return nil }
func (NoOpPlayerHandler) SetPosition(dbus.ObjectPath, time.Duration) error {
	return nil
}

type dbusPlayerHandler struct{ h PlayerHandler }

func (d *dbusPlayerHandler) Next() *dbus.Error {
	return toDBusError(d.h.Next())
}

func (d *dbusPlayerHandler) Previous() *dbus.Error {
	return toDBusError(d.h.Previous())
}

func (d *dbusPlayerHandler) Pause() *dbus.Error {
	return toDBusError(d.h.Pause())
}

func (d *dbusPlayerHandler) PlayPause() *dbus.Error {
	return toDBusError(d.h.PlayPause())
}

func (d *dbusPlayerHandler) Stop() *dbus.Error {
	return toDBusError(d.h.Stop())
}

func (d *dbusPlayerHandler) Play() *dbus.Error {
	return toDBusError(d.h.Play())
}

// NOTE: Seek is method for io.Seeker.

func (d *dbusPlayerHandler) MprisSeek(offset int64) *dbus.Error {
	return toDBusError(d.h.Seek(time.Duration(offset) * time.Microsecond))
}

func (d *dbusPlayerHandler) SetPosition(trackID dbus.ObjectPath, position int64) *dbus.Error {
	return toDBusError(d.h.SetPosition(trackID, time.Duration(position)*time.Microsecond))
}

func (d *dbusPlayerHandler) OpenUri(uri string) *dbus.Error {
	return toDBusError(d.h.OpenURI(uri))
}

// RegisterPlayerHandler registers the player handler on the server.
func (s *Server) RegisterPlayerHandler(handler PlayerHandler) error {
	h := dbusPlayerHandler{handler}
	methods := map[string]any{
		"Next":        h.Next,
		"Previous":    h.Previous,
		"Pause":       h.Pause,
		"PlayPause":   h.PlayPause,
		"Stop":        h.Stop,
		"Play":        h.Play,
		"Seek":        h.MprisSeek,
		"SetPosition": h.SetPosition,
		"OpenUri":     h.OpenUri,
	}
	return s.conn.ExportSubtreeMethodTable(methods, DBusObjectPath, PlayerInterface)
}

//---------
// CLIENT
// ---------

// Methods

// Next skips to the next track in the tracklist.
func (i *Client) Next() error {
	return i.obj.Call(PlayerInterface+".Next", 0).Err
}

// Previous skips to the previous track in the tracklist.
func (i *Client) Previous() error {
	return i.obj.Call(PlayerInterface+".Previous", 0).Err
}

// Pause pauses the current track.
func (i *Client) Pause() error {
	return i.obj.Call(PlayerInterface+".Pause", 0).Err
}

// PlayPause resumes the current track if it's paused and pauses it if it's
// playing.
func (i *Client) PlayPause() error {
	return i.obj.Call(PlayerInterface+".PlayPause", 0).Err
}

// Stop stops the current track.
func (i *Client) Stop() error {
	return i.obj.Call(PlayerInterface+".Stop", 0).Err
}

// Play starts or resumes playback of the current track.
func (i *Client) Play() error {
	return i.obj.Call(PlayerInterface+".Play", 0).Err
}

// Seek changes the current track position by the given offset.
// If the offset is negative, the playback position moves backward.
func (i *Client) Seek(offset time.Duration) error {
	micro := offset.Microseconds()
	return i.obj.Call(PlayerInterface+".Seek", 0, micro).Err
}

// SetTrackPosition sets the playback position of a specific track.
func (i *Client) SetTrackPosition(
	trackID *dbus.ObjectPath,
	position time.Duration,
) error {
	oms := position.Microseconds()
	return i.obj.Call(PlayerInterface+".SetPosition", 0, trackID, oms).Err
}

// SetPosition sets the playback position of the current track.
func (i *Client) SetPosition(position time.Duration) error {
	trackID, err := i.GetTrackID()
	if err != nil {
		return err
	}
	return i.SetTrackPosition(&trackID, position)
}

// OpenURI opens and plays the given URI if supported.
func (i *Client) OpenURI(uri string) error {
	return i.obj.Call(PlayerInterface+".OpenUri", 0, uri).Err
}

// Signals

// OnSeeked listens for "Seeked" signal and sends the new position as
// time.Duration to position until ctx is canceled.
func (i *Client) OnSeeked(ctx context.Context, position chan<- time.Duration) error {
	sigChan := make(chan *dbus.Signal, 10) // buffered to avoid blocking
	defer close(sigChan)

	var sender string
	err := i.conn.BusObject().
		Call("org.freedesktop.DBus.GetNameOwner", 0, i.name).
		Store(&sender)
	if err != nil {
		return err
	}

	err = i.conn.AddMatchSignal(
		dbus.WithMatchInterface(PlayerInterface),
		dbus.WithMatchMember("Seeked"),
		dbus.WithMatchSender(sender),
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

			var dur time.Duration
			err := dbus.Store(signal.Body, &dur)
			if err != nil {
				continue
			}

			position <- dur * time.Microsecond
		}
	}
}

// Properties

// PlaybackStatus represents the playback status. It can be "Playing", "Paused"
// or "Stopped".
type PlaybackStatus string

//revive:disable:exported

const (
	PlaybackPlaying PlaybackStatus = "Playing"
	PlaybackPaused  PlaybackStatus = "Paused"
	PlaybackStopped PlaybackStatus = "Stopped"
)

//revive:enable:exported

// GetPlaybackStatus returns the current playback status.
func (i *Client) GetPlaybackStatus() (PlaybackStatus, error) {
	str, err := getPlayerPropertyCast(i, "PlaybackStatus", cast.ToStringE)
	return PlaybackStatus(str), err
}

// LoopStatus represents the loop status of the player. It can be "None",
// "Track" or "Playlist".
type LoopStatus string

//revive:disable:exported
const (
	LoopNone     LoopStatus = "None"
	LoopTrack    LoopStatus = "Track"
	LoopPlaylist LoopStatus = "Playlist"
)

//revive:enable:exported

// GetLoopStatus returns the current loop status.
func (i *Client) GetLoopStatus() (LoopStatus, error) {
	str, err := getPlayerPropertyCast(i, "LoopStatus", cast.ToStringE)
	return LoopStatus(str), err
}

// SetLoopStatus sets the loop status.
func (i *Client) SetLoopStatus(loopStatus LoopStatus) error {
	return i.SetPlayerProperty("LoopStatus", loopStatus)
}

// GetRate returns the current playback rate.
func (i *Client) GetRate() (float64, error) {
	return getPlayerPropertyCast(i, "Rate", cast.ToFloat64E)
}

// SetRate sets the playback rate.
func (i *Client) SetRate(rate float64) error {
	return i.SetPlayerProperty("Rate", rate)
}

// GetShuffle returns true if shuffle mode is enabled, false if playing linearly
// through a playlist.
func (i *Client) GetShuffle() (bool, error) {
	return getPlayerPropertyCast(i, "Shuffle", cast.ToBoolE)
}

// SetShuffle sets the shuffle mode.
func (i *Client) SetShuffle(value bool) error {
	return i.SetPlayerProperty("Shuffle", value)
}

// Metadata represents the metadata of the current track.
type Metadata map[string]dbus.Variant

// Has checks if the metadata key exists.
func (m Metadata) Has(key string) bool {
	v, ok := m[key]
	return ok && v.Value() != nil
}

// Set sets a value to metadata.
func (m Metadata) Set(key string, v any) {
	m[key] = dbus.MakeVariant(v)
}

// Get returns the value for the given metadata key.
func (m Metadata) Get(key string) (any, error) {
	v, ok := m[key]
	if !ok || v.Value() == nil {
		return v, fmt.Errorf(
			"%s.Metadata missing or nil for key %q: %w",
			PlayerInterface, key, ErrValueMissing,
		)
	}
	return v.Value(), nil
}

// GetMetadata returns the current track metadata.
func (i *Client) GetMetadata() (Metadata, error) {
	return getPlayerPropertyCast(i, "Metadata", func(a any) (Metadata, error) {
		v, ok := a.(map[string]dbus.Variant)
		if !ok {
			return Metadata{}, fmt.Errorf(
				"failed to cast %s.Metadata value to map[string]dbus.Variant: %w",
				PlayerInterface, ErrInvalidType,
			)
		}
		return Metadata(v), nil
	})
}

// GetVolume returns the current volume.
func (i *Client) GetVolume() (float64, error) {
	return getPlayerPropertyCast(i, "Volume", cast.ToFloat64E)
}

// SetVolume sets the current volume.
func (i *Client) SetVolume(volume float64) error {
	return i.SetPlayerProperty("Volume", volume)
}

// GetPosition returns the current playback position.
func (i *Client) GetPosition() (time.Duration, error) {
	micro, err := getPlayerPropertyCast(i, "Position", cast.ToInt64E)
	return time.Duration(micro) * time.Microsecond, err
}

// GetMinimumRate returns the minimum playback rate.
func (i *Client) GetMinimumRate() (float64, error) {
	return getPlayerPropertyCast(i, "MinimumRate", cast.ToFloat64E)
}

// GetMaximumRate returns the maximum playback rate.
func (i *Client) GetMaximumRate() (float64, error) {
	return getPlayerPropertyCast(i, "MaximumRate", cast.ToFloat64E)
}

// CanGoNext returns whether the player can skip to the next track.
func (i *Client) CanGoNext() (bool, error) {
	return getPlayerPropertyCast(i, "CanGoNext", cast.ToBoolE)
}

// CanGoPrevious returns whether the player can skip to the previous track.
func (i *Client) CanGoPrevious() (bool, error) {
	return getPlayerPropertyCast(i, "CanGoPrevious", cast.ToBoolE)
}

// CanPlay returns whether the player can start or resume playback.
func (i *Client) CanPlay() (bool, error) {
	return getPlayerPropertyCast(i, "CanPlay", cast.ToBoolE)
}

// CanPause returns whether the player can pause playback.
func (i *Client) CanPause() (bool, error) {
	return getPlayerPropertyCast(i, "CanPause", cast.ToBoolE)
}

// CanSeek returns whether the player can seek within the current track.
func (i *Client) CanSeek() (bool, error) {
	return getPlayerPropertyCast(i, "CanSeek", cast.ToBoolE)
}

// CanControl returns whether the player can be controlled.
func (i *Client) CanControl() (bool, error) {
	return getPlayerPropertyCast(i, "CanControl", cast.ToBoolE)
}
