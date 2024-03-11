package main

import (
	"fmt"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

func NewApplication() (*Application, error) {
	app, err := gtk.ApplicationNew("com.github.stewi1014.glfractal", glib.APPLICATION_FLAGS_NONE)
	if err != nil {
		return nil, fmt.Errorf("gtk.ApplicationNew failed: %w", err)
	}

	a := &Application{
		Application: app,
	}

	return a, nil
}

type Application struct {
	*gtk.Application
}
