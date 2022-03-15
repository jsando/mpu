        import "stdio"
        import "strconv"

// requires: strconv, stdio
//---------------------------------------------------------
//  Print the numbers from 1 to 100 to the console.
//---------------------------------------------------------

        dw main

IOREQ   = 6
IORES   = 8

        org 0x10
main():
        var i word

        cpy i, #1
.loop
        psh i
        jsr PrintInteger
        pop #2
        jsr Println
        inc i
        cmp i, #101
        jlt loop
        hlt
