package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"unicode"
)

var opcodes = []struct {
	code       int
	mnemonic   string
	hasOperand bool
}{
	{0, "HLT", false},
	{1, "LOD", true},
	{2, "STO", true},
	{3, "ADD", true},
	{4, "BZE", true},
	{5, "BNE", true},
	{6, "BRA", true},
	{7, "INP", false},
	{8, "OUT", false},
	{9, "CLA", false},
}

type symbol struct {
	name   rune
	value  int
	status rune
}

type printLine struct {
	locationCounter int
	label           string
	source          string
	opcode          uint8
	operand         uint8
}

var symbolTable = []symbol{}

// M = 256 words
// N = 8 bits
var memory [256]uint8 // uint8: decimal 0–255
var locationCounter int
var line printLine

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Program requires at least one argument, indicating the source file.")
	}
	file, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	fmt.Println("First Pass (labels unresolved):")
	w := tabwriter.NewWriter(os.Stdout, 2, 0, 3, ' ', 0)
	fmt.Fprintln(w, fmt.Sprint("LC\tLabel\tSource\tOpcode\tOperand\t"))

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line = printLine{locationCounter, "", "", 0, 0}
		parseLine(scanner.Text())

		// printing shenanigans
		temp := fmt.Sprintf("%d\t%s\t%s\t%08b\t", line.locationCounter, line.label, line.source, line.opcode)
		if line.operand != 0 {
			temp = fmt.Sprintf("%s%08b\t", temp, line.operand)
		}
		fmt.Fprintln(w, temp)
	}

	w.Flush()
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("\nSymbol table:")
	for i := range symbolTable {
		fmt.Println(string(symbolTable[i].name), "\t", symbolTable[i].value, "\t", string(symbolTable[i].status))
	}

	fmt.Println("\nAssembled Object:")
	printMem()
}

/*
   Valid lines can have the following shapes:
   L OPC OPRD
   L OPC
   OPC OPRD
   OPC
*/
func parseLine(s string) {
	s = strings.Trim(s, " \t")
	splat := strings.Split(s, " ")
	ok := false
	if len(splat) == 3 && len(splat[0]) == 1 {
		defineSymbol(splat[0])
		assembleLine(splat[1], splat[2])
		line.source = fmt.Sprint(splat[1], " ", splat[2])
		ok = true
	}
	if line.source == "" {
		line.source = s
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
}

func defineSymbol(s string) {
	r := rune(s[0])
	if !unicode.IsLetter(r) {
		log.Fatal(fmt.Sprintln("Symbol must be a letter! Invalid:", s))
	}
	line.label = s
	for i := range symbolTable {
		if symbolTable[i].name == r {
			if symbolTable[i].status == 'D' {
				log.Fatal(fmt.Sprintln("Label is doubly defined! Invalid:", s))
			}
			v := symbolTable[i].value
			symbolTable[i].status = 'D'
			symbolTable[i].value = locationCounter
			for v != 0 {
				next := int(memory[v+1])
				memory[v+1] = uint8(locationCounter)
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
		memory[locationCounter] = uint8(opcodes[index].code)
		line.opcode = uint8(opcodes[index].code)
		locationCounter++
		return
	}
	// opcode with symbol operand
	r := rune(operand[0])
	if unicode.IsLetter(r) {
		memory[locationCounter] = uint8(opcodes[index].code)
		memory[locationCounter+1] = uint8(lookupSymbol(r))
		line.opcode = uint8(opcodes[index].code)
		line.operand = memory[locationCounter+1]
		locationCounter += 2
		return
	}
	// opcode with numerical operand
	intOperand, err := strconv.Atoi(operand)
	if err != nil {
		log.Fatal(fmt.Sprintln("Error converting operand", operand, "to integer."))
	}
	if intOperand >= 0b100000000 {
		log.Fatal(fmt.Sprintf("Operand %d exceeds max value of %d.", intOperand, 0b100000000-1))
	}
	memory[locationCounter] = uint8(opcodes[index].code)
	memory[locationCounter+1] = uint8(intOperand)
	line.opcode = uint8(opcodes[index].code)
	line.operand = uint8(intOperand)
	locationCounter += 2
}

type memInstr struct {
	lc     int
	opcBin uint8
	oprBin uint8
	opcStr string
	oprStr string
}

func printMem() {
	var output memInstr

	w := tabwriter.NewWriter(os.Stdout, 2, 0, 3, ' ', 0)
	fmt.Fprintln(w, fmt.Sprint("LC\tObject\t\tOpcode\tOperand\t"))

	for {
		output.oprBin = 0
		output.opcBin = memory[output.lc]
		if opcodes[output.opcBin].hasOperand {
			output.oprBin = memory[output.lc+1]
			fmt.Fprintln(w, fmt.Sprintf("%d\t%08b\t%08b\t%s\t%d\t", output.lc, output.opcBin, output.oprBin, opcodes[output.opcBin].mnemonic, output.oprBin))
			output.lc++
		} else {
			fmt.Fprintln(w, fmt.Sprintf("%d\t%08b\t\t%s\t\t\t", output.lc, output.opcBin, opcodes[output.opcBin].mnemonic))
		}
		output.lc++

		if memory[output.lc] == 0 {
			output.opcBin = memory[output.lc]
			fmt.Fprintln(w, fmt.Sprintf("%d\t%08b\t\t%s\t\t\t", output.lc, output.opcBin, opcodes[output.opcBin].mnemonic))
			break
		}
	}
	w.Flush()
}
