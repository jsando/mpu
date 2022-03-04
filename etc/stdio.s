        import "strconv"

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

//
// Print the given value as a 4-digit hexadecimal number.
//
PrintHex(hi byte, lo byte):
                .t1 local byte

                seb
                cpy t1, hi
                div t1, #16
                psh t1
                jsr ToHex
                pop buffer+2
                cpy t1, hi
                psh t1
                jsr ToHex
                pop buffer+3
                cpy t1, lo
                div t1, #16
                psh t1
                jsr ToHex
                pop buffer+4
                cpy t1, lo
                psh t1
                jsr ToHex
                pop buffer+5
                cpy buffer+6, #0
                clb
                cpy 6, #ioPrintReq
                ret
.ioPrintReq
                dw 0x0101       // stdout putchars
                dw buffer       // pointer to zero terminated string
.buffer         db "0x0000", 0

ToHex(value word):
                and value, #0x0f
                cmp value, #0x0a
                jlt number
                sub value, #0x0a
                add value, #'a'
                ret
.number         add value, #'0'
                ret
