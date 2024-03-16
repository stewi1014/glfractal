package programs

import (
	_ "embed"
	"math"

	"github.com/go-gl/mathgl/mgl32"
)

//go:embed shaders/julia.frag
var juliaFragment string

func init() {
	NewProgram(Program{
		Name:           "julia",
		VertexShader:   defaultVertexShader,
		FragmentShader: juliaFragment,
		getPixel: func(uniforms Uniforms, x, y float64) mgl32.Vec3 {
			iterations := 0

			z := complex(x*uniforms.Zoom-uniforms.Pos[0], y*uniforms.Zoom-uniforms.Pos[1])

			for math.Abs(real(z))+math.Abs(imag(z)) <= 4 && iterations < int(uniforms.Iterations) {
				z = z*z + complex(uniforms.Sliders[0]-0.835, uniforms.Sliders[1]+0.2321)
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
