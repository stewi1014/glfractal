package main

import (
	"fmt"
	"net"

	"github.com/gotk3/gotk3/gtk"
)

func NewConfigWindow(
	app *gtk.Application,
	listener net.Listener,
	quit func(error),
) *ConfigWindow {
	var err error
	w := &ConfigWindow{
		quit: quit,
	}

	w.ApplicationWindow, err = gtk.ApplicationWindowNew(app)
	if err != nil {
		quit(fmt.Errorf("gtk.ApplicationWindowNew: %w", err))
	}

	w.SetDefaultSize(280, 700)

	l, _ := gtk.LabelNew("test")

	w.Add(l)
	w.ShowAll()

	return w
}

type ConfigWindow struct {
	*gtk.ApplicationWindow
	quit func(error)
}
