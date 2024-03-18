package main

// #cgo pkg-config: gdk-3.0 glib-2.0 gobject-2.0
// #include <gdk/gdk.h>
import "C"

import (
	"context"
	"image"
	"image/color"
	"log"
	"sync"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/stewi1014/glfractal/programs"
)

func WrapWithProgress(img *image.Image) func() float64 {
	p := &ProgressImage{
		Image: *img,
	}

	*img = p
	return p.Progress
}

type ProgressImage struct {
	image.Image
	count int
}

func (i *ProgressImage) At(x, y int) color.Color {
	i.count++
	return i.Image.At(x, y)
}

func (i *ProgressImage) Progress() float64 {
	end := i.Bounds().Dx() * i.Bounds().Dy()
	return float64(i.count) / float64(end)
}

func (i *ProgressImage) Opaque() bool {
	return true
}

// AntiAlias9x samples 9 posititions for each sampled position,
// returning the average colour.
//
// antialias is the number of pixels apart the sampled locations are.
func AntiAlias9x(img programs.Image, antialias float32) programs.Image {
	if antialias == 0 {
		log.Println("image uselessly antialiased with distance of 0")
	}

	scaleFactor := float32(img.Bounds().Dx())
	if img.Bounds().Dy() > img.Bounds().Dx() {
		scaleFactor = float32(img.Bounds().Dy())
	}

	return &antialias9xImage{
		Image:  img,
		offset: antialias / scaleFactor,
	}
}

type antialias9xImage struct {
	programs.Image
	offset float32
}

func (i *antialias9xImage) GetPixel(pos mgl32.Vec2) mgl32.Vec3 {
	avg := mgl32.Vec3{}
	avg = avg.Add(i.Image.GetPixel(mgl32.Vec2{pos[0] + i.offset, pos[1] + i.offset}))
	avg = avg.Add(i.Image.GetPixel(mgl32.Vec2{pos[0] + i.offset, pos[1]}))
	avg = avg.Add(i.Image.GetPixel(mgl32.Vec2{pos[0] + i.offset, pos[1] - i.offset}))
	avg = avg.Add(i.Image.GetPixel(mgl32.Vec2{pos[0], pos[1] + i.offset}))
	avg = avg.Add(i.Image.GetPixel(mgl32.Vec2{pos[0], pos[1]}))
	avg = avg.Add(i.Image.GetPixel(mgl32.Vec2{pos[0], pos[1] - i.offset}))
	avg = avg.Add(i.Image.GetPixel(mgl32.Vec2{pos[0] - i.offset, pos[1] + i.offset}))
	avg = avg.Add(i.Image.GetPixel(mgl32.Vec2{pos[0] - i.offset, pos[1]}))
	avg = avg.Add(i.Image.GetPixel(mgl32.Vec2{pos[0] - i.offset, pos[1] - i.offset}))
	return avg.Mul(1 / float32(9))
}

func BufferImage(img image.Image) *BufferedImage {
	return &BufferedImage{
		Image:  img,
		height: img.Bounds().Dy(),
	}
}

type BufferedImage struct {
	image.Image
	height int
	buff   []color.Color
}

func (b *BufferedImage) Bounds() image.Rectangle {
	return image.Rect(0, 0, b.Image.Bounds().Dx(), b.Image.Bounds().Dy())
}

func (b *BufferedImage) At(x, y int) color.Color {
	return b.buff[x*b.height+y]
}

func (b *BufferedImage) Buffer(ctx context.Context) error {
	b.buff = make([]color.Color, b.Image.Bounds().Dx()*b.Image.Bounds().Dy())

	min, max := b.Image.Bounds().Min, b.Image.Bounds().Max
	chunkSize := 50
	var wg sync.WaitGroup

	for chunkMin := min.X; chunkMin < max.X; chunkMin += chunkSize {
		chunkMax := chunkMin + chunkSize
		if chunkMax > max.X {
			chunkMax = max.X
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			i := (chunkMin - min.X) * b.Image.Bounds().Dy()
			for x := chunkMin; x < chunkMax; x++ {
				if ctx.Err() != nil {
					return
				}

				for y := min.Y; y < max.Y; y++ {
					b.buff[i] = b.Image.At(x, y)
					i++
				}
			}
		}()
	}

	wg.Wait()

	return ctx.Err()
}

func (i *BufferedImage) Opaque() bool {
	return true
}

func ToImage(img programs.Image) image.Image {
	scaleFactor := img.Bounds().Dx()
	if img.Bounds().Dy() > img.Bounds().Dx() {
		scaleFactor = img.Bounds().Dy()
	}

	return &imageImage{
		Image:       img,
		scaleFactor: float32(scaleFactor) / 2,
	}
}

type imageImage struct {
	programs.Image
	scaleFactor float32
}

func (i *imageImage) At(x, y int) color.Color {
	// oh how I wish I understood why this was needed
	y = -y

	c := i.GetPixel(mgl32.Vec2{
		float32(x) / i.scaleFactor,
		float32(y) / i.scaleFactor,
	})

	return color.NRGBA{
		R: uint8(c[0] * 255),
		G: uint8(c[1] * 255),
		B: uint8(c[2] * 255),
		A: 0xff,
	}
}

func (i *imageImage) ColorModel() color.Model {
	return color.NRGBAModel
}

func (i *imageImage) Opaque() bool {
	return true
}
