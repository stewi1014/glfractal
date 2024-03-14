package main

import (
	"context"
	"encoding/gob"
	"fmt"
	"log"
	"net"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/stewi1014/glfractal/programs"
)

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
	y := 0

	label, _ := gtk.LabelNew("Program")
	programMenu, _ := gtk.ComboBoxTextNew()
	for i := 0; i < programs.NumPrograms(); i++ {
		programMenu.AppendText(programs.GetProgram(i).Name)
	}
	programMenu.SetActive(0)
	programMenu.Connect("changed", func(c *gtk.ComboBoxText) {
		w.program = programs.GetProgram(c.GetActive())
		w.sendMessage <- w.program
	})
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

	for i := range w.uniforms.Sliders {
		label, _ := gtk.LabelNew(fmt.Sprintf("Slider %v", i))
		slider, _ := gtk.ScaleNewWithRange(gtk.ORIENTATION_HORIZONTAL, -1, 1, 0.0001)
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
}

func (w *ConfigWindow) realize(_ *gtk.ApplicationWindow) {

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
