package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/natalieparellano/assembler/hackfile"
	"github.com/natalieparellano/translator/codewriter"
	"github.com/natalieparellano/translator/parser"
)

func parseFile(path string) string {
	fmt.Printf("Parsing file: %s\n", path)
	basename := filepath.Base(path)
	codewriter.Filename = strings.TrimSuffix(basename, filepath.Ext(basename))

	commands := parser.Parse(path)
	var ret string

	for _, command := range commands {
		fmt.Printf("Command: type: %s, arg1: %s, arg2: %d\n",
			command.Type, command.Arg1, command.Arg2)

		toWrite := codewriter.Write(command)
		fmt.Printf("Assembly: %s\n", toWrite)

		ret += toWrite
	}

	return ret
}

func filterFiles(input []os.FileInfo, f func(os.FileInfo) bool) []string {
	output := make([]string, 0)
	for _, el := range input {
		if f(el) {
			output = append(output, el.Name())
		}
	}
	return output
}

func main() {
	if len(os.Args) != 2 {
		panic("USAGE: provide path to .vm file or directory")
	}
	path := os.Args[1]

	pathInfo, err := os.Stat(path)
	if err != nil {
		panic(err)
	}

	var filesToParse []string
	var ret, newpath string

	if pathInfo.IsDir() {
		dirFiles, err := ioutil.ReadDir(path)
		if err != nil {
			panic(err)
		}

		vmFiles := filterFiles(dirFiles, func(f os.FileInfo) bool {
			return strings.HasSuffix(f.Name(), ".vm")
		})

		for _, file := range vmFiles {
			filesToParse = append(filesToParse, filepath.Join(path, file))
		}

		fmt.Printf("Assembly: %s\n", codewriter.WriteBootstap())
		ret += codewriter.WriteBootstap()
		newpath = filepath.Join(path, fmt.Sprintf("%s.asm", pathInfo.Name()))
	} else {
		filesToParse = append(filesToParse, path)
		newpath = hackfile.NewPath(path, "vm", "asm")
	}

	for _, file := range filesToParse {
		ret += parseFile(file)
	}

	hackfile.WriteFile(newpath, ret)
}
