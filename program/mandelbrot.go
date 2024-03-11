package program

import "image/color"

//go:embed mandelbrot.glsl
var mandelbrot string

var _ Program = Mandelbrot{}

type Mandelbrot struct{}

func (m Mandelbrot) At(x int, y int) color.Color {
	return color.White
}

func (m Mandelbrot) Name() string {
	return "Mandelbrot"
}

func (m Mandelbrot) Shader() string {
	return mandelbrot
}
