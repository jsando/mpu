            dw main
IOREQ   = 6
            org 0x20
main():
            psh #0
            psh #225
            jsr sqrt
            pop #2
            jsr PrintInteger
            pop #2
            hlt

//
// Sqrt(n)
//
// I literally just coded this from the algorthim on wikipedia :) 
// https://en.wikipedia.org/wiki/Methods_of_computing_square_roots#Binary_numeral_system_(base_2)
//            
sqrt(res word, n word):                
            .x local word
            .c local word
            .d local word
            .t1 local word

            cpy x, n
            cpy c, #0
            cpy d, #0b0100_0000_0000_0000
.loop1
            cmp d, n
            jlt loop2
            jeq loop2
            div d, #4
            jmp loop1
.loop2
            cmp d, #0
            jeq done
            cpy t1, c
            add t1, d
            cmp x, t1
            jlt else
            sub x, t1
            div c, #2
            add c, d
            jmp next
.else
            div c, #2
.next
            div d, #4
            jmp loop2            
.done
            cpy res, c
            ret                        

//
// Print the word passed on the stack to stdout in decimal.
//
PrintInteger(value word):
        psh value
        psh #buffer
        jsr WordToAscii
        pop ioPrintReq+2
        pop #2
        cpy IOREQ, #ioPrintReq
        ret

.ioPrintReq
        dw 0x0101   // stdout putchars
        dw 0        // pointer to zero terminated string
.buffer ds 11       // max 10 digits + null

//
// Convert the value passed in into ASCII decimal representation
// in the buffer pointed to by 'buffer'.  On exit, buffer is left
// pointing to the first character (no longer the start of the buffer,
// as it generates chars right to left).
//
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
