package programs

import (
	_ "embed"
	"image/color"
)

//go:embed default.vert
var VertexShader string

type Program interface {
	At(x int, y int) color.Color
	Name() string
	FragmentShader() string
}

var programs []Program

func RegisterProgram(program Program) {
	programs = append(programs, program)
}

func Programs() []Program {
	return programs
}
