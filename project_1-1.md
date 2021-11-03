## Project 1-1

Implement the following description as both a one-pass and a two-pass assembler.

- It has a single register: Acc
- It has M words of storage
- Each word is N-bits long
- Support for 1-letter symbols/labels, incl. future symbols
- No directives
- Possible errors: invalid mnemonic, invalid label, multiply-defined label, invalid operand (both syntactically and > max. address)

Assume the program can't keep the source code in memory. It reads it line by line either once, or twice, and performs the right operations as that line is read.

One-pass should output the object program in memory, and a listing file.

Two-pass should output an object file and a listing file.

The listing file should contain:
- Location counter
- Label
- Source
- Object code

The instruction set is:

```
Opcode Mnemonic Operand Description
1      LOD      yes     acc ← mem(op)
2      STO      yes     mem(op) ← acc
3      ADD      yes     acc ← acc+mem(op)
4      BZE      yes     go to op if acc=0
5      BNE      yes     go to op if acc<0
6      BRA      yes     unconditional branch
7      INP      no      acc ← next char in input stream
8      OUT      no      next char in output stream ← acc
9      CLA      no      acc ← 0
0      HLT      no      stop
```

Run the following test program (numbers are decimal):

```
Label   Source      Object
        INP         7
        STO 50      2 50
        INP         7
        STO 51      2 51
        BZE X       4 X
        ADD 50      3 50
        OUT         8
        BRA Y       6 Y
    X   LOD 50      1 50
        ADD 50      3 50
    Y   STO 52      2 52
        HLT         0
```

1. For M=1024 and N=16, with 6-bit opcodes and 10-bit operands, the output should be:

```
                    ----- Object -----
LC  Label  Source   Opcode  Op
0          INP      000111  0
1          STO 50   000010  0000110010
2          INP      000111  0
3          STO 51   000010  0000110011
4          BZE X    000100  0000001000
5          ADD 50   000011  0000110010
6          OUT      001000  0
7          BRA Y    000110  0000001010
8       X  LOD 50   000001  0000110010
9          ADD 50   000011  0000110010
10      Y  STO 52   000010  0000110100
11         HLT      000000  0
```

2. Now use M=256 and N=8. M=256 means 8-bit addresses; so opcodes should be one word, and operands another word. Thus, instructions may be one or two words long. The output for the test program should be:

```
                    ----- Object -----
LC  Label  Source   Opcode    Op
0          INP      00000111  
1          STO 50   00000010  00110010
3          INP      00000111  
4          STO 51   00000010  00110011
6          BZE X    00000100  00001101
8          ADD 50   00000011  00110010
10         OUT      00001000  
11         BRA Y    00000110  00010001
13      X  LOD 50   00000001  00110010
15         ADD 50   00000011  00110010
17      Y  STO 52   00000010  00110100
19         HLT      00000000  
```

3. Now there's room for 256 opcodes. Too many. Take 4 bits from the opcode word and use 2 to add 4 different addressing modes (direct, relative, indirect, immediate); and the other 2 to add 4 extra general-purpose registers.

4. Write tests for everything.