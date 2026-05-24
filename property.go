package mpris

import (
	"fmt"
	"sync"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/spf13/cast"
)

// SERVER

// Properties contains all the properties supported by the MPRIS server.
type Properties struct {
	// Base properties
	CanQuit             bool
	Fullscreen          bool
	CanSetFullscreen    bool
	CanRaise            bool
	HasTrackList        bool
	Identity            string
	DesktopEntry        string
	SupportedUriSchemes []string
	SupportedMimeTypes  []string

	// Player properties
	PlaybackStatus PlaybackStatus
	LoopStatus     LoopStatus
	Rate           float64
	Shuffle        bool
	Metadata       Metadata
	Volume         float64
	Position       time.Duration
	MinimumRate    float64
	MaximumRate    float64
	CanGoNext      bool
	CanGoPrevious  bool
	CanPlay        bool
	CanPause       bool
	CanSeek        bool
	CanControl     bool
}

// PropertiesManager manages access to MPRIS properties.
type PropertiesManager struct {
	mu sync.RWMutex
	p  *Properties
}

// NewPropertiesManager creates a new PropertiesManager.
func NewPropertiesManager(p *Properties) *PropertiesManager {
	if p == nil {
		p = &Properties{}
	}
	return &PropertiesManager{p: p}
}

// SetProperty allows modifying properties safely.
func (pm *PropertiesManager) SetProperty(f func(p *Properties)) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	f(pm.p)
}

// PropertySetter handles write requests to MPRIS properties.
type PropertySetter interface {
	SetFullscreen(fullscreen bool) error
	SetLoopStatus(status LoopStatus) error
	SetRate(rate float64) error
	SetShuffle(shuffle bool) error
	SetVolume(volume float64) error
}

// NoOpPropertySetter implement All method of PropertySetter.
type NoOpPropertySetter struct{}

func (NoOpPropertySetter) SetFullscreen(fullscreen bool) error   { return nil }
func (NoOpPropertySetter) SetLoopStatus(status LoopStatus) error { return nil }
func (NoOpPropertySetter) SetRate(rate float64) error            { return nil }
func (NoOpPropertySetter) SetShuffle(shuffle bool) error         { return nil }
func (NoOpPropertySetter) SetVolume(volume float64) error        { return nil }

// dbusPropertyHandler handles org.freedesktop.DBus.Properties interface.
type dbusPropertyHandler struct {
	pm     *PropertiesManager
	setter PropertySetter
}

// Get returns the value of the specified property.
func (d *dbusPropertyHandler) Get(iface string, property string) (dbus.Variant, *dbus.Error) {
	d.pm.mu.RLock()
	defer d.pm.mu.RUnlock()
	p := d.pm.p

	switch iface {
	case BaseInterface:
		switch property {
		case "CanQuit":
			return dbus.MakeVariant(p.CanQuit), nil
		case "Fullscreen":
			return dbus.MakeVariant(p.Fullscreen), nil
		case "CanSetFullscreen":
			return dbus.MakeVariant(p.CanSetFullscreen), nil
		case "CanRaise":
			return dbus.MakeVariant(p.CanRaise), nil
		case "HasTrackList":
			return dbus.MakeVariant(p.HasTrackList), nil
		case "Identity":
			return dbus.MakeVariant(p.Identity), nil
		case "DesktopEntry":
			return dbus.MakeVariant(p.DesktopEntry), nil
		case "SupportedUriSchemes":
			return dbus.MakeVariant(p.SupportedUriSchemes), nil
		case "SupportedMimeTypes":
			return dbus.MakeVariant(p.SupportedMimeTypes), nil
		}
	case PlayerInterface:
		switch property {
		case "PlaybackStatus":
			return dbus.MakeVariant(string(p.PlaybackStatus)), nil
		case "LoopStatus":
			return dbus.MakeVariant(string(p.LoopStatus)), nil
		case "Rate":
			return dbus.MakeVariant(p.Rate), nil
		case "Shuffle":
			return dbus.MakeVariant(p.Shuffle), nil
		case "Metadata":
			if p.Metadata == nil {
				return dbus.MakeVariant(map[string]dbus.Variant{}), nil
			}
			return dbus.MakeVariant(p.Metadata), nil
		case "Volume":
			return dbus.MakeVariant(p.Volume), nil
		case "Position":
			return dbus.MakeVariant(p.Position.Microseconds()), nil
		case "MinimumRate":
			return dbus.MakeVariant(p.MinimumRate), nil
		case "MaximumRate":
			return dbus.MakeVariant(p.MaximumRate), nil
		case "CanGoNext":
			return dbus.MakeVariant(p.CanGoNext), nil
		case "CanGoPrevious":
			return dbus.MakeVariant(p.CanGoPrevious), nil
		case "CanPlay":
			return dbus.MakeVariant(p.CanPlay), nil
		case "CanPause":
			return dbus.MakeVariant(p.CanPause), nil
		case "CanSeek":
			return dbus.MakeVariant(p.CanSeek), nil
		case "CanControl":
			return dbus.MakeVariant(p.CanControl), nil
		}
	}
	return dbus.Variant{}, dbus.NewError("org.freedesktop.DBus.Error.UnknownProperty", nil)
}

