package programs

import (
	_ "embed"
	"math"

	"github.com/go-gl/mathgl/mgl32"
)

//go:embed shaders/mandelbrot.frag
var mandelbrotFragment string

func init() {
	NewProgram(Program{
		Name:           "Mandelbrot",
		VertexShader:   defaultVertexShader,
		FragmentShader: mandelbrotFragment,
		getPixel: func(uniforms Uniforms, pos mgl32.Vec2) mgl32.Vec3 {
			iterations := 0

			z := complex(
				float64(pos[0])*uniforms.Zoom-uniforms.Pos[0],
				float64(pos[1])*uniforms.Zoom-uniforms.Pos[1],
			)

			z_const := z
			for math.Abs(real(z))+math.Abs(imag(z)) <= 4 && iterations < int(uniforms.Iterations) {
				z = z*z + z_const
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
