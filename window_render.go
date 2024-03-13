package main

import (
	"fmt"
	"log"
	"net"
	"reflect"
	"runtime"
	"strings"
	"unsafe"

	"github.com/go-gl/gl/v4.6-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
	"github.com/stewi1014/glfractal/programs"
)

var (
	typeFloat32   = reflect.TypeOf(float32(0))
	typeFloat64   = reflect.TypeOf(float64(0))
	typeInt32     = reflect.TypeOf(int32(0))
	typeUint32    = reflect.TypeOf(uint32(0))
	typeMgl32Vec2 = reflect.TypeOf(mgl32.Vec2{})
	typeMgl32Vec3 = reflect.TypeOf(mgl32.Vec3{})
	typeMgl32Vec4 = reflect.TypeOf(mgl32.Vec4{})
	typeMgl64Vec2 = reflect.TypeOf(mgl64.Vec2{})
	typeMgl64Vec3 = reflect.TypeOf(mgl64.Vec3{})
	typeMgl64Vec4 = reflect.TypeOf(mgl64.Vec4{})
	typeMgl32Mat2 = reflect.TypeOf(mgl32.Mat2{})
	typeMgl32Mat3 = reflect.TypeOf(mgl32.Mat3{})
	typeMgl32Mat4 = reflect.TypeOf(mgl32.Mat4{})
)

func NewRenderWindow(
	app *gtk.Application,
	conn net.Conn,
	quit func(error),
) *RenderWindow {
	w := &RenderWindow{
		quit: quit,
	}

	display, err := gdk.DisplayGetDefault()
	if err != nil {
		quit(fmt.Errorf("gdk.DisplayGetDefault %w", err))
		return nil
	}

	seat, err := display.GetDefaultSeat()
	if err != nil {
		quit(fmt.Errorf("gdk.Display.GetDefaultSeat: %w", err))
		return nil
	}

	w.mouse, err = seat.GetPointer()
	if err != nil {
		quit(fmt.Errorf("gdk.Seat.GetPointer: %w", err))
		return nil
	}

	w.ApplicationWindow, err = gtk.ApplicationWindowNew(app)
	if err != nil {
		quit(fmt.Errorf("gtk.ApplicationWindowNew: %w", err))
		return nil
	}

	w.SetDefaultSize(getWindowSize(display))

	gla, err := gtk.GLAreaNew()
	if err != nil {
		quit(fmt.Errorf("gtk.GLAreaNew: %w", err))
		return nil
	}

	gla.SetRequiredVersion(4, 6)
	gla.Connect("realize", w.glaRealize)
	gla.Connect("render", w.glaRender)
	gla.Connect("unrealize", w.glaUnrealize)

	gla.Connect("resize", w.resize)

	w.Add(gla)
	w.ShowAll()

	return w
}

func getWindowSize(display *gdk.Display) (width, height int) {
	monitor, err := display.GetPrimaryMonitor()
	if err == nil {
		width = int(float32(monitor.GetGeometry().GetWidth()) * .9)
		height = int(float32(monitor.GetGeometry().GetHeight()) * .9)
	} else {
		log.Println("gdk.Display.GetPrimaryMonitor: %w", err)
		width = 1200
		height = 800
	}

	return
}

type RenderWindow struct {
	*gtk.ApplicationWindow
	mouse *gdk.Device
	quit  func(error)

	vao     uint32
	vbo     uint32
	program uint32

	uniforms         programs.Uniforms
	uniformLocations map[string]int32
	vertexAttrib     uint32
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
		-1.0, -1.0,
		1.0, -1.0,
		1.0, 1.0,
		-1.0, -1.0,
		-1.0, 1.0,
		1.0, 1.0,
	}

	gl.GenVertexArrays(1, &w.vao)
	gl.BindVertexArray(w.vao)

	gl.GenBuffers(1, &w.vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, w.vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(verticies)*4, gl.Ptr(verticies), gl.STATIC_DRAW)

	err = w.loadProgram(programs.Programs()[0])
	if err != nil {
		w.quit(err)
	}
}

