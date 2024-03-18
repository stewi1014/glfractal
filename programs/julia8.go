package programs

import (
	_ "embed"
	"math"

	"github.com/go-gl/mathgl/mgl32"
)

//go:embed shaders/julia8.frag
var julia8Fragment string

func init() {
	NewProgram(Program{
		Name:           "Julia (8th power)",
		VertexShader:   defaultVertexShader,
		FragmentShader: julia8Fragment,
		GetPixel: func(uniforms Uniforms, pos mgl32.Vec2) mgl32.Vec3 {
			iterations := 0

			z := complex(
				float64(pos[0])*uniforms.Zoom-uniforms.Pos[0],
				float64(pos[1])*uniforms.Zoom-uniforms.Pos[1],
			)

			for math.Abs(real(z))+math.Abs(imag(z)) <= 4 && iterations < int(uniforms.Iterations) {
				z = z*z*z*z*z*z*z*z + complex(uniforms.Sliders[0]-1.08458626270294189453125, uniforms.Sliders[1])
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
