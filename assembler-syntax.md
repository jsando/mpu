# Memory Processing Unit (MPU)

MPU is a (fake) 16 bit microprocessor, with a small enough instruction set to be easy to learn and play with.  It started as a 6502 emulator, but then evolved because hey that's been done a bunch before so why not try something different?

It is named MPU because unlike real processors it has no registers ... all operations are directly on main memory.  Its not *strictly* true that there are NO registers, there is still a Program Counter, a Stack Pointer, and Frame Pointer but these are located in memory at address 0, 2, and 4 respectively.  There are also flags such as the carry flag, zero flag, negative flag ... which are not mapped to main memory.

MPU features:

* 16 bit address space
* Instructions can operate on bytes or words
* DMA hardware for display, file i/o, network, etc
* Stack can be anywhere in memory and grows downward
* Frame pointer makes it easy to write reusable/reentrant functions

## Address Modes

The following address modes are supported:

* Implied - the operand(s) are implied by the instruction.  Ex, "ret" returns from subroutine using the address on the top of the stack.
* Immediate - the operand(s) are constant values encoded following the instruction.  Usually indicated with a number sign in soure code, ie "#1000" means "the number 1000", as opposed to "the value at address location 1000".  Most instructions using immediate mode will use the '#' to indicate as such, however the jump instructions allow it to be omitted since they *only* work with immediate mode.
* ImmediateByte - at least one instruction supports a single-byte immediate value and that is 'pop #', which pops and discards the given number of bytes from the stack.
* OffsetByte - The operand is a relative offset from the current program counter to the jump target.  Used by conditional jumps, it means the jump can be +127/-128 bytes forward/backward.
* Absolute - the operand(s) refer to the value at the given 16 bit memory address.  Ie, "0" means "the value stored at address 0, which would be the program counter (the address of the current instruction being executed).
* Indirect - the operand(s) refer to an address, which contains an address which contains a value.  If memory locations 100 and 101 contain "20 20", and location 2020 contains "12 34", then *100 is 12 34.
* Relative - frame pointer relative 8 bit offset. 
* Relative Indirect - frame pointer relative 8 bit offset, as an indirect reference.

One of the ways the 6502 reduced the number of bytes of a program was by leveraging "zero page" modes, which allowed a single byte to refer to a 16-bit pointer.  MPU gains a similar benefit by using frame-pointer relative addressing with a single byte.  The benefit over zero page is it makes it much easier to write reusable functions, since they aren't using global variables.

Consider a 6502 with a zero page pointer to a value, adding a constant to a 16 bit value.

    ptr = $10 // some zero page value
    lda #0
    sta ptr
    lda #$20
    sta ptr+1
    clc
    ldy #0
    lda (ptr), y
    adc #50
    sta (ptr), y
    iny
    lda (ptr), y
    adc #0
    sta (ptr), y

And the same in MPU, using fp relative-indirect mode.

    cpy fp+2, #$2000
    clc
    adc [fp+2], #50

The two relative modes using the frame pointer as a base.  The frame pointer can be directly manipulated by reading/writing to address 4, but better is to use the sav/rst opcodes.  The assembler also has a fp-automatic syntax that automates all of that.

## Byte vs Word Modes

MPU has a "bytes mode" flag, which can switch MPU into byte mode.  In this mode all instructions operate on bytes instead of words.  At startup, the bytes flag is cleared therefore at startup MPU always starts in word mode (16 bit).

Words are stored low byte first.

To enable bytes mode, use 'seb'.  To enable word mode, use 'clb'.  These instructions "set" or "clear" the bytes flag.

## Flags

All operations that update a value will update the flags, including pop/psh, cpy, add, and so on.

* Carry
* Zero
* Negative
* Bytes

## Instructions

Instructions that take two operands, such as 'add', perform the operation specified and store the result in the first operand.

For example,

    add a, b

Is effectively:

    a = a + b

