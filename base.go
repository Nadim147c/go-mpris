package mpris

import "github.com/spf13/cast"

// Methods

// Raise raises player priority.
func (i *Player) Raise() error {
	return i.obj.Call(BaseInterface+".Raise", 0).Err
}

// Quit closes the player.
func (i *Player) Quit() error {
	return i.obj.Call(BaseInterface+".Quit", 0).Err
}

// Properties

// CanQuit returns whether the player can be quit.
func (i *Player) CanQuit() (bool, error) { return getBasePropertyCast(i, "CanQuit", cast.ToBoolE) }

// GetFullscreen returns whether the player is in fullscreen mode.
func (i *Player) GetFullscreen() (bool, error) {
	return getBasePropertyCast(i, "Fullscreen", cast.ToBoolE)
}

// SetFullscreen sets the fullscreen state of the player.
func (i *Player) SetFullscreen(fullscreen bool) error {
	return i.SetBaseProperty("Fullscreen", fullscreen)
}

// CanSetFullscreen returns whether the player allows changing fullscreen state.
func (i *Player) CanSetFullscreen() (bool, error) {
	return getBasePropertyCast(i, "CanSetFullscreen", cast.ToBoolE)
}

// CanRaise returns whether the player can be raised.
func (i *Player) CanRaise() (bool, error) {
	return getBasePropertyCast(i, "CanRaise", cast.ToBoolE)
}

// HasTrackList returns whether the player has a track list.
func (i *Player) HasTrackList() (bool, error) {
	return getBasePropertyCast(i, "HasTrackList", cast.ToBoolE)
}

// GetIdentity returns the player identity.
func (i *Player) GetIdentity() (string, error) {
	return getBasePropertyCast(i, "Identity", cast.ToStringE)
}

// GetDesktopEntry returns the desktop entry name of the player.
func (i *Player) GetDesktopEntry() (string, error) {
	return getBasePropertyCast(i, "DesktopEntry", cast.ToStringE)
}

// GetSupportedUriSchemes returns the supported URI schemes of the player.
func (i *Player) GetSupportedUriSchemes() ([]string, error) {
	return getBasePropertyCast(i, "SupportedUriSchemes", cast.ToStringSliceE)
}

// SupportedMimeTypes returns the supported MIME types of the player.
func (i *Player) SupportedMimeTypes() ([]string, error) {
	return getBasePropertyCast(i, "SupportedMimeTypes", cast.ToStringSliceE)
}
