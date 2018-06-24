package codewriter

import (
	"fmt"
	"strings"

	"github.com/natalieparellano/translator/parser"
)

var doneCount = 0
var returnCount = 0
var baseAddressForTemp = 5
var Filename string

func Write(c parser.Command) string {
	var lines []string
	switch c.Type {
	case "C_ARITHMETIC":
		lines = doArithmetic(c.Arg1)
	case "C_GOTO":
		lines = doGoto(c.Arg1)
	case "C_IF":
		lines = doIf(c.Arg1)
	case "C_LABEL":
		lines = label(c.Arg1)
	case "C_POP":
		lines = pop(c.Arg1, c.Arg2)
	case "C_PUSH":
		lines = push(c.Arg1, c.Arg2)
	case "C_FUNCTION":
		lines = declareFunction(c.Arg1, c.Arg2)
	case "C_RETURN":
		lines = returnFunction()
	case "C_CALL":
		lines = callFunction(c.Arg1, c.Arg2)
	}
	return strings.Join(lines, "\n")
}

func WriteBootstap() string {
	var lines []string
	lines = append(lines, indentedLines(comment("Setting SP to 256"),
		"@256",
		"D=A",
		"@SP",
		"M=D")...)
	lines = append(lines, doGoto("Sys.init")...)
	return strings.Join(lines, "\n")
}

func callFunction(name string, kArgs int) []string {
	var lines []string
	lines = append(lines, comment(fmt.Sprintf("Calling function %s with %d args", name, kArgs)))
	lines = append(lines, pushRegister(fmt.Sprintf("RETURN%d", returnCount))...)
	lines = append(lines, pushRegister("LCL")...)
	lines = append(lines, pushRegister("ARG")...)
	lines = append(lines, pushRegister("THIS")...)
	lines = append(lines, pushRegister("THAT")...)
	lines = append(lines, "@SP")
	lines = append(lines, "D=M")
	lines = append(lines, fmt.Sprintf("@%d", kArgs))
	lines = append(lines, "D=D-A")
	lines = append(lines, "@5")
	lines = append(lines, "D=D-A")
	lines = append(lines, "@ARG")
	lines = append(lines, "M=D")
	lines = append(lines, "@SP")
	lines = append(lines, "D=M")
	lines = append(lines, "@LCL")
	lines = append(lines, "M=D")
	lines = append(lines, doGoto(name)...)
	lines = append(lines, label(fmt.Sprintf("RETURN%d", returnCount))...)
	returnCount++
	return (indentedLines(lines...))
}

func doArithmetic(operation string) []string {
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
	return append(ret, incrementSP()...)
}

func declareFunction(functionName string, kLocals int) []string {
	var lines []string
	lines = append(lines, comment(fmt.Sprintf("Declaring function %s", functionName)))
	lines = append(lines, label(functionName)...)
	for i := 0; i < kLocals; i++ {
		lines = append(lines, push("constant", 0)...)
	}
	return indentedLines(lines...)
}

func doGoto(label string) []string {
	return indentedLines(comment(fmt.Sprintf("Adding jump to %s", label)),
		fmt.Sprintf("@%s", label),
		"0;JMP")
}

func doIf(label string) []string {
	var lines []string
	lines = append(lines, comment(fmt.Sprintf("Adding conditional jump to %s", label)))
	lines = append(lines, decrementSP()...)
	lines = append(lines, dereferenceSP()...)
	lines = append(lines, "D=M")
	lines = append(lines, fmt.Sprintf("@%s", label))
	lines = append(lines, "D;JNE")
	return indentedLines(lines...)
}

func label(label string) []string {
	return indentedLines(comment("Adding label"),
		fmt.Sprintf("(%s)", label))
}

func pop(segment string, index int) []string {
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
	return indentedLines(lines...)
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

func pushRegister(address string) []string {
	var lines []string
	lines = append(lines, comment(fmt.Sprintf("push %s", address)))
	lines = append(lines, fmt.Sprintf("@%s", address))
	lines = append(lines, "D=M")
	lines = append(lines, dereferenceSP()...)
	lines = append(lines, "M=D")
	lines = append(lines, incrementSP()...)
	return indentedLines(lines...)
}

func returnFunction() []string {
	var lines []string
	lines = append(lines, comment("Returning from function"))
	lines = append(lines, fmt.Sprintf("@%s", standardizeSegment("LOCAL")))
	lines = append(lines, "D=M") // FRAME
	lines = append(lines, "@R14")
	lines = append(lines, "M=D") // store FRAME in memory
	lines = append(lines, "@5")
	lines = append(lines, "A=D-A") // *(FRAME - 5)
	lines = append(lines, "D=M")   // RET
	lines = append(lines, "@R15")
	lines = append(lines, "M=D") // store RET in memory
	lines = append(lines, pop("ARGUMENT", 0)...)
	lines = append(lines, dereferenceDefault(standardizeSegment("ARGUMENT"), 1, "+")...)
	lines = append(lines, "D=A") // WARN: inefficient, previous instruction doesn't need to "goto" ARG+1
	lines = append(lines, "@SP")
	lines = append(lines, "M=D")                                // SP = ARG + 1
	lines = append(lines, dereferenceDefault("R14", 1, "-")...) // *(FRAME - 1)
	lines = append(lines, "D=M")
	lines = append(lines, "@THAT")
	lines = append(lines, "M=D")
	lines = append(lines, dereferenceDefault("R14", 2, "-")...) // *(FRAME - 2)
	lines = append(lines, "D=M")
	lines = append(lines, "@THIS")
	lines = append(lines, "M=D")
	lines = append(lines, dereferenceDefault("R14", 3, "-")...) // *(FRAME - 3)
	lines = append(lines, "D=M")
	lines = append(lines, "@ARG")
	lines = append(lines, "M=D")
	lines = append(lines, dereferenceDefault("R14", 4, "-")...) // *(FRAME - 4)
	lines = append(lines, "D=M")
	lines = append(lines, "@LCL")
	lines = append(lines, "M=D")
	lines = append(lines, "@R15")
	lines = append(lines, "A=M")
	lines = append(lines, "0;JMP")

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
		return dereferenceDefault(standardizeSegment(segment), offset, "+")
	}
}

// Dereference default
func dereferenceDefault(segment string, offset int, operation string) []string {
	switch operation {
	case "+", "-":
		// nop
	default:
		panic(fmt.Sprintf("dereferenceDefault called with invalid operation: %s", operation))
	}
	return indentedLines(comment(fmt.Sprintf("Accessing value in %s %d", segment, offset)),
		fmt.Sprintf("@%s", segment), // goto segment e.g., LOCAL
		"D=M", // find address that LOCAL points to; store in data register
		fmt.Sprintf("@%d", offset),       // load offset
		fmt.Sprintf("A=D%sA", operation)) // add offset to base address, goto
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
