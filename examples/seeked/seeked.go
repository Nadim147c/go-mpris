package main

import (
	"context"
	"fmt"
	"log"
	"time"

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

	ch := make(chan time.Duration)

	go func() {
		err = player.OnSeeked(context.Background(), ch)
		if err != nil {
			panic(err)
		}
		close(ch)
	}()

	for dur := range ch {
		fmt.Println(dur)
	}
}
