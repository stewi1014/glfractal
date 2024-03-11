package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime"
	"sync"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/gotk3/gotk3/gtk"
)

func main() {
	mainContext, mainQuit := context.WithCancel(context.Background())

	startup := new(sync.WaitGroup)
	startup.Add(2)

	go func() {
		err := gtkMain(mainContext, startup)
		if err != nil {
			log.Println("gtkMain failed: ", err)
		}
		mainQuit()
	}()

	go func() {
		err := glfwMain(mainContext, startup)
		if err != nil {
			log.Println("glfwMain failed:", err)
		}
		mainQuit()
	}()

	<-mainContext.Done()
}

func gtkMain(ctx context.Context, startup *sync.WaitGroup) error {
	runtime.LockOSThread()

	gtk.Init(&os.Args)

	app, err := NewConfigurationWindow()
	if err != nil {
		return fmt.Errorf("NewApplication failed: %w", err)
	}

	go func() {
		<-ctx.Done()
		app.Quit()
	}()
	startup.Done()
	startup.Wait()
	app.Run(nil)
	return nil
}

func glfwMain(ctx context.Context, startup *sync.WaitGroup) error {
	runtime.LockOSThread()

	err := glfw.Init()
	if err != nil {
		return fmt.Errorf("glfw.Init failed: %w", err)
	}
	defer glfw.Terminate()

	monitor := glfw.GetPrimaryMonitor()
	window, err := NewRenderWindow(
		int(float32(monitor.GetVideoMode().Width)*.9),
		int(float32(monitor.GetVideoMode().Height)*.9),
	)

	if err != nil {
		return fmt.Errorf("NewRenderWindow failed: %w", err)
	}

	go func() {
		<-ctx.Done()
		glfw.PostEmptyEvent()
	}()
	startup.Done()
	startup.Wait()
	for !window.ShouldClose() && ctx.Err() == nil {
		glfw.WaitEvents()
	}

	return nil
}
