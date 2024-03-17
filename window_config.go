package main

import (
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/stewi1014/glfractal/programs"
)

var imageSizePresets = []struct {
	name   string
	width  int
	height int
}{{
	name:   "720p",
	width:  1280,
	height: 720,
}, {
	name:   "1080p",
	width:  1920,
	height: 1080,
}, {
	name:   "4K",
	width:  3840,
	height: 2160,
}, {
	name:   "8K",
	width:  7680,
	height: 4320,
}, {
	name:   "16K",
	width:  15360,
	height: 8640,
}, {
	name:   "32k",
	width:  30720,
	height: 17280,
}, {
	name:   "64K",
	width:  61440,
	height: 34560,
}}

func parseNumber(e *gtk.Entry) int {
	str, err := e.GetText()
	if err != nil {
		log.Println(err)
		return 0
	}

	str = strings.ReplaceAll(str, ",", "")
	str = strings.ReplaceAll(str, ".", "")
	str = strings.TrimSpace(str)

	n, err := strconv.Atoi(str)
	if err != nil {
		log.Println(err)
	}

	return n
}

func NewErrorDialog(
	parent *gtk.ApplicationWindow,
	title string,
	err error,
) {
	d, _ := gtk.DialogNewWithButtons(
		title,
		parent,
		gtk.DIALOG_DESTROY_WITH_PARENT,
		[]interface{}{"OK", gtk.RESPONSE_OK},
	)
	d.Connect("response", d.Destroy)

	ca, _ := d.GetContentArea()
	l, _ := gtk.LabelNew(err.Error())
	ca.Add(l)
	d.SetKeepAbove(true)
	d.ShowAll()
}

func NewProgressDialog(
	parent *gtk.ApplicationWindow,
	title string,
) *ProgressDialog {
	pd := &ProgressDialog{}
	pd.Dialog, _ = gtk.DialogNewWithButtons(
		title,
		parent,
		gtk.DIALOG_DESTROY_WITH_PARENT,
		[]interface{}{"CANCEL", gtk.RESPONSE_CANCEL},
	)

	ca, _ := pd.GetContentArea()
	pd.l, _ = gtk.LabelNew("starting...")
	ca.Add(pd.l)
	pd.pb, _ = gtk.ProgressBarNew()
	pd.pb.SetSizeRequest(500, 80)
	ca.Add(pd.pb)
	pd.SetKeepAbove(true)
	pd.ShowAll()

	return pd
}

type ProgressDialog struct {
	*gtk.Dialog
	pb *gtk.ProgressBar
	l  *gtk.Label
}

