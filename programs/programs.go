package programs

import (
	_ "embed"
	"errors"
	"image"

	"github.com/go-gl/mathgl/mgl32"
)

var ErrNoCPUImplementation = errors.New("fractal does not have a CPU implementation")

var (
	NullColour = mgl32.Vec3{0.1, 0.1, 0.1}
)

//go:embed default.vert
var defaultVertexShader string

func NumPrograms() int {
	return len(programs)
}

func GetProgram(i int) Program {
	return programs[i]
}

func SetProgram(i int, p Program) error {
	programs[i] = p
	return nil
}

func NewProgram(p Program) error {
	programs = append(programs, p)
	return nil
}

var programs []Program

type PixelFunc func(uniforms Uniforms, pos mgl32.Vec2) mgl32.Vec3

type Program struct {
	Name           string
	VertexShader   string
	FragmentShader string
	GetPixel       PixelFunc
}

func (p *Program) GetImage(uniforms Uniforms, width, height int) (Image, error) {
	if p.GetPixel == nil {
		return nil, ErrNoCPUImplementation
	}

	width = width / 2
	height = height / 2

	return &programImage{
		uniforms: uniforms,
		bounds: image.Rect(
			-width,
			-height,
			width,
			height,
		),
		pixelFunc: p.GetPixel,
	}, nil
}

type Image interface {
	GetPixel(mgl32.Vec2) mgl32.Vec3
	Bounds() image.Rectangle
}

type programImage struct {
	uniforms  Uniforms
	bounds    image.Rectangle
	pixelFunc PixelFunc
}

func (i *programImage) GetPixel(pos mgl32.Vec2) mgl32.Vec3 {
	return i.pixelFunc(i.uniforms, pos)
}

func (i *programImage) Bounds() image.Rectangle {
	return i.bounds
}
