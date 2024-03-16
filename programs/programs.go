package programs

import (
	_ "embed"
	"errors"
	"image"
	"image/color"
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

var _ image.Image = &programImage{}

var ErrNoCPUImplementation = errors.New("fractal does not have a CPU implementation")

type PixelFunc func(uniforms Uniforms, x, y float64) color.Color

type Program struct {
	Name           string
	VertexShader   string
	FragmentShader string
	getPixel       PixelFunc
}

type Image interface {
	image.Image
	Progress() float64
}

func (p *Program) GetImage(uniforms Uniforms, width int, height int) (Image, error) {
	if p.getPixel == nil {
		return nil, ErrNoCPUImplementation
	}

	width = width / 2
	height = height / 2

	scaleFactor := width
	if height > width {
		scaleFactor = height
	}

	return &programImage{
		uniforms:    uniforms,
		getPixel:    p.getPixel,
		scaleFactor: float64(scaleFactor),
		bounds: image.Rect(
			-width,
			-height,
			width,
			height,
		),
	}, nil
}

type programImage struct {
	uniforms    Uniforms
	bounds      image.Rectangle
	scaleFactor float64
	getPixel    PixelFunc
	count       int64
}

func (i *programImage) At(x, y int) color.Color {
	i.count++
	return i.getPixel(i.uniforms, float64(x)/i.scaleFactor, float64(y)/i.scaleFactor)
}

func (i *programImage) Bounds() image.Rectangle {
	return i.bounds
}

func (i *programImage) ColorModel() color.Model {
	return color.RGBAModel
}

func (i *programImage) Progress() float64 {
	// for some reason, PNG probes every pixel exactly twice,
	// so we can get an accurate progress by counting up to 2*pixel count.
	end := i.bounds.Dx() * i.bounds.Dy() * 2
	return float64(i.count) / float64(end)
}
