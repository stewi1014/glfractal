package programs

import (
	"embed"
	_ "embed"
	"io"
	"io/fs"
	"path"
	"strings"
)

//go:embed default.vert
var defaultVertexShader string

//go:embed shaders
var shaderFiles embed.FS

var programs []Program

func init() {
	shaders, err := fs.Glob(shaderFiles, "shaders/*.frag")
	if err != nil {
		panic(err)
	}

	for _, file := range shaders {
		frag, err := shaderFiles.Open(file)
		if err != nil {
			panic(err)
		}

		fragStr, err := io.ReadAll(frag)
		if err != nil {
			panic(err)
		}

		frag.Close()

		name, _ := strings.CutSuffix(path.Base(file), ".frag")

		NewProgram(Program{
			Name:           name,
			VertexShader:   defaultVertexShader,
			FragmentShader: string(fragStr),
		})
	}
}

type Program struct {
	Name           string
	VertexShader   string
	FragmentShader string
}

func NumPrograms() int {
	return len(programs)
}

func GetProgram(i int) Program {
	return programs[i]
}

func SetProgram(i int, p Program) error {
	programs[i] = p
	return nil
}

func NewProgram(p Program) error {
	programs = append(programs, p)
	return nil
}
