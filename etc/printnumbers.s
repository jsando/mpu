//  Print the numbers from 1 to 100 to the console.

        dw main

IOREQ   = 6
IORES   = 8

        org 0x10
myreq:  dw 0x0101   // stdout putchars
        dw 0        // pointer to zero terminated string

buffer: ds 10       // max 10 digits

main():
.i local word
        cpy i, #1
.loop
        psh #0
        psh #1000
        jsr Random
        pop #2
        jsr PrintInteger
        pop #2
        inc i
        cmp i, #101
        jlt loop
        hlt

Random(result word, range word):
    .i local word
    .j local word
    cpy i, 10   // get a random number in range 0-65535
    cpy j, i        // value / range * range
    div j, range
    mul j, range
    cpy result, i
    sec
    sub result, j
    ret

PrintInteger(value word):
        psh value
        psh #buffer
        jsr WordToAscii
        pop myreq+2
        pop #2
        cpy IOREQ, #myreq
        ret

WordToAscii(value word, buffer word):
    .next local word
    .t1 local word
    .t2 local word
        clc
        add buffer, #9 // start at right side of buffer
        cpy *buffer, #0
        dec buffer
        cpy *buffer, #10
.loop
        cmp value, #10
        jlt last
        dec buffer
        cpy next, value
        div next, #10
        cpy t1, next
        mul t1, #10
        clc
        cpy t2, #'0'
        add t2, value
        sec
        sub t2, t1
        seb
        cpy *buffer, t2
        clb
        cpy value, next
        jmp loop
.last
        dec buffer
        clc
        cpy t2, #'0'
        add t2, value
        seb
        cpy *buffer, t2
        clb
        ret
