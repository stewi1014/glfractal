package programs

import (
	_ "embed"
	"errors"
	"image"
	"image/color"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
)

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

var _ image.Image = &programImage{}

var ErrNoCPUImplementation = errors.New("fractal does not have a CPU implementation")

type PixelFunc func(uniforms Uniforms, x, y float64) mgl32.Vec3

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

func (p *Program) GetImage(uniforms Uniforms, width int, height int, antialias float64) (Image, error) {
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
		antialias:   antialias,
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
	getPixel    PixelFunc
	scaleFactor float64
	antialias   float64
	bounds      image.Rectangle
	count       int64
}

func (i *programImage) At(x, y int) color.Color {
	i.count++

	// oh how I wish I understood why this was needed
	y = -y

	if i.antialias == 0 {
		c := i.getPixel(i.uniforms, float64(x)/i.scaleFactor, float64(y)/i.scaleFactor)
		return color.RGBA{
			R: uint8(c[0] * 255),
			G: uint8(c[1] * 255),
			B: uint8(c[2] * 255),
			A: 255,
		}
	}

	antialias := i.antialias / i.scaleFactor
	xf, yf := float64(x)/i.scaleFactor, float64(y)/i.scaleFactor

	to_average := []mgl64.Vec2{
		{xf + antialias, yf + antialias},
		{xf + antialias, yf},
		{xf + antialias, yf - antialias},
		{xf, yf + antialias},
		{xf, yf},
		{xf, yf - antialias},
		{xf - antialias, yf + antialias},
		{xf - antialias, yf},
		{xf - antialias, yf - antialias},
	}

	avg := mgl32.Vec3{}
	for _, pos := range to_average {
		avg = avg.Add(i.getPixel(i.uniforms, pos[0], pos[1]))
	}
	avg = avg.Mul(1 / float32(len(to_average)))

	return color.RGBA{
		R: uint8(avg[0] * 255),
		G: uint8(avg[1] * 255),
		B: uint8(avg[2] * 255),
		A: 255,
	}
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
