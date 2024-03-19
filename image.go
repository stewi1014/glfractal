package main

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"log"
	"runtime"
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/gotk3/gotk3/gdk"
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
	count atomic.Uint64
}

func (i *ProgressImage) At(x, y int) color.Color {
	i.count.Add(1)
	return i.Image.At(x, y)
}

func (i *ProgressImage) Progress() float64 {
	end := i.Bounds().Dx() * i.Bounds().Dy()
	return float64(i.count.Load()) / float64(end)
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
		Image: img,
	}
}

type BufferedImage struct {
	image.Image
	rowstride int
	buff      *gdk.Pixbuf
	asSlice   []byte
}

func (b *BufferedImage) Bounds() image.Rectangle {
	return image.Rect(0, 0, b.Image.Bounds().Dx(), b.Image.Bounds().Dy())
}

func (b *BufferedImage) At(x, y int) color.Color {
	return (*color.NRGBA)(unsafe.Pointer(&b.asSlice[y*b.rowstride+x*4]))
}

func (b *BufferedImage) Buffer(ctx context.Context) error {
	ctx, quit := context.WithCancelCause(ctx)
	AttachErrorDialog(nil, ctx)

	var err error
	b.buff, err = gdk.PixbufNew(
		gdk.COLORSPACE_RGB,
		true,
		8,
		b.Bounds().Dx(),
		b.Bounds().Dy(),
	)
	if err != nil {
		return err
	}

	if b.buff.GetNChannels() != 4 {
		return fmt.Errorf("gdk.Pixbuf does not have 4 channels")
	}

	b.rowstride = b.buff.GetRowstride()
	b.asSlice = b.buff.GetPixels()

	min, max := b.Image.Bounds().Min, b.Image.Bounds().Max
	chunkSize := 50 * 2000 / image.Black.Bounds().Dx()
	if chunkSize == 0 {
		chunkSize = 1
	}

	numThreads := int(float64(runtime.NumCPU()-1) * .9)
	if numThreads < 1 {
		numThreads = 1
	}

	threads := make(chan struct{}, numThreads)
	var wg sync.WaitGroup

	for chunkMin := min.Y; chunkMin < max.Y; chunkMin += chunkSize {
		chunkMax := chunkMin + chunkSize
		if chunkMax > max.Y {
			chunkMax = max.Y
		}

		wg.Add(1)
		go func() {
			threads <- struct{}{}
			defer func() { <-threads; wg.Done() }()
			defer CatchPanicToContext(quit)

			if ctx.Err() != nil {
				return
			}

			i := (chunkMin - min.Y) * b.rowstride
			for y := chunkMin; y < chunkMax; y++ {
				for x := min.X; x < max.X; x++ {
					c := b.Image.At(x, y)
					if nrgba, ok := c.(color.NRGBA); ok {
						*(*color.NRGBA)(unsafe.Pointer(&b.asSlice[i])) = nrgba
					}
					i += 4
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

func (i *BufferedImage) Scale(dest *gdk.Pixbuf, width, height int, interpType gdk.InterpType) {
	if i.buff == nil {
		return
	}

	i.buff.Scale(
		dest,
		0, 0,
		width,
		height,
		0, 0,
		float64(dest.GetWidth())/float64(i.buff.GetWidth()),
		float64(dest.GetHeight())/float64(i.buff.GetHeight()),
		interpType,
	)
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
