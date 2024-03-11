package program

import "image/color"

type Program interface {
	At(x int, y int) color.Color
	Name() string
	Shader() string
}

var programs []Program

func RegisterProgram(program Program) {
	programs = append(programs, program)
}

func Programs() []Program {
	return programs
}
