package main

import (
	"context"
	"encoding/gob"
	"fmt"
	"net"
	"sync"

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
		ctx:          ctx,
		quit:         quit,
		sendUniforms: make(chan programs.Uniforms),
	}

	go w.listen(listener)

	w.ApplicationWindow, err = gtk.ApplicationWindowNew(app)
	if err != nil {
		quit(fmt.Errorf("gtk.ApplicationWindowNew: %w", err))
	}
	w.SetDefaultSize(280, 700)

	w.Connect("realize", w.realize)

	g, _ := gtk.GridNew()

	label, _ := gtk.LabelNew("Program")
	programMenu, _ := gtk.ComboBoxTextNew()
	programMenu.Connect("changed", func(c *gtk.ComboBoxText) {

	})
	for _, program := range programs.Programs() {
		programMenu.AppendText(program.Name())
	}
	programMenu.SetActive(0)

	g.Attach(label, 0, 0, 1, 1)
	g.Attach(programMenu, 1, 0, 1, 1)

	label, _ = gtk.LabelNew("Colour Pallet")
	colourButton, _ := gtk.ButtonNewWithLabel("Randomize")
	colourButton.Connect("clicked", func(button *gtk.Button) {
		w.uniforms.ColourPallet = programs.RandomColourPallet()
		w.sendUniforms <- w.uniforms
	})
	g.Attach(label, 0, 1, 1, 1)
	g.Attach(colourButton, 1, 1, 1, 1)

	w.Add(g)
	w.ShowAll()

	return w
}

type ConfigWindow struct {
	*gtk.ApplicationWindow

	ctx  context.Context
	quit func(error)

	uniforms      programs.Uniforms
	uniformsMutex sync.Mutex
	sendUniforms  chan programs.Uniforms
}

func (w *ConfigWindow) realize(_ *gtk.ApplicationWindow) {

}

func (w *ConfigWindow) listen(listener net.Listener) {
	clients := make(map[net.Addr]*gob.Encoder, 1)
	var mutex sync.Mutex

	go func() {
		defer close(w.sendUniforms)

		for {
			select {
			case uniform := <-w.sendUniforms:
				mutex.Lock()
				for _, client := range clients {
					client.Encode(&uniform)
				}
				mutex.Unlock()
			case <-w.ctx.Done():
				listener.Close()
				return
			}
		}
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			w.quit(err)
			return
		}

		go func(conn net.Conn) {
			enc := gob.NewEncoder(conn)
			dec := gob.NewDecoder(conn)
			mutex.Lock()
			clients[conn.RemoteAddr()] = enc
			mutex.Unlock()

			for {
				var uniforms programs.Uniforms
				err := dec.Decode(&uniforms)
				if err != nil {
					mutex.Lock()
					delete(clients, conn.RemoteAddr())
					mutex.Unlock()
					return
				}

				w.uniformsMutex.Lock()
				w.uniforms = uniforms
				w.uniformsMutex.Unlock()
			}
		}(conn)
	}
}
