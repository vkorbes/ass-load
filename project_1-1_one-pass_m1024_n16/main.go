package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"unicode"
)

var opcodes = []struct {
	code       int
	mnemonic   string
	hasOperand bool
}{
	{1, "LOD", true},
	{2, "STO", true},
	{3, "ADD", true},
	{4, "BZE", true},
	{5, "BNE", true},
	{6, "BRA", true},
	{7, "INP", false},
	{8, "OUT", false},
	{9, "CLA", false},
	{0, "HLT", false},
}

type symbol struct {
	name   rune
	value  int
	status rune
}

var symbolTable = []symbol{}

// M = 1024 words
// N = 16 bits
var memory [1024]uint16 // uint16: decimal 0–65535
var locationCounter int

func main() {
	if len(os.Args) > 3 {
		log.Fatal("Program requires at least one argument, indicating the source file.")
	}
	file, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		parse(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("\nSymbol table:")
	for i := range symbolTable {
		fmt.Println(string(symbolTable[i].name), "\t", symbolTable[i].value, "\t", string(symbolTable[i].status))
	}
	fmt.Println("\nObject:")
	for i := range memory {
		if memory[i] != 0 {
			fmt.Printf("%02d - %06b %010b\n", i, wordOpcode(memory[i]), wordOperand(memory[i]))
		}
	}
}

/*
   Valid lines can have the following shapes:
   L OPC OPRD
   L OPC
   OPC OPRD
   OPC
*/
func parse(s string) {
	s = strings.Trim(s, " \t")
	splat := strings.Split(s, " ")
	ok := false
	if len(splat) == 3 && len(splat[0]) == 1 {
		defineSymbol(splat[0])
		assembleLine(splat[1], splat[2])
		ok = true
	}
	if len(splat) == 2 && len(splat[0]) == 1 {
		defineSymbol(splat[0])
		assembleLine(splat[1], "")
		ok = true
	}
	if len(splat) == 2 && len(splat[0]) == 3 {
		assembleLine(splat[0], splat[1])
		ok = true
	}
	if len(splat) == 1 && len(splat[0]) == 3 {
		assembleLine(splat[0], "")
		ok = true
	}
	if !ok {
		log.Fatal(fmt.Sprintln("Error parsing line:", s))
	}
	x := memory[locationCounter-1]
	fmt.Printf("%02d - %06b %010b\t%s\n", locationCounter-1, wordOpcode(x), wordOperand(x), splat)
}

func defineSymbol(s string) {
	r := rune(s[0])
	if !unicode.IsLetter(r) {
		log.Fatal(fmt.Sprintln("Symbol must be a letter! Invalid:", s))
	}
	for i := range symbolTable {
		if symbolTable[i].name == r {
			if symbolTable[i].status == 'D' {
				log.Fatal(fmt.Sprintln("Label is doubly defined! Invalid:", s))
			}
			v := symbolTable[i].value
			symbolTable[i].status = 'D'
			symbolTable[i].value = locationCounter
			for v != 0 {
				next := int(wordOperand(memory[v]))
				memory[v] = (memory[v] & 0000000000) | uint16(locationCounter)
				v = next
			}
			return
		}
	}
	symbolTable = append(symbolTable, symbol{r, locationCounter, 'D'})
}

func lookupSymbol(r rune) int {
	for i := range symbolTable {
		if symbolTable[i].name == r {
			if symbolTable[i].status == 'D' {
				return symbolTable[i].value
			} else {
				prev := symbolTable[i].value
				symbolTable[i].value = locationCounter
				return prev
			}
		}
	}
	symbolTable = append(symbolTable, symbol{r, locationCounter, 'U'})
	return 0
}

func assembleLine(opc, operand string) {
	var index int
	for i := range opcodes {
		if opcodes[i].mnemonic == opc {
			index = i
			if opcodes[i].hasOperand == false && operand != "" {
				log.Fatal(fmt.Sprintln("OpCode ", opc, "should not have an operand!"))
			}
		}
	}
	// opcode with no operand
	if !opcodes[index].hasOperand {
		memory[locationCounter] = uint16(opcodes[index].code << 10)
		locationCounter++
		return
	}
	// opcode with symbol operand
	r := rune(operand[0])
	if unicode.IsLetter(r) {
		memory[locationCounter] = uint16(opcodes[index].code<<10 | lookupSymbol(r))
		locationCounter++
		return
	}
	// opcode with numerical operand
	intOperand, err := strconv.Atoi(operand)
	if err != nil {
		log.Fatal(fmt.Sprintln("Error converting operand", operand, "to integer."))
	}
	if intOperand > 0b1000000000 {
		log.Fatal(fmt.Sprintf("Operand %d exceeds max value of %d.", intOperand, 0b1000000000))
	}
	memory[locationCounter] = uint16((opcodes[index].code << 10) | intOperand)
	locationCounter++
}

func wordOpcode(w uint16) uint16 {
	return w >> 10
}

func wordOperand(w uint16) uint16 {
	return w & 0b0000001111111111
}
