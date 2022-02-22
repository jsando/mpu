//  Print the numbers from 1 to 100.

        dw main

IOREQ   = 6
IORES   = 8

        org 0x10
myreq:  dw 1        // stdout
        dw 1        // putchars
        dw 0        // pointer to zero terminated string

buffer: ds 10       // max 10 digits

main():
    .i local word
        cpy i, #1
.loop
        psh i
        jsr PrintWord
        pop #2
        inc i
        cmp i, #100
        jlt loop
        hlt

PrintWord(value word):
        psh value
        psh #buffer
        jsr WordToAscii
        pop myreq+4
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
        cpy *buffer, #0x0a
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
        cpy t2, #'0
        add t2, value
        seb
        cpy *buffer, t2
        clb
        ret
