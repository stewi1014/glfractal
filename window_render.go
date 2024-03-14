package main

// #cgo pkg-config: gdk-3.0 glib-2.0 gobject-2.0
// #include <gdk/gdk.h>
import "C"

import (
	"context"
	"encoding/gob"
	"fmt"
	"log"
	"math"
	"net"
	"reflect"
	"runtime"
	"strings"
	"unsafe"

	"github.com/go-gl/gl/v4.6-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/stewi1014/glfractal/programs"
)

func NewRenderWindow(
	app *gtk.Application,
	conn net.Conn,
	ctx context.Context,
	quit func(error),
) *RenderWindow {
	var err error
	w := &RenderWindow{
		ctx:  ctx,
		quit: quit,
	}

	go w.handleSend(conn)

	w.ApplicationWindow, err = gtk.ApplicationWindowNew(app)
	if err != nil {
		quit(fmt.Errorf("gtk.ApplicationWindowNew: %w", err))
		return nil
	}

	w.SetDefaultSize(getWindowSize())

	w.gla, err = gtk.GLAreaNew()
	if err != nil {
		quit(fmt.Errorf("gtk.GLAreaNew: %w", err))
		return nil
	}

	w.gla.SetRequiredVersion(4, 6)
	w.gla.Connect("realize", w.glaRealize)
	w.gla.Connect("render", w.glaRender)
	w.gla.Connect("unrealize", w.glaUnrealize)

	w.gla.SetEvents(
		int(gdk.BUTTON_PRESS_MASK) |
			int(gdk.BUTTON_RELEASE_MASK) |
			int(gdk.SCROLL_MASK),
	)
	w.gla.Connect("resize", w.resize)
	w.gla.Connect("scroll-event", w.scroll)
	w.gla.Connect("button-press-event", w.button)
	w.gla.Connect("button-release-event", w.button)

	w.Add(w.gla)
	w.ShowAll()

	go w.handleReceive(conn)

	return w
}

func getWindowSize() (width, height int) {
	width = 1200
	height = 800

	display, err := gdk.DisplayGetDefault()
	if err != nil {
		return
	}

	monitor, err := display.GetPrimaryMonitor()
	if err != nil {
		return
	}

	width = int(float32(monitor.GetGeometry().GetWidth()) * .6)
	height = int(float32(monitor.GetGeometry().GetHeight()) * .6)
	return
}

type RenderWindow struct {
	*gtk.ApplicationWindow
	gla           *gtk.GLArea
	clickingMouse *gdk.Device
	clickPos      mgl32.Vec2
	width         int
	height        int

	ctx  context.Context
	quit func(error)

	vao              uint32
	vbo              uint32
	program          uint32
	vertexAttrib     uint32
	uniformLocations map[string]int32

	uniforms    programs.Uniforms
	sendMessage chan interface{}
}

