package programs

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
)

type Uniforms struct {
	Zoom   float64
	Pos    mgl64.Vec2
	Camera mgl32.Mat4
}
