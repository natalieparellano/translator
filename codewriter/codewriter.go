package codewriter

import (
	"fmt"
	"strings"

	"github.com/natalieparellano/translator/parser"
)

var doneCount = 0
var baseAddressForTemp = 5
var Filename string

func Write(c parser.Command) string {
	switch c.Type {
	case "C_ARITHMETIC":
		return writeArithmetic(c)
	case "C_GOTO":
		return writeGoto(c)
	case "C_IF":
		return writeIf(c)
	case "C_LABEL":
		return writeLabel(c)
	case "C_POP":
		return writePop(c.Arg1, c.Arg2)
	case "C_PUSH":
		return writePush(c.Arg1, c.Arg2)
	default:
		return ""
	}
}

func writeArithmetic(c parser.Command) string {
	if c.Type != "C_ARITHMETIC" {
		panic(fmt.Sprintf("Error: called writeArithmetic with invalid command: type %s", c.Type))
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

func writeGoto(c parser.Command) string {
	return fmt.Sprintf("// Adding jump to %s\n", c.Arg1) +
		fmt.Sprintf("@%s\n", c.Arg1) +
		"0;JMP\n" +
		"// done\n"
}

func writeIf(c parser.Command) string {
	return fmt.Sprintf("// Adding conditional jump to %s\n", c.Arg1) +
		decrementSP() +
		dereferenceSP() +
		"D=M\n" +
		fmt.Sprintf("@%s\n", c.Arg1) +
		"D;JNE\n" +
		"// done\n"
}

func writeLabel(c parser.Command) string {
	return "// Adding label\n" +
		fmt.Sprintf("(%s)\n", c.Arg1) +
		"// done\n"
}

// Pop into segment from stack
func writePop(segment string, index int) string {
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
func writePush(segment string, index int) string {
	return fmt.Sprintf("// push %s %d\n", segment, index) +
		loadSegment(segment, index) +
		dereferenceSP() +
		"M=D\n" +
		incrementSP() +
		"// done\n"
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
		loadXY() + // D contains y, M contains x, A contains SP
		"D=M-D\n" + // D contains x-y
		"M=-1\n" + // save "true" to top of stack
		fmt.Sprintf("@DONE%d\n", doneCount) + // A contains address of code to jump to if condition satisfied; M no longer contains top of stack!
		fmt.Sprintf("D;%s\n", comparison) + // jump ahead in code if condition satisfied
		dereferenceSP() + // reset A & M
		"M=0\n" + // save "false" to top of stack
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
func standardizeSegment(segment string) string {
	segMap := map[string]string{
		"LOCAL":    "LCL",
		"ARGUMENT": "ARG",
		"THIS":     "THIS",
		"THAT":     "THAT",
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
	return fmt.Sprintf("@%s\n", standardizeSegment(segment)) + // goto segment e.g., LOCAL
		"D=M\n" + // find address that LOCAL points to; store in data register
		fmt.Sprintf("@%d\n", offset) + // load offset
		"A=D+A\n" // add offset to base address, goto
}

// Dereference pointer
func dereferencePointer(offset int) string {
	switch offset {
	case 0:
		return "@THIS\n"
	case 1:
		return "@THAT\n"
	default:
		panic(fmt.Sprintf("Error: invalid offset for pointer: %d", offset)) // TODO: this function should return an error
	}
}

// Dereference static
func dereferenceStatic(offset int) string {
	symbol := fmt.Sprintf("%s.%d", Filename, offset)
	return fmt.Sprintf("@%s\n", symbol)
}

// Dereference temp
func dereferenceTemp(offset int) string {
	addr := baseAddressForTemp + offset
	if addr < 5 || addr > 15 {
		panic(fmt.Sprintf("Error: invalid offset for temp: %d", offset))
	}
	return fmt.Sprintf("@%d\n", addr)
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