func glDebugMessage(
	source,
	gltype,
	id,
	severity uint32,
	length int32,
	message string,
	user unsafe.Pointer,
) {
	severityStr := "unknown"
	switch severity {
	case gl.DEBUG_SEVERITY_HIGH:
		severityStr = "high"
	case gl.DEBUG_SEVERITY_LOW:
		severityStr = "low"
	case gl.DEBUG_SEVERITY_MEDIUM:
		severityStr = "medium"
	}

	sourceStr := "unknownSource"
	switch source {
	case gl.DEBUG_SOURCE_API:
		sourceStr = "api"
	case gl.DEBUG_SOURCE_APPLICATION:
		sourceStr = "application"
	case gl.DEBUG_SOURCE_OTHER:
		sourceStr = "other"
	case gl.DEBUG_SOURCE_SHADER_COMPILER:
		sourceStr = "shaderCompiler"
	case gl.DEBUG_SOURCE_THIRD_PARTY:
		sourceStr = "thirdParty"
	case gl.DEBUG_SOURCE_WINDOW_SYSTEM:
		sourceStr = "windowSystem"
	}

	typeStr := "unknownType"
	switch gltype {
	case gl.DEBUG_TYPE_ERROR:
		typeStr = "error"
	case gl.DEBUG_TYPE_DEPRECATED_BEHAVIOR:
		typeStr = "depreciatedBehavior"
	case gl.DEBUG_TYPE_MARKER:
		typeStr = "marker"
	case gl.DEBUG_TYPE_OTHER:
		typeStr = "other"
	case gl.DEBUG_TYPE_PERFORMANCE:
		typeStr = "performance"
	case gl.DEBUG_TYPE_POP_GROUP:
		typeStr = "popGroup"
	case gl.DEBUG_TYPE_PORTABILITY:
		typeStr = "portability"
	case gl.DEBUG_TYPE_PUSH_GROUP:
		typeStr = "pushGroup"
	case gl.DEBUG_TYPE_UNDEFINED_BEHAVIOR:
		typeStr = "undefinedBehavior"
	}

	log.Printf("%v(%v): %v; %v\n", sourceStr, severityStr, typeStr, message)
}

func (w *RenderWindow) glaRealize(gla *gtk.GLArea) {
	gla.MakeCurrent()

	err := gl.Init()
	if err != nil {
		w.quit(fmt.Errorf("gl.Init: %w", err))
		return
	}
	version := gl.GoStr(gl.GetString(gl.VERSION))
	fmt.Println("OpenGL version", version)

	gl.DebugMessageCallback(glDebugMessage, nil)
	if debug {
		gl.Enable(gl.DEBUG_OUTPUT)
	}

	verticies := []float32{
		-3, -2,
		0, 3,
		3, -2,
	}

	gl.GenVertexArrays(1, &w.vao)
	gl.BindVertexArray(w.vao)

	gl.GenBuffers(1, &w.vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, w.vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(verticies)*4, gl.Ptr(verticies), gl.STATIC_DRAW)

	err = w.loadProgram(programs.GetProgram(0))
	if err != nil {
		w.quit(err)
	}
}

func (w *RenderWindow) glaRender(gla *gtk.GLArea) {
	if w.clickingMouse != nil {
		pos := w.getMousePos()
		d := pos.Sub(w.clickPos)
		w.uniforms.Pos = w.uniforms.Pos.Add(mgl64.Vec2{float64(d.X()), -float64(d.Y())}.Mul(w.uniforms.Zoom * 2))
		w.clickPos = pos
		gla.QueueDraw()
	}

	w.gla.AttachBuffers()
	gl.Clear(gl.COLOR_BUFFER_BIT)
	gl.UseProgram(w.program)
	w.loadUniforms()
	gl.BindVertexArray(w.vao)
	gl.DrawArrays(gl.TRIANGLES, 0, 3)
}

func (w *RenderWindow) glaUnrealize(gla *gtk.GLArea) {}

func (w *RenderWindow) resize(gla *gtk.GLArea, width, height int) {
	w.width, w.height = width, height

	if w.height > w.width {
		w.uniforms.Camera = mgl32.Scale3D(float32(w.height)/float32(w.width), 1, 1)
	} else {
		w.uniforms.Camera = mgl32.Scale3D(1, float32(w.width)/float32(w.height), 1)
	}

	gl.Viewport(0, 0, int32(w.width), int32(w.height))
}

func (w *RenderWindow) getMousePos() mgl32.Vec2 {
	var x, y int
	screen := w.GetScreen()
	err := w.clickingMouse.GetPosition(&screen, &x, &y)
	if err != nil {
		log.Println(err)
		return mgl32.Vec2{}
	}

	if w.height > w.width {
		return mgl32.Vec2{float32(x) / float32(w.height), float32(y) / float32(w.height)}
	} else {
		return mgl32.Vec2{float32(x) / float32(w.width), float32(y) / float32(w.width)}
	}
}

