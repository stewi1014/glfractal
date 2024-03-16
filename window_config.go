package main

import (
	"context"
	"encoding/gob"
	"fmt"
	"image/png"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/stewi1014/glfractal/programs"
)

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
	pd.Connect("response", pd.Destroy)

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
		colourSeed:     time.Now().Unix(),
		colourWalkRate: 0.3,
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

	colourStartR, _ := gtk.SpinButtonNewWithRange(0, 1, 0.01)
	colourStartR.SetValue(float64(w.startingColour[0]))
	colourStartR.Connect("value-changed", func(b *gtk.SpinButton) {
		w.startingColour[0] = float32(b.GetValue())
		w.generateColour()
	})

	colourStartG, _ := gtk.SpinButtonNewWithRange(0, 1, 0.01)
	colourStartG.SetValue(float64(w.startingColour[1]))
	colourStartG.Connect("value-changed", func(b *gtk.SpinButton) {
		w.startingColour[1] = float32(b.GetValue())
		w.generateColour()
	})

	colourStartB, _ := gtk.SpinButtonNewWithRange(0, 1, 0.01)
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

	for i := range w.uniforms.Sliders {
		label, _ := gtk.LabelNew(fmt.Sprintf("Slider %v", i))
		slider, _ := gtk.ScaleNewWithRange(gtk.ORIENTATION_HORIZONTAL, -3, 3, 0.000001)
		slider.SetValue(0)
		slider.Connect("value-changed", func(s *gtk.Scale) {
			w.uniforms.Sliders[i] = s.GetValue()
			w.sendMessage <- w.uniforms
		})

		slider.SetSizeRequest(300, 20)

		g.Attach(label, 0, y, 1, 1)
		g.Attach(slider, 1, y, 3, 1)

		y++
	}

	seperator, _ = gtk.SeparatorNew(gtk.ORIENTATION_HORIZONTAL)
	g.Attach(seperator, 0, y, 4, 1)
	y++

	w.saveWidth = 15360
	w.saveHeight = 8640
	label, _ = gtk.LabelNew("Save")
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
	widthEntry, _ := gtk.EntryNew()
	widthEntry.SetText(strconv.Itoa(w.saveWidth))
	widthEntry.Connect("changed", func(e *gtk.Entry) {
		w.saveWidth = parseNumber(e)
	})
	heightEntry, _ := gtk.EntryNew()
	heightEntry.SetText(strconv.Itoa(w.saveHeight))
	heightEntry.Connect("changed", func(e *gtk.Entry) {
		w.saveHeight = parseNumber(e)
	})

	g.Attach(label, 0, y, 1, 1)
	g.Attach(saveButton, 1, y, 1, 1)
	y++
	label, _ = gtk.LabelNew("Width")
	g.Attach(label, 0, y, 1, 1)
	g.Attach(widthEntry, 1, y, 3, 1)
	y++
	label, _ = gtk.LabelNew("Height")
	g.Attach(label, 0, y, 1, 1)
	g.Attach(heightEntry, 1, y, 3, 1)
	y++

	w.Add(g)
	w.ShowAll()
	w.SetKeepAbove(true)

	w.uniforms.DefaultValues()
	w.generateColour()
	w.sendMessage <- w.uniforms

	return w
}

type ConfigWindow struct {
	*gtk.ApplicationWindow

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

func (w *ConfigWindow) save(pd *ProgressDialog) error {
	image, err := w.program.GetImage(w.uniforms, w.saveWidth, w.saveHeight)
	if err != nil {
		return err
	}

	if w.saveName == "" {
		w.saveName = w.getSaveName()
	}

	file, err := os.Create(w.saveName)
	if err != nil {
		return err
	}

	pd.Connect("response", func() {
		file.Close()
	})

	done := false
	go func() {
		err := png.Encode(file, image)
		if err != nil {
			log.Println(err)
		}
		fmt.Println(image.Progress())
		file.Close()
		done = true
	}()

	glib.IdleAdd(func() bool {
		if done {
			pd.Destroy()
			return false
		}
		progress := image.Progress()
		pd.l.SetText(fmt.Sprintf("saving %v: %2.2f%%", file.Name(), progress*100))
		pd.pb.SetText(fmt.Sprintf("saving %v: %2.2f%%", file.Name(), progress*100))
		pd.pb.SetFraction(image.Progress())
		return true
	})

	w.saveName = w.getSaveName()

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
				w.uniforms = *msg
			})
			w.sendMessage <- skipClient{
				msg:  *msg,
				addr: conn.RemoteAddr(),
			}
		}
	}
}

func (w *ConfigWindow) listen(listener net.Listener) {
	w.sendMessage = make(chan interface{})
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
			case client := <-newClient:
				clients[client.RemoteAddr()] = struct {
					conn net.Conn
					enc  *gob.Encoder
				}{
					conn: client,
					enc:  gob.NewEncoder(client),
				}

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
