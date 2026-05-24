package mpris

import (
	"github.com/godbus/dbus/v5"
	"github.com/spf13/cast"
)

// SERVER

// BaseHandler defines the methods for the org.mpris.MediaPlayer2 interface.
type BaseHandler interface {
	Raise() error
	Quit() error
}

// NoOpBaseHandler implement All method of BaseHandler.
type NoOpBaseHandler struct{}

func (NoOpBaseHandler) Raise() error { return nil }
func (NoOpBaseHandler) Quit() error  { return nil }

type dbusBaseHandler struct{ h BaseHandler }

func (d *dbusBaseHandler) Raise() *dbus.Error {
	return toDBusError(d.h.Raise())
}

func (d *dbusBaseHandler) Quit() *dbus.Error {
	return toDBusError(d.h.Quit())
}

// RegisterBaseHandler registers the base handler on the server.
func (s *Server) RegisterBaseHandler(handler BaseHandler) error {
	h := dbusBaseHandler{handler}
	methods := map[string]any{
		"Raise": h.Raise,
		"Quit":  h.Quit,
	}
	return s.conn.ExportSubtreeMethodTable(methods, DBusObjectPath, BaseInterface)
}

// CLIENT

// Methods

// Raise raises player priority.
func (i *Client) Raise() error {
	return i.obj.Call(BaseInterface+".Raise", 0).Err
}

// Quit closes the player.
func (i *Client) Quit() error {
	return i.obj.Call(BaseInterface+".Quit", 0).Err
}

// Properties

// CanQuit returns whether the player can be quit.
func (i *Client) CanQuit() (bool, error) {
	return getBasePropertyCast(i, "CanQuit", cast.ToBoolE)
}

// GetFullscreen returns whether the player is in fullscreen mode.
func (i *Client) GetFullscreen() (bool, error) {
	return getBasePropertyCast(i, "Fullscreen", cast.ToBoolE)
}

// SetFullscreen sets the fullscreen state of the player.
func (i *Client) SetFullscreen(fullscreen bool) error {
	return i.SetBaseProperty("Fullscreen", fullscreen)
}

// CanSetFullscreen returns whether the player allows changing fullscreen state.
func (i *Client) CanSetFullscreen() (bool, error) {
	return getBasePropertyCast(i, "CanSetFullscreen", cast.ToBoolE)
}

// CanRaise returns whether the player can be raised.
func (i *Client) CanRaise() (bool, error) {
	return getBasePropertyCast(i, "CanRaise", cast.ToBoolE)
}

// HasTrackList returns whether the player has a track list.
func (i *Client) HasTrackList() (bool, error) {
	return getBasePropertyCast(i, "HasTrackList", cast.ToBoolE)
}

// GetIdentity returns the player identity.
func (i *Client) GetIdentity() (string, error) {
	return getBasePropertyCast(i, "Identity", cast.ToStringE)
}

// GetDesktopEntry returns the desktop entry name of the player.
func (i *Client) GetDesktopEntry() (string, error) {
	return getBasePropertyCast(i, "DesktopEntry", cast.ToStringE)
}

//revive:disable:var-naming

// GetSupportedUriSchemes returns the supported URI schemes of the player.
func (i *Client) GetSupportedURISchemes() ([]string, error) {
	return getBasePropertyCast(i, "SupportedUriSchemes", cast.ToStringSliceE)
}

//revive:enable:var-naming

// SupportedMimeTypes returns the supported MIME types of the player.
func (i *Client) SupportedMimeTypes() ([]string, error) {
	return getBasePropertyCast(i, "SupportedMimeTypes", cast.ToStringSliceE)
}
