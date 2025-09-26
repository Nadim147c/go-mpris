package mpris

import (
	"slices"
	"testing"

	"github.com/godbus/dbus/v5"
)

// TestPlayerGetMethods runs all get method tests as subtests
func TestPlayerGetMethods(t *testing.T) {
	// Connect to the session bus
	conn, err := dbus.SessionBus()
	if err != nil {
		t.Skipf("Could not connect to session bus: %v", err)
	}

	// Find available players
	players, err := List(conn)
	if err != nil || len(players) == 0 {
		t.Skipf("No MPRIS players found: %v", err)
	}

	// Create a player instance to test
	player := New(conn, players[0])

	// Test GetName
	t.Run("GetName", func(t *testing.T) {
		name := player.GetName()

		if name == "" {
			t.Error("Expected non-empty name, got empty string")
		}

		t.Logf("Player name: %s", name)
	})
	t.Run("GetSupportedUriSchemes", func(t *testing.T) {
		b, err := player.GetSupportedUriSchemes()
		if err != nil {
			t.Errorf("GetSupportedUriSchemes returned error: %v", err)
		}

		t.Logf("SupportedUriSchemes: %v", b)
	})

	t.Run("HasTrackList", func(t *testing.T) {
		b, err := player.HasTrackList()
		if err != nil {
			t.Errorf("HasTrackList returned error: %v", err)
		}

		t.Logf("HasTrackList: %v", b)
	})

	t.Run("CanPlay", func(t *testing.T) {
		b, err := player.CanPlay()
		if err != nil {
			t.Errorf("CanPlay returned error: %v", err)
		}

		t.Logf("Can play: %v", b)
	})

	t.Run("CanControl", func(t *testing.T) {
		b, err := player.CanPlay()
		if err != nil {
			t.Errorf("CanControl returned error: %v", err)
		}

		t.Logf("Can control: %v", b)
	})

	t.Run("CanGoPrevious", func(t *testing.T) {
		b, err := player.CanPlay()
		if err != nil {
			t.Errorf("CanGoPrevious returned error: %v", err)
		}

		t.Logf("Can previous: %v", b)
	})

	t.Run("CanEditTraks", func(t *testing.T) {
		b, err := player.CanEditTracks()
		if err != nil {
			t.Errorf("CanPlay returned error: %v", err)
		}

		t.Logf("Can play: %v", b)
	})

	// Test GetIdentity
	t.Run("GetIdentity", func(t *testing.T) {
		identity, err := player.GetIdentity()
		if err != nil {
			t.Errorf("GetIdentity returned error: %v", err)
		}

		if identity == "" {
			t.Error("Expected non-empty identity, got empty string")
		}

		t.Logf("Player identity: %s", identity)
	})

	// Test GetPlaybackStatus
	t.Run("GetPlaybackStatus", func(t *testing.T) {
		status, err := player.GetPlaybackStatus()
		if err != nil {
			t.Errorf("GetPlaybackStatus returned error: %v", err)
		}

		validStatuses := []PlaybackStatus{PlaybackPlaying, PlaybackPaused, PlaybackStopped}
		valid := slices.Contains(validStatuses, status)

		if !valid {
			t.Errorf("Expected valid PlaybackStatus, got %s", status)
		}

		t.Logf("Player status: %s", status)
	})

	// Test GetLoopStatus
	t.Run("GetLoopStatus", func(t *testing.T) {
		loopStatus, err := player.GetLoopStatus()
		// Some players might not support loop status
		if err != nil {
			t.Logf("GetLoopStatus returned error (might be unsupported): %v", err)
			return
		}

		validStatuses := []LoopStatus{LoopNone, LoopTrack, LoopPlaylist}
		valid := slices.Contains(validStatuses, loopStatus)

		if !valid {
			t.Errorf("Expected valid LoopStatus, got %s", loopStatus)
		}

		t.Logf("Loop status: %s", loopStatus)
	})

	// Test GetRate
	t.Run("GetRate", func(t *testing.T) {
		rate, err := player.GetRate()
		if err != nil {
			t.Errorf("GetRate returned error: %v", err)
		}

		// Rate is typically 1.0 (normal speed) but check it's positive
		if rate <= 0 {
			t.Errorf("Expected positive rate, got %f", rate)
		}

		t.Logf("Player rate: %f", rate)
	})

	// Test GetShuffle
	t.Run("GetShuffle", func(t *testing.T) {
		shuffle, err := player.GetShuffle()
		// Some players might not support shuffle
		if err != nil {
			t.Logf("GetShuffle returned error (might be unsupported): %v", err)
			return
		}

		// Just log the value, no assertion since both true and false are valid
		t.Logf("Shuffle mode: %v", shuffle)
	})

	// Test GetMetadata
	t.Run("GetMetadata", func(t *testing.T) {
		metadata, err := player.GetMetadata()
		// Some players might not have metadata if nothing is playing
		if err != nil {
			t.Logf("GetMetadata returned error (might be no active track): %v", err)
			return
		}

		// Log some common metadata values if they exist
		if trackid, ok := metadata["mpris:trackid"]; ok {
			t.Logf("Track ID: %v", trackid.Value())
		}

		if title, ok := metadata["xesam:title"]; ok {
			t.Logf("Title: %v", title.Value())
		}

		if artist, ok := metadata["xesam:artist"]; ok {
			t.Logf("Artist: %v", artist.Value())
		}
	})

	// Test GetVolume
	t.Run("GetVolume", func(t *testing.T) {
		volume, err := player.GetVolume()
		if err != nil {
			t.Errorf("GetVolume returned error: %v", err)
		}

		// Volume should be between 0.0 and 1.0, or slightly higher for amplification
		if volume < 0.0 || volume > 2.0 {
			t.Errorf("Expected volume between 0.0 and 2.0, got %f", volume)
		}

		t.Logf("Player volume: %f", volume)
	})

	// Test GetLength
	t.Run("GetLength", func(t *testing.T) {
		length, err := player.GetLength()
		// Length might return an error if no track is playing
		if err != nil {
			t.Logf("GetLength returned error (might be no active track): %v", err)
			return
		}

		// Length should be positive
		if length < 0 {
			t.Errorf("Expected positive length, got %s seconds", length.String())
		}

		t.Logf("Track length: %sf ", length.String())
	})

	// Test GetPosition
	t.Run("GetPosition", func(t *testing.T) {
		position, err := player.GetPosition()
		// Position might return an error if no track is playing
		if err != nil {
			t.Logf("GetPosition returned error (might be no active track): %v", err)
			return
		}

		// Check that position is non-negative
		if position < 0 {
			t.Errorf("Expected non-negative position, got %s", position.String())
		}

		// Get the track length and check that position is not beyond it
		length, lengthErr := player.GetLength()
		if lengthErr == nil && length > 0 && position > length+5 {
			// Adding some tolerance (5 seconds) for timing issues
			t.Errorf("Position (%s) exceeds track length (%s)", position.String(), length.String())
		}

		t.Logf("Track position: %s", position.String())
	})

	// Test GetProperty
	t.Run("GetProperty", func(t *testing.T) {
		// Try to get the Identity property from BaseInterface
		variant, err := player.GetProperty(BaseInterface, "Identity")
		if err != nil {
			t.Errorf("GetProperty returned error: %v", err)
		}

		if variant.Value() == nil {
			t.Error("Expected non-nil variant value")
		} else {
			t.Logf("Property value: %v", variant.Value())
		}
	})

	// Test GetPlayerProperty
	t.Run("GetPlayerProperty", func(t *testing.T) {
		// Try to get the PlaybackStatus property
		variant, err := player.GetPlayerProperty("PlaybackStatus")
		if err != nil {
			t.Errorf("GetPlayerProperty returned error: %v", err)
		}

		if variant.Value() == nil {
			t.Error("Expected non-nil variant value")
		} else {
			t.Logf("Player property value: %v", variant.Value())
		}
	})
}
