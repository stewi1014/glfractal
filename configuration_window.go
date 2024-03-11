package main

import (
	"fmt"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/stewi1014/glfractal/program"
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

func (w *ConfigurationWindow) onActivate(app *gtk.Application) {
	window, err := gtk.ApplicationWindowNew(app)
	if err != nil {
		panic(err)
	}

	window.SetName("GLFractal")
	window.SetDefaultSize(280, 700)

	grid, _ := gtk.GridNew()

	modeMenu, _ := gtk.ComboBoxTextNew()
	modeMenu.Connect("changed", w.onProgramChange)
	for _, program := range program.Programs() {
		modeMenu.AppendText(program.Name())
	}
	modeMenu.SetActive(0)
	grid.Add(modeMenu)

	window.Add(grid)
	window.ShowAll()
}

func (w *ConfigurationWindow) onProgramChange(comboBox *gtk.ComboBoxText) {
	fmt.Println("program changed to", comboBox.GetActive())
}
