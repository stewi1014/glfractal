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
		Name:           "mandelbrot",
		VertexShader:   defaultVertexShader,
		FragmentShader: mandelbrotFragment,
		getPixel: func(uniforms Uniforms, x, y float64) mgl32.Vec3 {
			iterations := 0

			z_const := complex(x*uniforms.Zoom-uniforms.Pos[0], y*uniforms.Zoom-uniforms.Pos[1])
			z := z_const
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
