package codewriter

import (
	"fmt"
	"strings"

	"github.com/natalieparellano/translator/parser"
)

var doneCount = 0

func WriteArithmetic(c parser.Command) string {
	if c.Type != "C_ARITHMETIC" {
		panic(fmt.Sprintf("Error: called WriteArithmetic with invalid command: type %s", c.Type))
	}
	var ret string
	switch c.Arg1 {
	case "add":
		ret = arithmeticAdd()
	case "sub":
		ret = arithmeticSub()
	case "neg":
		ret = arithmeticNeg()
	case "eq":
		ret = arithmeticEq()
	case "gt":
		ret = arithmeticGt()
	case "lt":
		ret = arithmeticLt()
	case "and":
		ret = arithmeticAnd()
	case "or":
		ret = arithmeticOr()
	case "not":
		ret = arithmeticNot()
	}
	return ret + incrementSP()
}

func WritePushPop(c parser.Command) string {
	if c.Type == "C_PUSH" {
		return push(c.Arg1, c.Arg2) +
			incrementSP()
	}
	return pop(c.Arg1, c.Arg2)
}

// Helper Methods
func arithmeticAdd() string {
	return "// Adding x and y\n" +
		loadXY() +
		"M=D+M\n" +
		"// done\n"
}

func arithmeticSub() string {
	return "// Subtracting y from x\n" +
		loadXY() +
		"M=M-D\n" +
		"// done\n"
}

func arithmeticNeg() string {
	return "// Negating x\n" +
		decrementSP() +
		dereferenceSP() +
		"M=-M\n" +
		"// done\n"
}

func arithmeticComparison(comparison string) string {
	ret := fmt.Sprintf("// Comparing x and y using %s\n", comparison) +
		loadXY() +
		"D=M-D\n" +
		dereferenceSP() +
		"M=-1\n" +
		fmt.Sprintf("@DONE%d\n", doneCount) +
		fmt.Sprintf("D;%s\n", comparison) +
		dereferenceSP() +
		"M=0\n" +
		fmt.Sprintf("(DONE%d)\n", doneCount) +
		"// done\n"
	doneCount++
	return ret
}

func arithmeticEq() string {
	return arithmeticComparison("JEQ")
}

func arithmeticGt() string {
	return arithmeticComparison("JGT")
}

func arithmeticLt() string {
	return arithmeticComparison("JLT")
}

func arithmeticAnd() string {
	return "// And-ing x and y\n" +
		loadXY() +
		"M=D&M\n" +
		"// done\n"
}

func arithmeticOr() string {
	return "// Or-ing x and y\n" +
		loadXY() +
		"M=D|M\n" +
		"// done\n"
}

func arithmeticNot() string {
	return "// Not-ing x\n" +
		decrementSP() +
		dereferenceSP() +
		"M=!M\n" +
		"// done\n"
}

// Return base address for segment
func segPointer(segment string) int {
	segMap := map[string]int{
		"SP":       0,
		"LOCAL":    1,
		"ARGUMENT": 2,
		"THIS":     3,
		"THAT":     4,
		"TEMP":     5,
	}
	return segMap[strings.ToUpper(segment)]
}

// Decrement Stack Pointer
func decrementSP() string {
	return "// Decrementing Stack Pointer\n" +
		"@SP\n" +
		"M=M-1\n" +
		"// done\n"
}

// Dereference Stack Pointer
func dereferenceSP() string {
	return "// Dereferencing Stack Pointer\n" +
		"@SP\n" +
		"A=M\n" +
		"// done\n"
}

// Dereference Segment
func dereferenceSegment(segment string, offset int) string {
	comment := fmt.Sprintf("// Accessing value in %s %d\n", segment, offset)
	var ret string
	switch segment {
	case "pointer":
		ret = dereferencePointer(offset)
	case "static":
		ret = dereferenceStatic(offset)
	case "temp":
		ret = dereferenceTemp(offset)
	default:
		ret = dereferenceDefault(segment, offset)
	}
	return comment + ret
}

// Dereference default
func dereferenceDefault(segment string, offset int) string {
	return fmt.Sprintf("@%d\n", segPointer(segment)) + // goto segment e.g., LOCAL
		"D=M\n" + // find address that LOCAL points to; store in data register
		fmt.Sprintf("@%d\n", offset) + // load offset
		"A=D+A\n" + // add offset to base address, goto
		"// done\n"
}

// Dereference pointer
func dereferencePointer(offset int) string {
	switch offset {
	case 0:
		return fmt.Sprintf("@%d\n", segPointer("this"))
	case 1:
		return fmt.Sprintf("@%d\n", segPointer("that"))
	default:
		panic(fmt.Sprintf("Error: invalid offset for pointer: %s", offset)) // TODO: this function should return an error
	}
}

// Dereference static
func dereferenceStatic(offset int) string {
	panic(fmt.Sprintf("Error: not implemented"))
}

// Dereference temp
func dereferenceTemp(offset int) string {
	addr := segPointer("temp") + offset
	if addr < 5 || addr > 15 {
		panic(fmt.Sprintf("Error: invalid offset for temp: %d", offset))
	}
	return fmt.Sprintf("@%d\n", addr) +
		"// done\n"
}

// Increment Stack Pointer
func incrementSP() string {
	return "// Incrementing Stack Pointer\n" +
		"@SP\n" +
		"M=M+1\n" +
		"// done\n"
}

// Load value from segment into data register
func loadSegment(segment string, index int) string {
	switch segment {
	case "constant":
		return "// Loading constant\n" +
			fmt.Sprintf("@%d\n", index) +
			"D=A\n" +
			"// done\n"
	default:
		return "// Loading segment\n" +
			dereferenceSegment(segment, index) +
			"D=M\n" +
			"// done\n"
	}
}

// Load top of stack (y) into D register, dereference second-to-top of stack (x)
func loadXY() string {
	return "// Accessing top two values from stack\n" +
		decrementSP() +
		dereferenceSP() +
		"D=M\n" +
		decrementSP() +
		dereferenceSP() +
		"// done\n"
}

// Pop into segment from stack
func pop(segment string, index int) string {
	return fmt.Sprintf("// pop %s %d\n", segment, index) +
		dereferenceSegment(segment, index) +
		"D=A\n" + // get the address to come back to
		"@R13\n" +
		"M=D\n" + // store the value in memory
		decrementSP() +
		dereferenceSP() +
		"D=M\n" + // get the value to pop
		"@R13\n" +
		"A=M\n" +
		"M=D\n" +
		"// done\n"
}

// Push onto stack from segment
func push(segment string, index int) string {
	return fmt.Sprintf("// push %s %d\n", segment, index) +
		loadSegment(segment, index) +
		dereferenceSP() +
		"M=D\n" +
		"// done\n"
}
