package main

import (
	"fmt"

	"github.com/Nadim147c/go-mpris"
	"github.com/godbus/dbus/v5"
)

func main() {
	conn, err := dbus.SessionBus()
	if err != nil {
		panic(err)
	}

	signals := make(chan *dbus.Signal)
	if err = mpris.OnSignal(conn, signals); err != nil {
		panic(err)
	}

	for signal := range signals {
		fmt.Println(signal)
	}
}
