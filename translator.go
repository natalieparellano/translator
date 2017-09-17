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
		var res string
		switch command.Type {
		case "C_ARITHMETIC":
			res = codewriter.WriteArithmetic(command)
		case "C_PUSH", "C_POP":
			res = codewriter.WritePushPop(command)
		}
		ret += res
	}

	newpath := hackfile.NewPath(path, "vm", "asm")
	hackfile.WriteFile(newpath, ret)
}
