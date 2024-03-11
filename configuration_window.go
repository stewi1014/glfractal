package main

import (
	"fmt"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

func NewConfigurationWindow() (*ConfigurationWindow, error) {
	app, err := gtk.ApplicationNew("com.github.stewi1014.glfractal", glib.APPLICATION_FLAGS_NONE)
	if err != nil {
		return nil, fmt.Errorf("gtk.ApplicationNew failed: %w", err)
	}

	a := &ConfigurationWindow{
		Application: app,
	}

	app.Connect("activate", a.onActivate)

	return a, nil
}

type ConfigurationWindow struct {
	*gtk.Application
}

func (a *ConfigurationWindow) onActivate(app *gtk.Application) {
	window, err := gtk.ApplicationWindowNew(app)
	if err != nil {
		panic(err)
	}

	window.SetName("GLFractal")

	l, _ := gtk.LabelNew("Colour Pallet")

	window.Add(l)
	window.ShowAll()
}
