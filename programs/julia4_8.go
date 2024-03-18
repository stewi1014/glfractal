package programs

import (
	_ "embed"
	"math"

	"github.com/go-gl/mathgl/mgl32"
)

//go:embed shaders/julia4_8.frag
var julia4_8Fragment string

func init() {
	NewProgram(Program{
		Name:           "Julia (4th + 8th power)",
		VertexShader:   defaultVertexShader,
		FragmentShader: julia4_8Fragment,
		GetPixel: func(uniforms Uniforms, pos mgl32.Vec2) mgl32.Vec3 {
			iterations := 0

			c := complex(uniforms.Sliders[0]-0.98487460613250732421875, uniforms.Sliders[1])
			z := complex(
				float64(pos[0])*uniforms.Zoom-uniforms.Pos[0],
				float64(pos[1])*uniforms.Zoom-uniforms.Pos[1],
			)

			for math.Abs(real(z))+math.Abs(imag(z)) <= 4 && iterations < int(uniforms.Iterations) {
				z = z*z*z*z + z*z*z*z*z*z*z*z + c
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