* Hlt - Halt processing.  If run from the command line this signals an exit.
* Add - Add, with carry.  a += b
* Sub - Subtract, with carry. a -= b
* Mul - multiply. a *= b
* Div - divide.  a /= b
* And - bitwise and.  a &= b
* Or - bitwise or.  a |= b
* Xor - bitwise exclusive-or.  a ^= b
* Cpy - copy. a = b
* Cmp - Compare.  Its like sub but without storing the result, but it updates the flags.
* Inc - Increment by 1.  Can use clc/add but this is shorter.
* Dec - Decrement by 1.  Can use sec/sub but this is shorter.
* Psh - Push operand on the stack.
* Pop - Pop values off the stack.
* Jsr - Jump to subroutine.  Pushes the return address on the stack, so ret/rst can return.
* Jmp - Unconditional jump.  Like goto.  Operand is a 16 bit address.
* Jeq - Jump if equal/zero.  Operand is 8 bit offset, max 127/-128.
* Jne - Jump if not equal/not zero.  Operand is 8 bit offset, max 127/-128.
* Jge - Jump if greater than or equal to.  Operand is 8 bit offset, max 127/-128.
* Jlt - Jump if less than. Operand is 8 bit offset, max 127/-128.
* Jcc - Jump if carry clear.  Operand is 8 bit offset, max 127/-128.
* Jcs - Jump if carry set.  Operand is 8 bit offset, max 127/-128.
* Sav - Save the framepointer, set it to the stack pointer, and allocate space for local vars (#bytes).
* Seb - Set bytes mode flag.
* Clb - Clear bytes mode flag.
* Clc - Clear carry flag.
* Sec - Set carry flag.
* Ret - Return from subroutine, using 16 bit address on top of stack.
* Rst - Restore framepointer and return from subroutine.

# My todo list

- Profile performance
- Add include directive so I can start building reusable functions
- Ugh ... really wish I had an indirect-indexed mode, when given a pointer to a struct I need constant offsets off the pointer
    - Suppose I could add a 1 byte (uint8) to relative-indirect, if none specified its zero?  My lovely byte savings go away :(
    - Tried it on some sample code and it cleans it up a lot
- a gofmt-equivalent would be nice.  I think the lexer would need a facelift, and the parser would need to emit 100% of the file as Statements, so the formatter could then walk that and output it cleanly.
- Is there a way to add unit tests?  That would make writing short programs much more fun and easy to test.
- Need to cleanup uint16 vs int everywhere, make up your mind
- cleanup
    - lexer
        - encapsulate Scanner completely ... find calls to lexer.s.foo and fix them
        - text.scanner leaves quotes on strings, ticks on chars, etc
        - would token category help?  directive, opcode, etc
        - don't use text.Scanner ... just use my own.  Use $ instead of 0x, ; instead of //.
    - parser
        - review all the parse functions and make sure they follow the same pattern ... do they call lexer.next?
        - sometimes I use tok := lexer.next and sometimes lexer.tok
    - machine
        - make updating the flags more explicit so I don't accidentally use writeTarget() for updating the sp for example
- monitor needs a way to view stack contents ... not sure how though unless we know whether they are bytes or words
- could I actually build a debugger that could inspect variables?

# Input/Output

- RNG
    - [0, n)
- Current time?  There are two purposes, one is to measure changes in time.  ms since 1970 is a big number, but ms since 2022 is smaller :) Or maybe 1/4 seconds is enough?  That would overflow a 16 bit word every 4 hours.
- Graphics
    - Clear screen (w/color)
    - Set color
    - Draw line
    - Draw rect
    - Fill rect

Wondering if I can attach devices purely through DMA, with no ROM.  Its probably more flexible if I can have a function table for each device, and pass params on the stack ... but then those are all 'fake' addresses, a user won't be able to disassemble and see MPU code at those addresses (becaus I'm planning to build device drivers in Go).

That means I need some kind of command packet format.  Then programs write those packets to the IO device.  Or, write the packet into main memory and tell the device the address.  For that matter, I don't need separate memory addresses for different devices ... that can be part of the packet.

deviceId dw 0
command  dw 0
params   ds 50

In this case the params are just structs of whatever each device/command expects.  Ah ... and the results can be DMA'd to the wherever the user wants them.

At this point I think I'll put the IO hook into the first 16 bytes along with my other registers.

Console I/O:
    write
    read

I was able to use encoding/binary to decode handler structs from mpu to Go.  I can generalize that, and make deviceId/Command into bytes in a single word.  Have devices register handlers and don't register devices at all.  Prob just use a map of handler in machine.

## Graphics

Init (open window)
PollEvents -> event
SetColor(r, g, b, a)
DrawLine(x, y, x2, y2)
DrawRect(x, y, w, h)
FillRect(x, y, w, h)
Present

Random uint16 at address 0x0a

# System Monitor

command params

d/dump [start [end]]
l/list [start]
set [address value [value]*]
run address
s/step [address]

mpu -i file.bin -m - load file and enter monitor

# Assembler

## Symbols & Local Symbols

Symbols are declared and equate to the current offset, unless defined by an equate.

symbol

Local labels are prefixed with '.' and are scoped to the previous symbol.

foo:
.l1
.l2

bar:
.l1
.l2

currently I have bar.l1 defined ... resolving symbols however is only looking for 'l1'.  Maybe I should have it search first for bar.l1, then .l1?  Yes.  What do I need to make that work ... need to know the current global label in scope for each statement.

## Define space

ds <count>[, pattern]
db number[, number] | string
dw word

## Constants

symbol = expression     // global
.symbol = expression    // local

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

## Assembly Grammar

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

