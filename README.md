<h1 align="center"><pre>Go MPRIS</pre></h1>
<h3 align="center">Go library for DBus-MPRIS.</h3>

<h1 align="center">
<a href="https://pkg.go.dev/github.com/Nadim147c/go-mpris">
<img src="https://img.shields.io/github/go-mod/go-version/Nadim147c/go-mpris?style=for-the-badge&logo=go&labelColor=11140F&color=BBE9AA">
</a>
<a href="https://github.com/Nadim147c/go-mpris">
<img src="https://img.shields.io/github/stars/Nadim147c/go-mpris?style=for-the-badge&logo=github&labelColor=11140F&color=BBE9AA">
</a>
<a href="https://github.com/Nadim147c/go-mpris/blob/main/LICENSE">
<img src="https://img.shields.io/github/license/Nadim147c/go-mpris?style=for-the-badge&logo=gplv3&labelColor=11140F&color=BBE9AA">
</a>
<a href="https://github.com/Nadim147c/go-mpris/commits">
<img src="https://img.shields.io/github/last-commit/Nadim147c/go-mpris?style=for-the-badge&logo=git&labelColor=11140F&color=BBE9AA">
</a>
</h1>

> [!IMPORTANT]
> 🔥 Found this useful? A quick star goes a long way.

> [!DANGER]
> `v2` is very unstable!

## Features

1. Full MPRIS 2.2 support for `Base` and `Player` interfaces (`TrackList` and
   `Playlists` are incomplete).
1. Type-safe D-Bus access ensuring reliable data handling across all players.
1. Native `time.Duration` usage for playback times instead of raw microseconds.
1. Simple, high-level API for playback control, metadata, and property
   management.

## Install

```bash
go get github.com/Nadim147c/go-mpris/v2
```

> The dependency github.com/godbus/dbus/v5 is going to be installed as well.

## Client Example

Printing the current playback status and then changing it:

```go
import (
	"log"

	"github.com/Nadim147c/go-mpris/v2"
	"github.com/godbus/dbus/v5"
)

func main() {
	conn, err := dbus.SessionBus()
	if err != nil {
		panic(err)
	}
	names, err := mpris.List(conn)
	if err != nil {
		panic(err)
	}
	if len(names) == 0 {
		log.Fatal("No player found")
	}

	name := names[0]
	client := mpris.NewClient(conn, name)

	status, err := client.GetPlaybackStatus()
	if err != nil {
		log.Fatal("Could not get current playback status")
	}

	log.Printf("The player was %s...", status)
	err = client.PlayPause()
	if err != nil {
		log.Fatal("Could not play/pause player")
	}
}
```

## Server Example

Creating a mock MPRIS server:

```go
import (
	"context"
	"log"

	"github.com/Nadim147c/go-mpris/v2"
	"github.com/godbus/dbus/v5"
)

func main() {
	conn, err := dbus.SessionBus()
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	name := mpris.BaseInterface + ".my_server"
	server := mpris.NewServer(conn, name)

	meta := mpris.Metadata{}
	meta.Set(mpris.KeyTitle, "My Track")
	meta.Set(mpris.KeyArtist, []string{"My Artist"})

	pm := mpris.NewPropertiesManager(&mpris.Properties{
		PlaybackStatus: mpris.PlaybackPlaying,
		Metadata:       meta,
	})

	server.RegisterPropertiesManager(pm, mpris.NoOpPropertySetter{})
	server.RegisterBaseHandler(mpris.NoOpBaseHandler{})
	server.RegisterPlayerHandler(mpris.NoOpPlayerHandler{})

	log.Printf("Listening on %s", name)
	if err := server.Listen(context.Background()); err != nil {
		log.Fatal(err)
	}
}
```

**For more examples, see the [examples folder](./examples).**

## Go Docs

Read the docs at https://pkg.go.dev/github.com/Nadim147c/go-mpris.

## Credits

[emersion](https://github.com/emersion/go-mpris) and
[Pauloo27](https://github.com/Pauloo27/go-mpris) for the original code.
