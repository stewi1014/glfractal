package programs

import (
	_ "embed"
	"image/color"
)

//go:embed mandelbrot.frag
var mandelbrot string

func init() {
	RegisterProgram(Mandelbrot{})
}

type Mandelbrot struct{}

func (m Mandelbrot) At(x int, y int) color.Color {
	return color.White
}

func (m Mandelbrot) Name() string {
	return "Mandelbrot"
}

func (m Mandelbrot) FragmentShader() string {
	return mandelbrot
}
