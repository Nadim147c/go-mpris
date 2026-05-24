package mpris

import (
	"context"
	"errors"
	"os/exec"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/godbus/dbus/v5"
)

type playPauseHandler struct {
	t *testing.T
	NoOpPlayerHandler
}

func (p playPauseHandler) PlayPause() error {
	p.t.Log("plase-pause is called")
	return nil
}

func Test_ServerTest(t *testing.T) {
	conn, err := dbus.SessionBus()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })

	const playerName = BaseInterface + ".go-mpris"

	server := NewServer(conn, playerName)
	meta := Metadata{}
	meta.Set(KeyTitle, "title")
	meta.Set(KeyAlbum, "album")
	pm := NewPropertiesManager(&Properties{
		Volume:     0.5,
		Metadata:   meta,
		CanPlay:    true,
		CanControl: true,
	})

	server.RegisterPropertiesManager(pm, NoOpPropertySetter{})
	server.RegisterBaseHandler(NoOpBaseHandler{})
	server.RegisterPlayerHandler(playPauseHandler{t: t})

	ctx, cancel := context.WithCancel(t.Context())
	t.Cleanup(cancel)

	errChan := make(chan error, 1)
	go func() {
		errChan <- server.Listen(ctx)
	}()

	time.Sleep(50 * time.Millisecond)

	clientConn, err := dbus.SessionBus()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { clientConn.Close() })

	list, err := List(clientConn)
	if err != nil {
		t.Fatal(err)
	}

	if !slices.Contains(list, playerName) {
		t.Error("Player bus name does not exist")
	}

	execute(t, "playerctl", "-l")
	execute(t, "playerctl", "metadata", "-p", strings.TrimPrefix(playerName, BaseInterface)[1:])

	player := NewClient(clientConn, playerName)
	fetchedMeta, err := player.GetMetadata()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Fetched Metadata: %v", fetchedMeta)

	player.PlayPause()

	fetchedVolume, err := player.GetVolume()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Fetched volume: %v", fetchedVolume)

	cancel()
	if err, ok := <-errChan; ok && err != nil && !errors.Is(err, context.Canceled) {
		t.Errorf("Server exited with unexpected error: %v", err)
	}
}

func execute(t *testing.T, name string, args ...string) {
	t.Helper()

	output, err := exec.CommandContext(t.Context(), name, args...).CombinedOutput()
	if err != nil {
		t.Error(err)
	}
	t.Log("\n" + string(output))
}
