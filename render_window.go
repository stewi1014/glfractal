package main

import (
	"fmt"

	"github.com/go-gl/gl/v4.6-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/gotk3/gotk3/glib"
)

func NewRenderWindow(width, height int) (*RenderWindow, error) {
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 6)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	window, err := glfw.CreateWindow(
		width,
		height,
		"GLFractal Render",
		nil,
		nil,
	)

	if err != nil {
		return nil, fmt.Errorf("glfw.CreateWindow failed: %w", err)
	}

	w := &RenderWindow{
		Window: window,
	}

	w.MakeContextCurrent()
	err = gl.Init()
	if err != nil {
		return nil, fmt.Errorf("gl.Init failed: %w", err)
	}

	glib.IdleAdd(func() {

	})

	return w, nil
}

type RenderWindow struct {
	*glfw.Window
}
