package parser

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type Command struct {
	Type string
	Arg1 string
	Arg2 int
}

func Parse(filepath string) []Command {
	commands := make([]Command, 1)

	f, err := os.Open(filepath)
	check(err)

	scanner := bufio.NewScanner(f)
	fmt.Printf("\nParsing %s\n", filepath)

	var line string
	var command Command
	count := 0

	for scanner.Scan() {
		fmt.Printf("Parsing: %s\n", scanner.Text())
		line = trimLine(scanner.Text())
		if isWhitespace(line) {
			continue
		}
		command = parseCommand(line)
		commands = append(commands, command)
		fmt.Printf("Command: type: %s, arg1: %s, arg2: %d\n",
			command.Type, command.Arg1, command.Arg2)
		count += 1
	}
	fmt.Printf("Found %d lines of code\n", count)

	return commands
}

// Helper Methods

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func arg1(line string) string {
	lineType := commandType(line)
	switch lineType {
	case "C_ARITHMETIC":
		return line
	case "C_RETURN":
		return ""
	default:
		return strings.Split(line, " ")[1]
	}
	return ""
}

func arg2(line string) int {
	lineType := commandType(line)

	switch lineType {
	case "C_PUSH", "C_POP", "C_FUNCTION", "C_CALL":
		ret, err := strconv.Atoi(strings.Split(line, " ")[2])
		check(err)
		return ret
	}
	return 0
}

func commandType(line string) string {
	lineMap := map[string]string{
		"add|sub|neg|eq|gt|lt|and|or|not": "C_ARITHMETIC",
		"push":     "C_PUSH",
		"pop":      "C_POP",
		"label":    "C_LABEL",
		"^goto":    "C_GOTO",
		"if-goto":  "C_IF",
		"function": "C_FUNCTION",
		"return":   "C_RETURN",
		"call":     "C_CALL",
	}

	for expr, ret := range lineMap {
		re := regexp.MustCompile(expr)
		if res := re.FindStringSubmatch(line); res != nil {
			return ret
		}
	}

	return ""
}

func parseCommand(line string) Command {
	var c Command
	c.Type = commandType(line)
	c.Arg1 = arg1(line)
	c.Arg2 = arg2(line)
	return c
}

// Duplicated code

func isWhitespace(line string) bool {
	return len(strings.TrimSpace(line)) == 0
}

func trimLine(line string) string {
	line = strings.Split(line, "//")[0]
	line = strings.TrimSpace(line)
	return line
}