func NewConfigWindow(
	app *gtk.Application,
	listener net.Listener,
	ctx context.Context,
	quit func(error),
) *ConfigWindow {
	var err error
	w := &ConfigWindow{
		ctx:            ctx,
		quit:           quit,
		app:            app,
		colourSeed:     time.Now().Unix(),
		colourWalkRate: 0.3,
		saveAntiAlias:  float32(1) / 3,
		startingColour: mgl32.Vec3{
			rand.Float32(),
			rand.Float32(),
			rand.Float32(),
		},
	}

	go w.listen(listener)

	w.ApplicationWindow, err = gtk.ApplicationWindowNew(app)
	if err != nil {
		quit(fmt.Errorf("gtk.ApplicationWindowNew: %w", err))
	}
	w.Connect("realize", w.realize)

	g, _ := gtk.GridNew()
	g.SetRowSpacing(10)
	g.SetHExpand(true)
	y := 0

	label, _ := gtk.LabelNew("Program")
	programMenu, _ := gtk.ComboBoxTextNew()
	for i := 0; i < programs.NumPrograms(); i++ {
		programMenu.AppendText(programs.GetProgram(i).Name)
	}
	w.program = programs.GetProgram(0)
	programMenu.SetActive(0)
	programMenu.Connect("changed", func(c *gtk.ComboBoxText) {
		w.program = programs.GetProgram(c.GetActive())
		w.sendMessage <- w.program
		w.sendMessage <- w.uniforms
	})
	programMenu.SetHExpand(true)
	g.Attach(label, 0, y, 1, 1)
	g.Attach(programMenu, 1, y, 3, 1)
	y++

	seperator, _ := gtk.SeparatorNew(gtk.ORIENTATION_HORIZONTAL)
	g.Attach(seperator, 0, y, 4, 1)
	y++

	colourStartR, _ := gtk.SpinButtonNewWithRange(0, 1, 0.06)
	colourStartR.SetValue(float64(w.startingColour[0]))
	colourStartR.Connect("value-changed", func(b *gtk.SpinButton) {
		w.startingColour[0] = float32(b.GetValue())
		w.generateColour()
	})

	colourStartG, _ := gtk.SpinButtonNewWithRange(0, 1, 0.06)
	colourStartG.SetValue(float64(w.startingColour[1]))
	colourStartG.Connect("value-changed", func(b *gtk.SpinButton) {
		w.startingColour[1] = float32(b.GetValue())
		w.generateColour()
	})

	colourStartB, _ := gtk.SpinButtonNewWithRange(0, 1, 0.06)
	colourStartB.SetValue(float64(w.startingColour[2]))
	colourStartB.Connect("value-changed", func(b *gtk.SpinButton) {
		w.startingColour[2] = float32(b.GetValue())
		w.generateColour()
	})

	label, _ = gtk.LabelNew("Colour Pallet")
	colourSeedButton, _ := gtk.ButtonNewWithLabel("Randomize Seed")
	colourSeedButton.Connect("clicked", func(button *gtk.Button) {
		w.colourSeed = rand.Int63()
		w.generateColour()
	})
	colourStartButton, _ := gtk.ButtonNewWithLabel("Randomize Start")
	colourStartButton.Connect("clicked", func(button *gtk.Button) {
		w.startingColour = mgl32.Vec3{
			rand.Float32(),
			rand.Float32(),
			rand.Float32(),
		}
		colourStartR.SetValue(float64(w.startingColour[0]))
		colourStartG.SetValue(float64(w.startingColour[1]))
		colourStartB.SetValue(float64(w.startingColour[2]))
		w.generateColour()
	})
	colourWalkRate, _ := gtk.SpinButtonNewWithRange(0, 1, 0.02)
	colourWalkRate.SetValue(0.3)
	colourWalkRate.Connect("value-changed", func(b *gtk.SpinButton) {
		w.colourWalkRate = float32(b.GetValue())
		w.generateColour()
	})
	g.Attach(label, 0, y, 1, 1)
	g.Attach(colourWalkRate, 1, y, 1, 1)
	g.Attach(colourSeedButton, 2, y, 1, 1)
	g.Attach(colourStartButton, 3, y, 1, 1)
	y++
	label, _ = gtk.LabelNew("Start R,G,B")
	g.Attach(label, 0, y, 1, 1)
	g.Attach(colourStartR, 1, y, 1, 1)
	g.Attach(colourStartG, 2, y, 1, 1)
	g.Attach(colourStartB, 3, y, 1, 1)
	y++

	seperator, _ = gtk.SeparatorNew(gtk.ORIENTATION_HORIZONTAL)
	g.Attach(seperator, 0, y, 4, 1)
	y++

	label, _ = gtk.LabelNew("Iterations")
	iterationsButton, _ := gtk.SpinButtonNewWithRange(0, 10000, 1)
	iterationsButton.SetValue(500)
	iterationsButton.Connect("value-changed", func(b *gtk.SpinButton) {
		w.uniforms.Iterations = uint32(b.GetValueAsInt())
		w.sendMessage <- w.uniforms
	})
	g.Attach(label, 0, y, 1, 1)
	g.Attach(iterationsButton, 1, y, 3, 1)
	y++

	sliders := make([]*gtk.Scale, len(w.uniforms.Sliders))
	for i := range w.uniforms.Sliders {
		label, _ := gtk.LabelNew(fmt.Sprintf("Slider %v", i))
		sliders[i], _ = gtk.ScaleNewWithRange(gtk.ORIENTATION_HORIZONTAL, -2, 2, 0.00000001)
		sliders[i].SetProperty("digits", 7)
		sliders[i].SetValue(0)
		sliders[i].Connect("value-changed", func(s *gtk.Scale) {
			w.uniforms.Sliders[i] = s.GetValue()
			w.sendMessage <- w.uniforms
		})

		sliders[i].SetSizeRequest(300, 20)

		g.Attach(label, 0, y, 1, 1)
		g.Attach(sliders[i], 1, y, 3, 1)

		y++
	}

	sliderReset, _ := gtk.ButtonNewWithLabel("Reset Sliders")
	sliderReset.Connect("clicked", func(button *gtk.Button) {
		for i := range w.uniforms.Sliders {
			w.uniforms.Sliders[i] = 0
			sliders[i].SetValue(0)
		}
		w.sendMessage <- w.uniforms
	})

	positionReset, _ := gtk.ButtonNewWithLabel("Reset Position")
	positionReset.Connect("clicked", func(button *gtk.Button) {
		w.uniforms.Zoom = 2
		w.uniforms.Pos = mgl64.Vec2{}
		w.sendMessage <- w.uniforms
	})

	g.Attach(sliderReset, 1, y, 1, 1)
	g.Attach(positionReset, 2, y, 1, 1)
	y++

	seperator, _ = gtk.SeparatorNew(gtk.ORIENTATION_HORIZONTAL)
	g.Attach(seperator, 0, y, 4, 1)
	y++

	label, _ = gtk.LabelNew("Image Render")
	defaultImageSize := 2
	widthEntry, _ := gtk.EntryNew()
	widthEntry.SetText(strconv.Itoa(imageSizePresets[defaultImageSize].width))
	w.saveWidth = imageSizePresets[defaultImageSize].width
	widthEntry.Connect("changed", func(e *gtk.Entry) {
		w.saveWidth = parseNumber(e)
	})
	heightEntry, _ := gtk.EntryNew()
	heightEntry.SetText(strconv.Itoa(imageSizePresets[defaultImageSize].height))
	w.saveHeight = imageSizePresets[defaultImageSize].height
	heightEntry.Connect("changed", func(e *gtk.Entry) {
		w.saveHeight = parseNumber(e)
	})
	saveButton, _ := gtk.ButtonNewWithLabel("Save")
	saveButton.Connect("clicked", func() {
		pd := NewProgressDialog(
			w.ApplicationWindow,
			"Saving Image",
		)

		err := w.save(pd)
		if err != nil {
			NewErrorDialog(w.ApplicationWindow, "Save Error", err)
			return
		}
	})
	imageSizeChooser, _ := gtk.ComboBoxTextNew()
	for _, preset := range imageSizePresets {
		imageSizeChooser.AppendText(preset.name)
	}
	imageSizeChooser.SetActive(defaultImageSize)
	imageSizeChooser.Connect("changed", func(c *gtk.ComboBoxText) {
		imageSize := imageSizePresets[c.GetActive()]
		widthEntry.SetText(strconv.Itoa(imageSize.width))
		w.saveWidth = imageSize.width
		heightEntry.SetText(strconv.Itoa(imageSize.height))
		w.saveWidth = imageSize.width
	})
	imageAntiAlias, _ := gtk.CheckButtonNewWithLabel("Antialias")
	imageAntiAlias.SetActive(true)
	w.saveAntiAlias = float32(1) / 3
	imageAntiAlias.Connect("toggled", func(b *gtk.CheckButton) {
		if b.GetActive() {
			w.saveAntiAlias = float32(1) / 3
		} else {
			w.saveAntiAlias = 0
		}
	})
	imageMultithread, _ := gtk.CheckButtonNewWithLabel("Multithread")
	imageMultithread.SetTooltipText("Multithreading requires buffering entire image in memory")
	imageMultithread.Connect("toggled", func(b *gtk.CheckButton) {
		w.saveMultithread = b.GetActive()
	})
	g.Attach(label, 0, y, 1, 1)
	g.Attach(saveButton, 1, y, 1, 1)
	g.Attach(imageAntiAlias, 2, y, 1, 1)
	g.Attach(imageMultithread, 3, y, 1, 1)
	y++
	label, _ = gtk.LabelNew("Size")
	g.Attach(label, 0, y, 1, 1)

	b, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 10)
	label, _ = gtk.LabelNew("x")
	b.Add(widthEntry)
	b.Add(label)
	b.Add(heightEntry)

	g.Attach(imageSizeChooser, 1, y, 1, 1)
	g.Attach(b, 2, y, 2, 1)
	y++

	w.Add(g)
	w.ShowAll()
	w.SetKeepAbove(true)

	w.uniforms.DefaultValues()
	w.generateColour()

	return w
}