func (w *RenderWindow) button(gla *gtk.GLArea, event *gdk.Event) {
	button := gdk.EventButtonNewFromEvent(event)
	gla.QueueRender()

	if button.Type() == gdk.EVENT_BUTTON_PRESS {
		C_GdkDevice := (*C.GdkEventButton)(unsafe.Pointer(button.Native())).device
		obj := &glib.Object{glib.ToGObject(unsafe.Pointer(C_GdkDevice))}
		w.clickingMouse = &gdk.Device{obj}
		w.clickPos = w.getMousePos()

	} else if button.Type() == gdk.EVENT_BUTTON_RELEASE {
		w.clickingMouse = nil
		w.sendMessage <- w.uniforms
	}
}

func (w *RenderWindow) scroll(gla *gtk.GLArea, event *gdk.Event) {
	scroll := gdk.EventScrollNewFromEvent(event)
	gla.QueueRender()

	if scroll.Direction() == gdk.SCROLL_DOWN {
		w.uniforms.Zoom += (w.uniforms.Zoom * .1)
	} else if scroll.Direction() == gdk.SCROLL_UP {
		w.uniforms.Zoom -= (w.uniforms.Zoom * .1)
	}

	if w.uniforms.Zoom > 2 {
		w.uniforms.Zoom = 2
	}
	if w.uniforms.Zoom*.1 == 0 {
		w.uniforms.Zoom = (1 / .1) * math.SmallestNonzeroFloat64
	}

	w.sendMessage <- w.uniforms
}

func (w *RenderWindow) handleSend(conn net.Conn) {
	enc := gob.NewEncoder(conn)
	w.sendMessage = make(chan interface{})
	defer conn.Close()
	defer close(w.sendMessage)
	defer w.quit(fmt.Errorf("unknown error"))

	for {
		select {
		case msg := <-w.sendMessage:
			err := enc.Encode(&msg)
			if err != nil {
				w.quit(err)
				return
			}
		case <-w.ctx.Done():
			return
		}
	}
}

func (w *RenderWindow) handleReceive(conn net.Conn) {
	dec := gob.NewDecoder(conn)
	defer w.quit(fmt.Errorf("unknown error"))

	for {
		var v interface{}
		err := dec.Decode(&v)
		if err != nil {
			w.quit(err)
			conn.Close()
			return
		}

		switch msg := v.(type) {
		case *programs.Program:
			glib.IdleAdd(func() {
				err := w.loadProgram(*msg)
				if err != nil {
					log.Println(err)
				}
				w.gla.QueueDraw()
			})

		case *programs.Uniforms:
			glib.IdleAdd(func() {
				w.uniforms = *msg
				w.resize(w.gla, w.width, w.height)
				w.gla.QueueDraw()
			})
		default:
			log.Println("unknown message received", reflect.TypeOf(v))
		}
	}
}

