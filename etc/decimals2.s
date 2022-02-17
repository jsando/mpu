        dw main
        dw 0x1000

result: dw  0

	    org 0x100
main:
        psh     #0 // space for result
        psh     #123
        psh     #10
        jsr     remainder
        pop     #4                // discard params
        pop     result
        hlt

// Compute the remainder of the two args on the stack.
//     a - (a / b * b)
remainder(result word, a word, b word):
    .temp local word

        cpy temp, a               // temp = a
        div temp, b               // temp /= b
        mul temp, b               // temp *= b
        cpy result, a             // result = a
        sec
        sub result, temp          // result -= temp
        ret

// with word fp relative 20+1+5+1+3 = 30 bytes
// with byte fp relative 15+1+1+2 = 19 bytes!  holy shit a 36% savings is big.