type ConfigWindow struct {
	*gtk.ApplicationWindow
	app *gtk.Application

	ctx  context.Context
	quit func(error)

	colourSeed     int64
	colourWalkRate float32
	startingColour mgl32.Vec3

	uniforms    programs.Uniforms
	program     programs.Program
	sendMessage chan interface{}

	saveWidth, saveHeight int
	saveName              string
	saveAntiAlias         float32
	saveMultithread       bool
}

func (w *ConfigWindow) realize(_ *gtk.ApplicationWindow) {

}

func (w *ConfigWindow) getSaveName() string {
	return fmt.Sprintf(
		"fractal_%v_%vx%v_%v.png",
		w.program.Name,
		w.saveWidth,
		w.saveHeight,
		rand.Intn(10000),
	)
}

func (w *ConfigWindow) generateColour() {
	w.uniforms.ColourPallet = programs.RandomColourPallet(
		w.startingColour,
		w.colourWalkRate,
		rand.New(rand.NewSource(w.colourSeed)),
	)

	w.sendMessage <- w.uniforms
}

func (w *ConfigWindow) encodeImage(writer io.Writer, img image.Image, ctx context.Context) error {
	if !w.saveMultithread {
		return png.Encode(writer, img)
	}

	buffered := bufferImage(img, ctx)
	return png.Encode(writer, buffered)
}

