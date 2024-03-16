package programs

import (
	_ "embed"
	"image/color"
	"math"
)

//go:embed shaders/mandelbrot.frag
var mandelbrotFragment string

func init() {
	NewProgram(Program{
		Name:           "mandelbrot",
		VertexShader:   defaultVertexShader,
		FragmentShader: mandelbrotFragment,
		getPixel: func(uniforms Uniforms, x, y float64) color.Color {
			iterations := 0

			z_const := complex(x*uniforms.Zoom-uniforms.Pos[0], y*uniforms.Zoom-uniforms.Pos[1])
			z := z_const
			for math.Abs(real(z))+math.Abs(imag(z)) <= 4 && iterations < int(uniforms.Iterations) {
				z = z*z + z_const
				iterations++
			}

			if iterations == int(uniforms.Iterations) {
				return color.Black
			} else {
				colour := uniforms.ColourPallet[iterations%colours]
				return color.RGBA{
					R: uint8(colour[0] * 255),
					G: uint8(colour[1] * 255),
					B: uint8(colour[2] * 255),
					A: 255,
				}
			}
		},
	})
}
