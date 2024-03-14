package main

import (
	"context"
	_ "embed"
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/stewi1014/glfractal/programs"
)

const debug = true

//go:embed icon.ico
var icon []byte

func init() {
	gob.Register(&programs.Uniforms{})
	gob.Register(&programs.Program{})
}

func main() {
	mainContext, mainQuit := context.WithCancelCause(context.Background())

	go func() {
		mainQuit(gtkMain(mainContext))
	}()

	<-mainContext.Done()
	if err := context.Cause(mainContext); err != nil {
		log.Println(err)
	}
}

func gtkMain(ctx context.Context) error {
	runtime.LockOSThread()

	gtk.Init(&os.Args)
	app, err := gtk.ApplicationNew("com.github.stewi1014.glfractal", glib.APPLICATION_FLAGS_NONE)
	if err != nil {
		return fmt.Errorf("gtk.ApplicationNew failed: %w", err)
	}

	iconPixbuf, _ := gdk.PixbufNewFromBytesOnly(icon)

	appContext, appQuit := context.WithCancelCause(ctx)
	app.Connect("activate", func() {
		client, listener := NewPipeListener(appContext)

		renderWindow := NewRenderWindow(app, client, ctx, appQuit)
		renderWindow.Connect("destroy", func() {
			appQuit(nil)
		})
		renderWindow.SetTitle("GLFractal Render")
		renderWindow.SetIcon(iconPixbuf)

		configWindow := NewConfigWindow(app, listener, ctx, appQuit)
		configWindow.Connect("destroy", func() {
			appQuit(nil)
		})
		configWindow.SetTitle("GLFractal Config")
		configWindow.SetIcon(iconPixbuf)
	})

	go func() {
		<-appContext.Done()
		app.Quit()
	}()
	app.Run(nil)
	return context.Cause(appContext)
}
