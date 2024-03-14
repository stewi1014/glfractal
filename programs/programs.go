package programs

import (
	_ "embed"
)

//go:embed default.vert
var defaultVertexShader string

type Program struct {
	Name           string
	VertexShader   string
	FragmentShader string
}

var Programs []Program

func registerProgram(program Program) {
	Programs = append(Programs, program)
}