// Set sets the value of the specified property.
func (d *dbusPropertyHandler) Set(iface string, property string, value dbus.Variant) *dbus.Error {
	if d.setter == nil {
		return dbus.NewError("org.freedesktop.DBus.Error.NotSupported", []any{"setting properties is not supported"})
	}

	switch iface {
	case BaseInterface:
		switch property {
		case "Fullscreen":
			v, err := cast.ToBoolE(value.Value())
			if err == nil {
				return toDBusError(d.setter.SetFullscreen(v))
			}
			return dbus.NewError("org.freedesktop.DBus.Error.InvalidArgs", []any{err.Error()})
		}
	case PlayerInterface:
		switch property {
		case "LoopStatus":
			v, err := cast.ToStringE(value.Value())
			if err == nil {
				return toDBusError(d.setter.SetLoopStatus(LoopStatus(v)))
			}
			return dbus.NewError("org.freedesktop.DBus.Error.InvalidArgs", []any{err.Error()})
		case "Rate":
			v, err := cast.ToFloat64E(value.Value())
			if err == nil {
				return toDBusError(d.setter.SetRate(v))
			}
			return dbus.NewError("org.freedesktop.DBus.Error.InvalidArgs", []any{err.Error()})
		case "Shuffle":
			v, err := cast.ToBoolE(value.Value())
			if err == nil {
				return toDBusError(d.setter.SetShuffle(v))
			}
			return dbus.NewError("org.freedesktop.DBus.Error.InvalidArgs", []any{err.Error()})
		case "Volume":
			v, err := cast.ToFloat64E(value.Value())
			if err == nil {
				return toDBusError(d.setter.SetVolume(v))
			}
			return dbus.NewError("org.freedesktop.DBus.Error.InvalidArgs", []any{err.Error()})
		}
	}
	return dbus.NewError("org.freedesktop.DBus.Error.UnknownProperty", nil)
}

// GetAll returns all properties of the specified interface.
func (d *dbusPropertyHandler) GetAll(iface string) (map[string]dbus.Variant, *dbus.Error) {
	d.pm.mu.RLock()
	defer d.pm.mu.RUnlock()
	p := d.pm.p

	props := make(map[string]dbus.Variant)
	switch iface {
	case BaseInterface:
		props["CanQuit"] = dbus.MakeVariant(p.CanQuit)
		props["Fullscreen"] = dbus.MakeVariant(p.Fullscreen)
		props["CanSetFullscreen"] = dbus.MakeVariant(p.CanSetFullscreen)
		props["CanRaise"] = dbus.MakeVariant(p.CanRaise)
		props["HasTrackList"] = dbus.MakeVariant(p.HasTrackList)
		props["Identity"] = dbus.MakeVariant(p.Identity)
		props["DesktopEntry"] = dbus.MakeVariant(p.DesktopEntry)
		props["SupportedUriSchemes"] = dbus.MakeVariant(p.SupportedUriSchemes)
		props["SupportedMimeTypes"] = dbus.MakeVariant(p.SupportedMimeTypes)
		return props, nil
	case PlayerInterface:
		props["PlaybackStatus"] = dbus.MakeVariant(string(p.PlaybackStatus))
		props["LoopStatus"] = dbus.MakeVariant(string(p.LoopStatus))
		props["Rate"] = dbus.MakeVariant(p.Rate)
		props["Shuffle"] = dbus.MakeVariant(p.Shuffle)
		if p.Metadata == nil {
			props["Metadata"] = dbus.MakeVariant(map[string]dbus.Variant{})
		} else {
			props["Metadata"] = dbus.MakeVariant(p.Metadata)
		}
		props["Volume"] = dbus.MakeVariant(p.Volume)
		props["Position"] = dbus.MakeVariant(p.Position.Microseconds())
		props["MinimumRate"] = dbus.MakeVariant(p.MinimumRate)
		props["MaximumRate"] = dbus.MakeVariant(p.MaximumRate)
		props["CanGoNext"] = dbus.MakeVariant(p.CanGoNext)
		props["CanGoPrevious"] = dbus.MakeVariant(p.CanGoPrevious)
		props["CanPlay"] = dbus.MakeVariant(p.CanPlay)
		props["CanPause"] = dbus.MakeVariant(p.CanPause)
		props["CanSeek"] = dbus.MakeVariant(p.CanSeek)
		props["CanControl"] = dbus.MakeVariant(p.CanControl)
		return props, nil
	}
	return nil, dbus.NewError("org.freedesktop.DBus.Error.UnknownInterface", nil)
}

// RegisterPropertiesManager registers the property manager on the server.
func (s *Server) RegisterPropertiesManager(pm *PropertiesManager, setter PropertySetter) error {
	p := &dbusPropertyHandler{pm, setter}
	methods := map[string]any{
		"Get":    p.Get,
		"GetAll": p.GetAll,
		"Set":    p.Set,
	}
	return s.conn.ExportSubtreeMethodTable(methods, DBusObjectPath, DBusPropertyInterface)
}

