package programs

import (
	_ "embed"
)

//go:embed mandelbrot.frag
var mandelbrot string

func init() {
	registerProgram(Program{
		Name:           "Mandelbrot",
		VertexShader:   defaultVertexShader,
		FragmentShader: mandelbrot,
	})
}
