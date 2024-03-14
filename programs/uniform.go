package programs

import (
	"math/rand"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
)

const sliders = 5
const colours = 170

type ColourPallet [colours]mgl32.Vec3

func RandomColourPallet() ColourPallet {
	var pallet ColourPallet
	pallet[0] = mgl32.Vec3{
		rand.Float32(),
		rand.Float32(),
		rand.Float32(),
	}
	for i := 1; i < colours; i += 1 {
		pallet[i] = mgl32.Vec3{
			limit(pallet[i-1].X() + (rand.Float32()-.5)*.3),
			limit(pallet[i-1].Y() + (rand.Float32()-.5)*.3),
			limit(pallet[i-1].Z() + (rand.Float32()-.5)*.3),
		}
	}
	return pallet
}

func limit(n float32) float32 {
	if n < 0 {
		return 0
	}

	if n > 1 {
		return 1
	}

	return n
}

type Uniforms struct {
	Zoom         float64          `uniform:"zoom"`
	Pos          mgl64.Vec2       `uniform:"pos"`
	Iterations   uint32           `uniform:"max_iterations"`
	Sliders      [sliders]float64 `uniform:"sliders"`
	Camera       mgl32.Mat4       `uniform:"camera"`
	ColourPallet ColourPallet     `uniform:"colour_pallet"`
}

func (u *Uniforms) DefaultValues() {
	u.Zoom = 2
	u.Pos = mgl64.Vec2{0, 0}
	u.Iterations = 500
	u.Sliders = [sliders]float64{}
	u.Camera = mgl32.Ident4()
	u.ColourPallet = RandomColourPallet()
}