// CLIENT

// SetProperty sets the value of a property in the interface.
func (i *Client) SetProperty(iface, property string, value any) error {
	call := i.obj.Call(
		SetPropertyMethod,
		0,
		iface,
		property,
		dbus.MakeVariant(value),
	)
	if call.Err != nil {
		return fmt.Errorf(
			"failed to set property %s.%s to value (%v): %w",
			iface,
			property,
			value,
			call.Err,
		)
	}
	return nil
}

// SetBaseProperty sets the propertyName from the base interface.
func (i *Client) SetBaseProperty(property string, value any) error {
	return i.SetProperty(BaseInterface, property, value)
}

// SetPlayerProperty sets the propertyName from the player interface.
func (i *Client) SetPlayerProperty(property string, value any) error {
	return i.SetProperty(PlayerInterface, property, value)
}

// SetTrackListProperty sets the propertyName from the tracklist interface.
func (i *Client) SetTrackListProperty(property string, value any) error {
	return i.SetProperty(TrackListInterface, property, value)
}

// SetPlaylistsProperty sets the propertyName from the playlists interface.
func (i *Client) SetPlaylistsProperty(property string, value any) error {
	return i.SetProperty(PlaylistsInterface, property, value)
}

// GetProperty returns the prop in the iface.
func (i *Client) GetProperty(iface, property string) (dbus.Variant, error) {
	result := dbus.Variant{}
	call := i.obj.Call(GetPropertyMethod, 0, iface, property)
	if call.Err != nil {
		return dbus.Variant{}, fmt.Errorf(
			"failed to get property %s.%s: %w",
			iface,
			property,
			call.Err,
		)
	}
	if err := call.Store(&result); err != nil {
		return dbus.Variant{}, fmt.Errorf(
			"failed to store property %s.%s result into variant: %w",
			iface,
			property,
			err,
		)
	}
	return result, nil
}

// GetBaseProperty returns the prop from the base interface.
func (i *Client) GetBaseProperty(property string) (dbus.Variant, error) {
	return i.GetProperty(BaseInterface, property)
}

// GetPlayerProperty returns the prop from the player interface.
func (i *Client) GetPlayerProperty(property string) (dbus.Variant, error) {
	return i.GetProperty(PlayerInterface, property)
}

// GetTrackListProperty returns the prop from the tracklist interface.
func (i *Client) GetTrackListProperty(property string) (dbus.Variant, error) {
	return i.GetProperty(TrackListInterface, property)
}

// GetPlaylistsProperty returns the prop from the playlists interface.
func (i *Client) GetPlaylistsProperty(property string) (dbus.Variant, error) {
	return i.GetProperty(PlaylistsInterface, property)
}

// getPropertyCast returns property and casts value using the provided caster
// function.
func getPropertyCast[T any](
	i *Client,
	iface, property string,
	caster func(any) (T, error),
) (T, error) {
	var v T
	variant, err := i.GetProperty(iface, property)
	if err != nil {
		return v, err
	}
	if variant.Value() == nil {
		return v, fmt.Errorf(
			"failed get %s.%s: %w",
			iface, property, ErrValueMissing,
		)
	}
	result, err := caster(variant.Value())
	if err != nil {
		return v, fmt.Errorf(
			"failed to cast %s.%s value (%v): %w",
			iface, property, variant.Value(), err,
		)
	}
	return result, nil
}

// getBasePropertyCast returns base interface property and casts value using the
// provided caster function.
func getBasePropertyCast[T any](
	i *Client,
	property string,
	caster func(any) (T, error),
) (T, error) {
	return getPropertyCast(i, BaseInterface, property, caster)
}

// getPlayerPropertyCast returns player interface property and casts value using
// the provided caster function.
func getPlayerPropertyCast[T any](
	i *Client,
	property string,
	caster func(any) (T, error),
) (T, error) {
	return getPropertyCast(i, PlayerInterface, property, caster)
}

// getTrackListPropertyCast returns tracklist interface property and casts value
// using the provided caster function.
func getTrackListPropertyCast[T any](
	i *Client,
	property string,
	caster func(any) (T, error),
) (T, error) {
	return getPropertyCast(i, TrackListInterface, property, caster)
}

// getPlaylistPropertyCast returns playlists interface property and casts value
// using the provided caster function.
// func getPlaylistPropertyCast[T any](
// 	i *Client,
// 	property string,
// 	caster func(any) (T, error),
// ) (T, error) {
// 	return getPropertyCast(i, PlaylistsInterface, property, caster)
// }

// getMetadataCast returns metadata value for the given key and casts it using
// the provided caster function.
func getMetadataCast[T any](
	i *Client,
	key string,
	caster func(any) (T, error),
) (T, error) {
	var v T
	m, err := i.GetMetadata()
	if err != nil {
		return v, err
	}
	val, err := m.Get(key)
	if err != nil {
		return v, err
	}
	v, err = caster(val)
	if err != nil {
		return v, fmt.Errorf(
			"%s.Metadata: failed to cast value (%v) of %q: %w",
			PlayerInterface, val, key, err,
		)
	}
	return v, nil
}
