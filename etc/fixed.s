            import "stdio"
            import "sqrt"

            org 0
            dw main

            org 0x10


main():
    .a local word
    .b local word

            psh #0
            cpy a, #0b0010_1000
            mul a, #16
            psh a
            jsr sqrt
            pop #2
            jsr printfp
            jsr Println
            hlt

printfp(value word):
            .t1 local word
            .fraction local word

            cpy t1, value
            div t1, #16
            psh t1
            jsr PrintInteger
            pop #2
            cpy 6, #ioPrintDec

            cpy fraction, #0
            cpy t1, value
            and t1, #0b_1000
            jeq b2
            add fraction, #5000
.b2
            cpy t1, value
            and t1, #0b_0100
            jeq b1            
            add fraction, #2500
.b1
            cpy t1, value
            and t1, #0b_0010
            jeq b0
            add fraction, #1250
.b0
            cpy t1, value
            and t1, #0b_0001
            jeq printit
            add fraction, #625
.printit
            cpy fracbuf, #0x3030
            cpy fracbuf+2, #0x3030
            psh fraction
            psh #fracbuf
            psh #5
            jsr Itoa
            pop #6
            cpy 6, #ioPrintFrac
            ret
.ioPrintDec
        dw 0x0101   // stdout putchars
        dw dec        // pointer to zero terminated string
.dec    db '.', 0x00

.ioPrintFrac
        dw 0x0101   // stdout putchars
        dw fracbuf       // pointer to zero terminated string
.fracbuf ds 5