func (w *RenderWindow) loadUniforms() {
	v := reflect.ValueOf(&w.uniforms).Elem()
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)

		ptr := f.Addr().UnsafePointer()
		loc := w.uniformLocations[v.Type().Field(i).Tag.Get("uniform")]

		count := int32(1)

	SwitchElem:
		switch f.Type() {
		// Natural Array types
		case reflect.TypeOf(mgl32.Vec2{}):
			gl.Uniform2fv(loc, count, (*float32)(ptr))
			continue
		case reflect.TypeOf(mgl32.Vec3{}):
			gl.Uniform3fv(loc, count, (*float32)(ptr))
			continue
		case reflect.TypeOf(mgl32.Vec4{}):
			gl.Uniform4fv(loc, count, (*float32)(ptr))
			continue
		case reflect.TypeOf(mgl64.Vec2{}):
			gl.Uniform2dv(loc, count, (*float64)(ptr))
			continue
		case reflect.TypeOf(mgl64.Vec3{}):
			gl.Uniform3dv(loc, count, (*float64)(ptr))
			continue
		case reflect.TypeOf(mgl64.Vec4{}):
			gl.Uniform4dv(loc, count, (*float64)(ptr))
			continue
		case reflect.TypeOf(mgl32.Mat2{}):
			gl.UniformMatrix2fv(loc, count, false, (*float32)(ptr))
			continue
		case reflect.TypeOf(mgl32.Mat3{}):
			gl.UniformMatrix3fv(loc, count, false, (*float32)(ptr))
			continue
		case reflect.TypeOf(mgl32.Mat4{}):
			gl.UniformMatrix4fv(loc, count, false, (*float32)(ptr))
			continue
		case reflect.TypeOf(int32(0)):
			gl.Uniform1iv(loc, count, (*int32)(ptr))
			continue
		case reflect.TypeOf(uint32(0)):
			gl.Uniform1uiv(loc, count, (*uint32)(ptr))
			continue
		case reflect.TypeOf(float32(0)):
			gl.Uniform1fv(loc, count, (*float32)(ptr))
			continue
		case reflect.TypeOf(float64(0)):
			gl.Uniform1dv(loc, count, (*float64)(ptr))
			continue
		}

		if f.Kind() == reflect.Array {
			count = int32(f.Len())
			f = f.Index(0)
			goto SwitchElem
		}

		log.Printf("unsupported uniform type %v", f.Type())
	}
}

func (w *RenderWindow) loadProgram(program programs.Program) error {
	vertexShader, err := compileShader(program.VertexShader+"\x00", gl.VERTEX_SHADER)
	if err != nil {
		return err
	}

	fragmentShader, err := compileShader(program.FragmentShader+"\x00", gl.FRAGMENT_SHADER)
	if err != nil {
		return err
	}

	w.program = gl.CreateProgram()
	gl.AttachShader(w.program, vertexShader)
	gl.AttachShader(w.program, fragmentShader)
	gl.LinkProgram(w.program)
	gl.UseProgram(w.program)

	defer gl.DeleteShader(vertexShader)
	defer gl.DeleteShader(fragmentShader)

	var status int32
	gl.GetProgramiv(w.program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var l int32
		gl.GetProgramiv(w.program, gl.INFO_LOG_LENGTH, &l)

		log := strings.Repeat("\x00", int(l+1))
		gl.GetProgramInfoLog(w.program, l, nil, gl.Str(log))
		return fmt.Errorf("failed to link program: %v", log)
	}

	w.uniformLocations = make(map[string]int32)
	t := reflect.TypeOf(w.uniforms)
	w.uniforms.DefaultValues()
	w.resize(w.gla, w.width, w.height)
	for i := 0; i < t.NumField(); i++ {
		name := strings.ToLower(t.Field(i).Tag.Get("uniform"))
		w.uniformLocations[name] = gl.GetUniformLocation(w.program, gl.Str(name+"\x00"))
	}

	gl.BindFragDataLocation(w.program, 0, gl.Str("outputColor\x00"))

	w.vertexAttrib = uint32(gl.GetAttribLocation(w.program, gl.Str("vert\x00")))
	gl.EnableVertexAttribArray(w.vertexAttrib)
	gl.VertexAttribPointerWithOffset(w.vertexAttrib, 2, gl.FLOAT, false, 2*4, 0)

	w.sendMessage <- w.uniforms

	return nil
}

func compileShader(source string, shaderType uint32) (uint32, error) {
	defer runtime.KeepAlive(source)
	cstring, free := gl.Strs(source)
	defer free()

	shader := gl.CreateShader(shaderType)
	gl.ShaderSource(shader, 1, cstring, nil)
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var l int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &l)

		log := strings.Repeat("\x00", int(l+1))
		gl.GetShaderInfoLog(shader, l, nil, gl.Str(log))
		return 0, fmt.Errorf("shader\n\"\n%v\n\"\nfailed to compile: %v", source, log)
	}

	return shader, nil
}
