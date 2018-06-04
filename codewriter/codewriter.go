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
		return writeArithmetic(c.Arg1)
	case "C_GOTO":
		return writeGoto(c.Arg1)
	case "C_IF":
		return writeIf(c.Arg1)
	case "C_LABEL":
		return writeLabel(c.Arg1)
	case "C_POP":
		return writePop(c.Arg1, c.Arg2)
	case "C_PUSH":
		return writePush(c.Arg1, c.Arg2)
	case "C_FUNCTION":
		return writeFunction(c.Arg1, c.Arg2)
	default:
		return ""
	}
}

func writeArithmetic(operation string) string {
	var ret []string
	switch operation {
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
	ret = append(ret, incrementSP()...)
	return strings.Join(ret, "\n")
}

func writeFunction(functionName string, kLocals int) string {
	var lines []string
	lines = append(lines, label(functionName)...)
	for i := 0; i < kLocals; i++ {
		lines = append(lines, push("constant", 0)...)
	}
	return strings.Join(indentedLines(lines...), "\n")
}

func writeGoto(label string) string {
	return strings.Join(indentedLines(comment(fmt.Sprintf("Adding jump to %s", label)),
		fmt.Sprintf("@%s", label),
		"0;JMP"), "\n")
}

func writeIf(label string) string {
	var lines []string
	lines = append(lines, comment(fmt.Sprintf("Adding conditional jump to %s", label)))
	lines = append(lines, decrementSP()...)
	lines = append(lines, dereferenceSP()...)
	lines = append(lines, "D=M")
	lines = append(lines, fmt.Sprintf("@%s", label))
	lines = append(lines, "D;JNE")
	return strings.Join(indentedLines(lines...), "\n")
}

func writeLabel(str string) string {
	return strings.Join(label(str), "\n")
}

func label(label string) []string {
	return indentedLines(comment("Adding label"),
		fmt.Sprintf("(%s)", label))
}

// Pop into segment from stack
func writePop(segment string, index int) string {
	var lines []string
	lines = append(lines, comment(fmt.Sprintf("pop %s %d", segment, index)))
	lines = append(lines, dereferenceSegment(segment, index)...)
	lines = append(lines, "D=A") // get the address to come back to
	lines = append(lines, "@R13")
	lines = append(lines, "M=D") // store the value in memory
	lines = append(lines, decrementSP()...)
	lines = append(lines, dereferenceSP()...)
	lines = append(lines, "D=M") // get the value to pop
	lines = append(lines, "@R13")
	lines = append(lines, "A=M")
	lines = append(lines, "M=D")
	return strings.Join(indentedLines(lines...), "\n")
}

// Push onto stack from segment
func writePush(segment string, index int) string {
	return strings.Join(push(segment, index), "\n")
}

func push(segment string, index int) []string {
	var lines []string
	lines = append(lines, comment(fmt.Sprintf("push %s %d", segment, index)))
	lines = append(lines, loadSegment(segment, index)...)
	lines = append(lines, dereferenceSP()...)
	lines = append(lines, "M=D")
	lines = append(lines, incrementSP()...)
	return indentedLines(lines...)
}

// Helper Methods
func arithmeticAdd() []string {
	var lines []string
	lines = append(lines, comment("Adding x and y"))
	lines = append(lines, loadXY()...)
	lines = append(lines, "M=D+M")
	return indentedLines(lines...)
}

func arithmeticSub() []string {
	var lines []string
	lines = append(lines, comment("Subtracting y from x"))
	lines = append(lines, loadXY()...)
	lines = append(lines, "M=M-D")
	return indentedLines(lines...)
}

func arithmeticNeg() []string {
	var lines []string
	lines = append(lines, comment("Negating x"))
	lines = append(lines, decrementSP()...)
	lines = append(lines, dereferenceSP()...)
	lines = append(lines, "M=-M")
	return indentedLines(lines...)
}

func arithmeticComparison(comparison string) []string {
	var lines []string
	lines = append(lines, comment(fmt.Sprintf("Comparing x and y using %s", comparison)))
	lines = append(lines, loadXY()...)                         // D contains y, M contains x, A contains SP
	lines = append(lines, "D=M-D")                             // D contains x-y
	lines = append(lines, "M=-1")                              // save "true" to top of stack
	lines = append(lines, fmt.Sprintf("@DONE%d\n", doneCount)) // A contains address of code to jump to if condition satisfied; M no longer contains top of stack!
	lines = append(lines, fmt.Sprintf("D;%s\n", comparison))   // jump ahead in code if condition satisfied
	lines = append(lines, dereferenceSP()...)                  // reset A & M
	lines = append(lines, "M=0")                               // save "false" to top of stack
	lines = append(lines, fmt.Sprintf("(DONE%d)", doneCount))
	lines = append(lines, comment("done"))
	doneCount++
	return indentedLines(lines...)
}

func arithmeticEq() []string {
	return arithmeticComparison("JEQ")
}

func arithmeticGt() []string {
	return arithmeticComparison("JGT")
}

func arithmeticLt() []string {
	return arithmeticComparison("JLT")
}

func arithmeticAnd() []string {
	var lines []string
	lines = append(lines, comment("And-ing x and y"))
	lines = append(lines, loadXY()...)
	lines = append(lines, "M=D&M")
	return indentedLines(lines...)
}

func arithmeticOr() []string {
	var lines []string
	lines = append(lines, comment("Or-ing x and y"))
	lines = append(lines, loadXY()...)
	lines = append(lines, "M=D|M")
	return indentedLines(lines...)
}

func arithmeticNot() []string {
	var lines []string
	lines = append(lines, comment("Not-ing x"))
	lines = append(lines, decrementSP()...)
	lines = append(lines, dereferenceSP()...)
	lines = append(lines, "M=!M")
	return indentedLines(lines...)
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
func decrementSP() []string {
	return indentedLines(comment("Decrementing Stack Pointer"),
		"@SP",
		"M=M-1")
}

// Dereference Stack Pointer
func dereferenceSP() []string {
	return indentedLines(comment("Dereferencing Stack Pointer"),
		"@SP",
		"A=M")
}

// Dereference Segment
func dereferenceSegment(segment string, offset int) []string {
	switch segment {
	case "pointer":
		return dereferencePointer(offset)
	case "static":
		return dereferenceStatic(offset)
	case "temp":
		return dereferenceTemp(offset)
	default:
		return dereferenceDefault(segment, offset)
	}
}

// Dereference default
func dereferenceDefault(segment string, offset int) []string {
	return indentedLines(comment(fmt.Sprintf("Accessing value in %s %d", segment, offset)),
		fmt.Sprintf("@%s", standardizeSegment(segment)), // goto segment e.g., LOCAL
		"D=M", // find address that LOCAL points to; store in data register
		fmt.Sprintf("@%d", offset), // load offset
		"A=D+A")                    // add offset to base address, goto
}

// Dereference pointer
func dereferencePointer(offset int) []string {
	comment := comment(fmt.Sprintf("Accessing value in pointer %d", offset))
	switch offset {
	case 0:
		return indentedLines(comment, "@THIS")
	case 1:
		return indentedLines(comment, "@THAT")
	default:
		panic(fmt.Sprintf("Error: invalid offset for pointer: %d", offset)) // TODO: this function should return an error
	}
}

// Dereference static
func dereferenceStatic(offset int) []string {
	symbol := fmt.Sprintf("%s.%d", Filename, offset)
	return indentedLines(comment(fmt.Sprintf("Accessing value in static %d", offset)),
		fmt.Sprintf("@%s", symbol))
}

// Dereference temp
func dereferenceTemp(offset int) []string {
	addr := baseAddressForTemp + offset
	if addr < 5 || addr > 15 {
		panic(fmt.Sprintf("Error: invalid offset for temp: %d", offset))
	}
	return indentedLines(comment(fmt.Sprintf("Accessing value in temp %d", offset)),
		fmt.Sprintf("@%d", addr))
}

// Increment Stack Pointer
func incrementSP() []string {
	return indentedLines(comment("Incrementing Stack Pointer"),
		"@SP",
		"M=M+1")
}

// Load value from segment into data register
func loadSegment(segment string, index int) []string {
	var lines []string
	switch segment {
	case "constant":
		lines = append(lines, comment("Loading constant"))
		lines = append(lines, fmt.Sprintf("@%d", index))
		lines = append(lines, "D=A")
	default:
		lines = append(lines, comment("Loading segment"))
		lines = append(lines, dereferenceSegment(segment, index)...)
		lines = append(lines, "D=M")
	}
	return indentedLines(lines...)
}

// Load top of stack (y) into D register, dereference second-to-top of stack (x)
func loadXY() []string {
	var lines []string
	lines = append(lines, comment("Accessing top two values from stack"))
	lines = append(lines, decrementSP()...)
	lines = append(lines, dereferenceSP()...)
	lines = append(lines, "D=M")
	lines = append(lines, decrementSP()...)
	lines = append(lines, dereferenceSP()...)
	return indentedLines(lines...)
}

func comment(comment string) string {
	return fmt.Sprintf("// %s", comment)
}

func indentedLines(lines ...string) []string {
	var indentedLines []string
	comment := lines[0]
	indentedLines = append(indentedLines, comment)
	for _, line := range lines[1:] {
		indentedLines = append(indentedLines, "  "+line)
	}
	return indentedLines
}
