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
		ctx:  ctx,
		quit: quit,
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
	})
	programMenu.SetHExpand(true)
	g.Attach(label, 0, y, 1, 1)
	g.Attach(programMenu, 1, y, 1, 1)
	y++

	label, _ = gtk.LabelNew("Colour Pallet")
	colourButton, _ := gtk.ButtonNewWithLabel("Randomize")
	colourButton.Connect("clicked", func(button *gtk.Button) {
		w.uniforms.ColourPallet = programs.RandomColourPallet()
		w.sendMessage <- w.uniforms
	})
	g.Attach(label, 0, y, 1, 1)
	g.Attach(colourButton, 1, y, 1, 1)
	y++

	label, _ = gtk.LabelNew("Iterations")
	iterationsButton, _ := gtk.SpinButtonNewWithRange(0, 10000, 1)
	iterationsButton.SetValue(500)
	iterationsButton.Connect("value-changed", func(b *gtk.SpinButton) {
		w.uniforms.Iterations = uint32(b.GetValueAsInt())
		w.sendMessage <- w.uniforms
	})
	g.Attach(label, 0, y, 1, 1)
	g.Attach(iterationsButton, 1, y, 1, 1)
	y++

	seperator, _ := gtk.SeparatorNew(gtk.ORIENTATION_HORIZONTAL)
	g.Attach(seperator, 0, y, 2, 1)
	y++

	for i := range w.uniforms.Sliders {
		label, _ := gtk.LabelNew(fmt.Sprintf("Slider %v", i))
		slider, _ := gtk.ScaleNewWithRange(gtk.ORIENTATION_HORIZONTAL, -2, 2, 0.00001)
		slider.SetValue(0)
		slider.Connect("value-changed", func(s *gtk.Scale) {
			w.uniforms.Sliders[i] = s.GetValue()
			w.sendMessage <- w.uniforms
		})

		slider.SetSizeRequest(300, 20)

		g.Attach(label, 0, y, 1, 1)
		g.Attach(slider, 1, y, 1, 1)

		y++
	}

	seperator, _ = gtk.SeparatorNew(gtk.ORIENTATION_HORIZONTAL)
	g.Attach(seperator, 0, y, 2, 1)
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
	g.Attach(widthEntry, 1, y, 1, 1)
	y++
	label, _ = gtk.LabelNew("Height")
	g.Attach(label, 0, y, 1, 1)
	g.Attach(heightEntry, 1, y, 1, 1)
	y++

	w.Add(g)
	w.ShowAll()
	w.SetKeepAbove(true)

	return w
}

type ConfigWindow struct {
	*gtk.ApplicationWindow

	ctx  context.Context
	quit func(error)

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
