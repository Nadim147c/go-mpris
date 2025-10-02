# GO-MPRIS

[![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/Nadim147c/go-mpris?style=for-the-badge&logo=go)](https://pkg.go.dev/github.com/Nadim147c/go-mpris)
[![GitHub Repo stars](https://img.shields.io/github/stars/Nadim147c/go-mpris?style=for-the-badge&logo=github)](https://github.com/Nadim147c/go-mpris)
[![GitHub License](https://img.shields.io/github/license/Nadim147c/go-mpris?style=for-the-badge)](./LICENSE)
[![GitHub Tag](https://img.shields.io/github/v/tag/Nadim147c/go-mpris?include_prereleases&sort=semver&style=for-the-badge&logo=github)](https://github.com/Nadim147c/go-mpris/tags)
[![GitHub last commit](https://img.shields.io/github/last-commit/Nadim147c/go-mpris?style=for-the-badge&logo=git)](https://github.com/Nadim147c/go-mpris/commits/master)

A Go library for DBus-MPRIS.

## Install

```bash
go get github.com/Nadim147c/go-mpris
```

> The dependency github.com/godbus/dbus/v5 is going to be installed as well.

## Example

Printing the current playback status and then changing it:

```go
import (
	"log"

	"github.com/Nadim147c/go-mpris"
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
	player := mpris.New(conn, name)

	status, err := player.GetPlaybackStatus()
	if err != nil {
		log.Fatal("Could not get current playback status")
	}

	log.Printf("The player was %s...", status)
	err = player.PlayPause()
	if err != nil {
		log.Fatal("Could not play/pause player")
	}
}
```

**For more examples, see the [examples folder](./examples).**

## Go Docs

Read the docs at https://pkg.go.dev/github.com/Nadim147c/go-mpris.

## Credits

[emersion](https://github.com/emersion/go-mpris) and [Pauloo27](https://github.com/Pauloo27/go-mpris) for the original code.
