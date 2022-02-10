# Memory Machine

What if instead of an anemic 6502 register machine, I make one with no registers but that has a rich set of operators against memory?  All arithmetic and comparison instructions can be done on memory.  The program counter and stack pointer are located in memory.

opcode op1 [,op2]

op1 can be:
    constant
    memory
    *memory
op2 can be:
    constant
    memory
    *memory
    
addw ox, -24 ' ox = ox - 24

## Opcodes

Arithmetic:

* add: a += b
* sub: a -= b 
* mul: a *= b
* div: a /= b
* cmp: a - b

Logical:
* and: a &= b
* or:  a |= b
* xor: a ^= b

Assignment:
* cpy: a = b
* psh: [sp++] = a
* pop: a = [--sp]

Flow Control:
* jmp
* jeq
* jne
* jge
* jlt

# Addressing Modes

* Immediate: the value is copied directly

Ex:
    mov a, #27

* Absolute: the target is a memory address

    mov a, b

* Indirect: the target is pointed to by a memory address

    mov *a, *b

3 modes, 20+ instructions, possibly 2 widths (byte vs word).

none
imm
abs/abs
abs/imm
abs/ind
ind/abs
ind/imm
ind/ind

I can eliminate a few instructions by using memory to expose cpu "registers".  Ie, program counter, carry status, stack pointer, would all map to addresses.

That reduces the instruction set to 16, leaving us one extra bit for byte/word size.

Encoding:

    MMMSIIII
        M = Addressing mode
        S = Word size
        I = Instruction opcode

# Assembler Directives

## Symbols & Local Symbols

Symbols are declared and equate to the current offset, unless defined by an equate.

symbol

Local symbols are prefixed with '.' and are scoped to the previous symbol.

.l1
.l2

## Define space

    .ds <count>[, pattern]
    .db number[, number] | string
    .dw word
    .eq

## Constants

symbol = expression

## Expressions

* Numeric
* String
* Bytes

Example:

```asm
    DrawBlock
        var bloc_index ' sp + #0
        var temp       ' sp + #4

        cpy mult, #1
        cpy ox, x
        cpy oy, y
        cpy i, #1
        
    loop:
        cmp i, #14
        bgt done

        mov bloc_index, temp_piece
        mul bloc_index, #4
        add bloc_index, temp_rot
        and bloc + bloc_index, mult
        bne end

        psh ox
        psh oy
        mov temp, ox
        add temp, #6
        psh temp
        mov temp, oy
        add temp, #6
        psh temp
        psh #1
        mov temp, temp_piece
        add temp, #1
        psh temp
        call Display.DrawRect

    notit:
        cmp i, #4
        beq gotit
        cmp i, #8
        beq gotit
        cmp i, #12
        bne else
    gotit:
        add ox, #-24
        add oy, #8
        bra endif
    else:
        add ox, #8
    endif:

        mul mult, 2
        add i, #1
        bra loop
    done:
        ret
```

This might be fun to envision as an Apple 2 type of machine?

* 16 bit address space
* Instructions can operate on bytes or words
* DMA hardware for display, file i/o, network, etc

It has a simple instruction set and addressing modes, but with all the operands being 16 bit addresses ... the instructions will require 1 byte for the opcode, and up to 4 bytes for operands.  Zero page on 6502 was precious just because you could save a byte. Meh, what the hell.

Do I make a ROM that bootstraps your program?  Nah.  I think the program specifies its own load address.  The assembler creates an object file that represents the 64kb of memory at startup, like a ROM ... and that's what it runs.  Maybe you need to include something at org $fff0 that is a jump to your actual entry point.

label: \n    - Associates label to current PC
.label \n (local label associated to most recent global label)

label: word arg[,arg]*
word arg[,arg]*

ident = expr
ident:

if we parse a '.', its a local label
if we get ident, peek to see if there's a '=' or ':'
    a ':' defines a global label
    a '=' defines an equate
    if neither in peekahead, it must be a opcode or directive

line :=
        label
    |   label statement
    |   statement

label := 
        ident 
    |   .ident

statement :=
        ident arg[,arg]*

Operand  :=
      '#' expr
    | '*' expr
    | expr

expr :=
    MulExpr [ ('+' | '-' | '|' | '^') MulExpr]*

MulExpr :=
    UnaryExpr ['*' | '/' | '%' | '<<' | '>>'  UnaryExpr]*

UnaryExpr :=
    ['+' | '-'] PrimaryExpr

PrimaryExpr :=
      '(' expr ')'
    | Identifier
    | Literal (int, String, Char)

expressions have a type which ends up being basically integer or []byte.

I need to build the AST tree for expressions, how to model that.

The type can be: literal (string, int), symbol, unary (+/-) expr, or binary op expr.

type Operand struct {
    mode (immediate, indirect, absolute)
    expr
}

type Expr interface {
    computeType() -> integer, []byte
}

type ExprLiteral struct {

}

type ExprSymbol struct {
    symbol
}

type ExprUnary struct {
    op
    expr
}

type ExprBinary struct {
    op
    expr1
    expr2
}