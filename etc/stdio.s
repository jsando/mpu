        imports "strconv"

//
// Print the word passed on the stack to stdout in decimal.
//
PrintInteger(value word):
        psh value
        psh #buffer
        psh #BUFSIZE-1
        jsr Itoa
        pop #2
        pop ioPrintReq+2
        pop #2
        cpy 6, #ioPrintReq
        ret
.ioPrintReq
        dw 0x0101   // stdout putchars
        dw 0        // pointer to zero terminated string
.BUFSIZE    =   7
.buffer ds BUFSIZE

Println():
        cpy 6, #ioPrintReq
        ret
.ioPrintReq
        dw 0x0101   // stdout putchars
        dw lf        // pointer to zero terminated string
.lf     db 0x0a, 0x00