func (w *RenderWindow) glaRender(gla *gtk.GLArea) {
	w.loadUniforms()
	gl.Clear(gl.COLOR_BUFFER_BIT)
	gl.UseProgram(w.program)
	gl.BindVertexArray(w.vao)
	gl.DrawArrays(gl.TRIANGLES, 0, 6)
}

func (w *RenderWindow) glaUnrealize(gla *gtk.GLArea) {}

func (w *RenderWindow) resize(gla *gtk.GLArea, width, height int) {
	if height > width {
		w.uniforms.Camera = mgl32.Scale3D(float32(height)/float32(width), 1, 1)
	} else {
		w.uniforms.Camera = mgl32.Scale3D(1, float32(width)/float32(height), 1)
	}
	gl.Viewport(0, 0, int32(width), int32(height))
}

func (w *RenderWindow) button(gla *gtk.GLArea, event *gdk.Event) {

}

func (w *RenderWindow) scroll(gla *gtk.GLArea, event *gdk.Event) {

}

func (w *RenderWindow) loadUniforms() {
	v := reflect.ValueOf(&w.uniforms).Elem()
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)

		ptr := f.Addr().UnsafePointer()
		loc := w.uniformLocations[strings.ToLower(v.Type().Field(i).Name)]

		count := int32(1)

	SwitchElem:
		switch f.Type() {
		// Natural Array types
		case typeMgl32Vec2:
			gl.Uniform2fv(loc, count, (*float32)(ptr))
			continue
		case typeMgl32Vec3:
			gl.Uniform3fv(loc, count, (*float32)(ptr))
			continue
		case typeMgl32Vec4:
			gl.Uniform4fv(loc, count, (*float32)(ptr))
			continue
		case typeMgl64Vec2:
			gl.Uniform2dv(loc, count, (*float64)(ptr))
			continue
		case typeMgl64Vec3:
			gl.Uniform3dv(loc, count, (*float64)(ptr))
			continue
		case typeMgl64Vec4:
			gl.Uniform4dv(loc, count, (*float64)(ptr))
			continue
		case typeMgl32Mat2:
			gl.UniformMatrix2fv(loc, count, false, (*float32)(ptr))
			continue
		case typeMgl32Mat3:
			gl.UniformMatrix3fv(loc, count, false, (*float32)(ptr))
			continue
		case typeMgl32Mat4:
			gl.UniformMatrix4fv(loc, count, false, (*float32)(ptr))
			continue
		case typeInt32:
			gl.Uniform1iv(loc, count, (*int32)(ptr))
			continue
		case typeUint32:
			gl.Uniform1uiv(loc, count, (*uint32)(ptr))
			continue
		case typeFloat32:
			gl.Uniform1fv(loc, count, (*float32)(ptr))
			continue
		case typeFloat64:
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
	vertexShader, err := compileShader(programs.VertexShader+"\x00", gl.VERTEX_SHADER)
	if err != nil {
		return err
	}

	fragmentShader, err := compileShader(program.FragmentShader()+"\x00", gl.FRAGMENT_SHADER)
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
	for i := 0; i < t.NumField(); i++ {
		name := strings.ToLower(t.Field(i).Name)

		w.uniformLocations[name] = gl.GetUniformLocation(w.program, gl.Str(name+"\x00"))
	}

	gl.BindFragDataLocation(w.program, 0, gl.Str("outputColor\x00"))

	w.vertexAttrib = uint32(gl.GetAttribLocation(w.program, gl.Str("vert\x00")))
	gl.EnableVertexAttribArray(w.vertexAttrib)
	gl.VertexAttribPointerWithOffset(w.vertexAttrib, 2, gl.FLOAT, false, 2*4, 0)

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