func bufferImage(img image.Image, ctx context.Context) image.Image {
	width, height := img.Bounds().Dx(), img.Bounds().Dy()

	imgBuff := imageBuffer{
		bounds: image.Rect(0, 0, width, height),
		buff:   make([]color.RGBA, width*height),
	}

	min, max := img.Bounds().Min, img.Bounds().Max
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
			i := (chunkMin - min.X) * height
			for x := chunkMin; x < chunkMax; x++ {
				if ctx.Err() != nil {
					return
				}

				for y := min.Y; y < max.Y; y++ {
					imgBuff.buff[i] = img.At(x, y).(color.RGBA)
					i++
				}
			}
		}()
	}

	wg.Wait()

	return imgBuff
}

type imageBuffer struct {
	bounds image.Rectangle
	buff   []color.RGBA
}

func (b imageBuffer) Bounds() image.Rectangle {
	return b.bounds
}

func (b imageBuffer) At(x, y int) color.Color {
	return b.buff[x*b.bounds.Max.Y+y]
}

func (b imageBuffer) ColorModel() color.Model {
	return color.RGBAModel
}

func (b imageBuffer) Opaque() bool {
	return false
}

func (w *ConfigWindow) save(pd *ProgressDialog) error {
	ctx, done := context.WithCancel(context.Background())

	image, err := w.program.GetImage(w.uniforms, w.saveWidth, w.saveHeight, w.saveAntiAlias)
	if err != nil {
		done()
		return err
	}

	name := w.saveName

	if name == "" {
		name = w.getSaveName()
	} else if _, err := os.Stat(name); !errors.Is(err, os.ErrNotExist) {
		name = w.getSaveName()
	}

	file, err := os.Create(name)
	if err != nil {
		done()
		return err
	}

	pd.Connect("response", func() {
		defer pd.Destroy()
		file.Close()
		err := os.Remove(file.Name())
		if err != nil {
			log.Println(err)
		}
		done()
	})

	go func() {
		err := w.encodeImage(file, image, ctx)
		if err != nil && ctx.Err() == nil {
			glib.IdleAdd(func() {
				NewErrorDialog(w.ApplicationWindow, "Rendering Image", err)
			})
		}
		file.Close()
		done()
	}()

	glib.IdleAdd(func() bool {
		if ctx.Err() != nil {
			pd.Destroy()

			previewWindow, err := gtk.ApplicationWindowNew(w.app)
			if err != nil {
				return false
			}

			width, height := getDisplaySize()
			pixbuf, err := gdk.PixbufNewFromFileAtSize(name, int(float64(width)*.8), int(float64(height)*.8))
			if err != nil {
				return false
			}

			previewImage, err := gtk.ImageNewFromPixbuf(pixbuf)
			if err != nil {
				return false
			}
			previewImage.SetHExpand(true)
			previewImage.SetVExpand(true)
			previewImage.SetSizeRequest(1280, 720)

			deleteButton, _ := gtk.ButtonNewWithLabel("Delete")
			deleteButton.Connect("clicked", func(button *gtk.Button) {
				err := os.Remove(name)
				if err != nil {
					NewErrorDialog(w.ApplicationWindow, "Deleting File", err)
				}
				previewWindow.Destroy()
			})

			grid, _ := gtk.GridNew()
			grid.Attach(previewImage, 0, 0, 5, 1)
			grid.Attach(deleteButton, 4, 1, 1, 1)

			previewWindow.Add(grid)
			previewWindow.ShowAll()

			return false
		}
		progress := image.Progress()
		pd.l.SetText(fmt.Sprintf("saving %v: %2.2f%%", file.Name(), progress*100))
		pd.pb.SetFraction(image.Progress())
		return true
	})

	return nil
}

