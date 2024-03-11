package main

import (
	"log"
	"os"
	"runtime"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/gotk3/gotk3/gtk"
)

var quitApp func()

func main() {
	runtime.LockOSThread()
	gtk.Init(&os.Args)

	app, err := NewApplication()
	if err != nil {
		panic(err)
	}

	app.Connect("startup", onAppStartup)
	app.Connect("activate", onAppActivate)
	quitApp = app.Quit
	app.Run(nil)
}

func onAppStartup(app *gtk.Application) {
	go glfwMain()
}

func onAppActivate(app *gtk.Application) {
	window, err := gtk.ApplicationWindowNew(app)
	if err != nil {
		panic(err)
	}

	window.SetName("GLFractal")

	l, _ := gtk.LabelNew("Colour Pallet")

	window.Add(l)
	window.ShowAll()
}

func glfwMain() {
	runtime.LockOSThread()
	defer quitApp()

	err := glfw.Init()
	if err != nil {
		log.Println("glfw.Init failed: ", err)
		return
	}
	defer glfw.Terminate()

	monitor := glfw.GetPrimaryMonitor()

	window, err := NewRenderWindow(
		int(float32(monitor.GetVideoMode().Width)*.9),
		int(float32(monitor.GetVideoMode().Height)*.9),
	)

	if err != nil {
		log.Println(err)
		return
	}

	for !window.ShouldClose() {
		glfw.WaitEvents()
	}
}
