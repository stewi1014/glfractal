package programs

import (
	_ "embed"
	"image/color"
	"math"
)

//go:embed shaders/julia3.frag
var julia3Fragment string

func init() {
	NewProgram(Program{
		Name:           "julia3",
		VertexShader:   defaultVertexShader,
		FragmentShader: julia3Fragment,
		getPixel: func(uniforms Uniforms, x, y float64) color.Color {
			iterations := 0

			z := complex(x*uniforms.Zoom-uniforms.Pos[0], y*uniforms.Zoom-uniforms.Pos[1])

			for math.Abs(real(z))+math.Abs(imag(z)) <= 4 && iterations < int(uniforms.Iterations) {
				z = z*z*z + complex(uniforms.Sliders[0]+0.08394, uniforms.Sliders[1]+0.77007)
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