type skipClient struct {
	msg  interface{}
	addr net.Addr
}

func (w *ConfigWindow) handleReceive(conn net.Conn) {
	defer conn.Close()
	dec := gob.NewDecoder(conn)

	for {
		var v interface{}
		err := dec.Decode(&v)
		if err != nil {
			log.Println(err)
			return
		}

		switch msg := v.(type) {
		case *programs.Uniforms:
			glib.IdleAdd(func() {
				w.uniforms.Zoom = msg.Zoom
				w.uniforms.Pos = msg.Pos
				w.sendMessage <- skipClient{
					msg:  *msg,
					addr: conn.RemoteAddr(),
				}
			})
		}
	}
}

func (w *ConfigWindow) listen(listener net.Listener) {
	w.sendMessage = make(chan interface{}, 10)
	defer close(w.sendMessage)
	defer listener.Close()
	defer w.quit(fmt.Errorf("unknown error"))

	newClient := make(chan net.Conn)

	go func() {
		clients := make(map[net.Addr]struct {
			conn net.Conn
			enc  *gob.Encoder
		})

		for {
			select {
			case conn := <-newClient:
				client := struct {
					conn net.Conn
					enc  *gob.Encoder
				}{
					conn: conn,
					enc:  gob.NewEncoder(conn),
				}

				var msg interface{}
				msg = w.uniforms
				client.enc.Encode(&msg)
				msg = w.program
				client.enc.Encode(&msg)

				clients[conn.RemoteAddr()] = client

			case msg, ok := <-w.sendMessage:
				if !ok {
					return
				}

				var skip net.Addr
				if sc, ok := msg.(skipClient); ok {
					skip = sc.addr
					msg = sc.msg
				}

				for addr, client := range clients {
					if addr == skip {
						continue
					}

					err := client.enc.Encode(&msg)
					if err != nil {
						client.conn.Close()
						delete(clients, addr)
						continue
					}
				}

			case <-w.ctx.Done():
				return
			}
		}
	}()

	for {
		client, err := listener.Accept()
		if err != nil {
			w.quit(err)
			return
		}

		go w.handleReceive(client)
		newClient <- client
	}
}
