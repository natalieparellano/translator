package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/natalieparellano/assembler/hackfile"
	"github.com/natalieparellano/translator/codewriter"
	"github.com/natalieparellano/translator/parser"
)

func main() {
	if len(os.Args) != 2 {
		panic("USAGE: provide path to .vm file")
	}
	path := os.Args[1]
	fmt.Printf("Parsing file: %s\n", path)
	basename := filepath.Base(path)
	codewriter.Filename = strings.TrimSuffix(basename, filepath.Ext(basename))

	commands := parser.Parse(path)
	var ret string

	for _, command := range commands {
		ret += codewriter.Write(command)
	}

	newpath := hackfile.NewPath(path, "vm", "asm")
	hackfile.WriteFile(newpath, ret)
}
