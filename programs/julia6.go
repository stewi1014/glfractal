package programs

import (
	_ "embed"
	"math"

	"github.com/go-gl/mathgl/mgl32"
)

//go:embed shaders/julia6.frag
var julia6Fragment string

func init() {
	NewProgram(Program{
		Name:           "julia6",
		VertexShader:   defaultVertexShader,
		FragmentShader: julia6Fragment,
		getPixel: func(uniforms Uniforms, x, y float64) mgl32.Vec3 {
			iterations := 0

			z := complex(x*uniforms.Zoom-uniforms.Pos[0], y*uniforms.Zoom-uniforms.Pos[1])

			for math.Abs(real(z))+math.Abs(imag(z)) <= 4 && iterations < int(uniforms.Iterations) {
				z = z*z*z*z*z*z + complex(uniforms.Sliders[0]-0.50517, uniforms.Sliders[1]-035667)
				iterations++
			}

			if iterations == int(uniforms.Iterations) {
				return NullColour
			} else {
				return uniforms.ColourPallet[iterations%colours]
			}
		},
	})
}
