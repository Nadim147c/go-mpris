package mpris

import (
	"fmt"

	"github.com/godbus/dbus/v5"
)

// SetProperty sets the value of a property in the interface.
func (i *Player) SetProperty(iface, property string, value any) error {
	call := i.obj.Call(setPropertyMethod, 0, iface, property, dbus.MakeVariant(value))
	if call.Err != nil {
		return fmt.Errorf("failed to set property %s.%s to value (%v): %w", iface, property, value, call.Err)
	}
	return nil
}

// SetBaseProperty sets the propertyName from the base interface.
func (i *Player) SetBaseProperty(property string, value any) error {
	return i.SetProperty(BaseInterface, property, value)
}

// SetPlayerProperty sets the propertyName from the player interface.
func (i *Player) SetPlayerProperty(property string, value any) error {
	return i.SetProperty(PlayerInterface, property, value)
}

// SetTrackListProperty sets the propertyName from the tracklist interface.
func (i *Player) SetTrackListProperty(property string, value any) error {
	return i.SetProperty(TrackListInterface, property, value)
}

// SetPlaylistsProperty sets the propertyName from the playlists interface.
func (i *Player) SetPlaylistsProperty(property string, value any) error {
	return i.SetProperty(PlaylistsInterface, property, value)
}

// GetProperty returns the prop in the iface.
func (i *Player) GetProperty(iface, property string) (dbus.Variant, error) {
	result := dbus.Variant{}
	call := i.obj.Call(getPropertyMethod, 0, iface, property)
	if call.Err != nil {
		return dbus.Variant{}, fmt.Errorf("failed to get property %s.%s: %w", iface, property, call.Err)
	}
	if err := call.Store(&result); err != nil {
		return dbus.Variant{}, fmt.Errorf("failed to store property %s.%s result into variant: %w", iface, property, err)
	}
	return result, nil
}

// GetBaseProperty returns the prop from the base interface.
func (i *Player) GetBaseProperty(property string) (dbus.Variant, error) {
	return i.GetProperty(BaseInterface, property)
}

// GetPlayerProperty returns the prop from the player interface.
func (i *Player) GetPlayerProperty(property string) (dbus.Variant, error) {
	return i.GetProperty(PlayerInterface, property)
}

// GetTrackListProperty returns the prop from the tracklist interface.
func (i *Player) GetTrackListProperty(property string) (dbus.Variant, error) {
	return i.GetProperty(TrackListInterface, property)
}

// GetPlaylistsProperty returns the prop from the playlists interface.
func (i *Player) GetPlaylistsProperty(property string) (dbus.Variant, error) {
	return i.GetProperty(PlaylistsInterface, property)
}

// getPropertyCast returns property and casts value using the provided caster function.
func getPropertyCast[T any](i *Player, iface, property string, caster func(any) (T, error)) (T, error) {
	var v T
	variant, err := i.GetProperty(iface, property)
	if err != nil {
		return v, err
	}
	if variant.Value() == nil {
		return v, fmt.Errorf("property %s.%s returned nil value", iface, property)
	}
	result, err := caster(variant.Value())
	if err != nil {
		return v, fmt.Errorf("failed to cast %s.%s value (%v): %w", iface, property, variant.Value(), err)
	}
	return result, nil
}

// getBasePropertyCast returns base interface property and casts value using the provided caster function.
func getBasePropertyCast[T any](i *Player, property string, caster func(any) (T, error)) (T, error) {
	return getPropertyCast(i, BaseInterface, property, caster)
}

// getPlayerPropertyCast returns player interface property and casts value using the provided caster function.
func getPlayerPropertyCast[T any](i *Player, property string, caster func(any) (T, error)) (T, error) {
	return getPropertyCast(i, PlayerInterface, property, caster)
}

// getTrackListPropertyCast returns tracklist interface property and casts value using the provided caster function.
func getTrackListPropertyCast[T any](i *Player, property string, caster func(any) (T, error)) (T, error) {
	return getPropertyCast(i, TrackListInterface, property, caster)
}

// getPlaylistPropertyCast returns playlists interface property and casts value using the provided caster function.
func getPlaylistPropertyCast[T any](i *Player, property string, caster func(any) (T, error)) (T, error) {
	return getPropertyCast(i, PlaylistsInterface, property, caster)
}

// getMetadataCast returns metadata value for the given key and casts it using the provided caster function.
func getMetadataCast[T any](i *Player, key string, caster func(any) (T, error)) (T, error) {
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
		return v, fmt.Errorf("%s.Metadata: failed to cast value (%v) of %q: %w", PlayerInterface, val, key, err)
	}
	return v, nil
}
