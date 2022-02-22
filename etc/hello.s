        dw main

IOREQ   = 6
IORES   = 8

        org 0x100
main():
    .i local word
        cpy i, #10
.loop
        cpy IOREQ, #myreq
        dec i
        jne loop
        hlt

        org 0x1000
myreq:  dw 1        // stdout
        dw 1        // putchars
        dw hello    // pointer to zero terminated string

hello:  db "Hello, world!", 0x0a, 0x00
