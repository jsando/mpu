        dw main
        dw 0xfffe

	    org 0x100

value:  dw     12345
result: ds     100 // space for output

main:
        psh     value
        psh     #result
        jsr     atoi
        db 0

atoi:
    .ptrResult   =   4
    .temp        =   4

        cpy     ptrResult, result
.loop
        cpy     [sp+temp], value

        seb
        rem     [sp+temp], #10
        add     temp, #'0'
        cpy     *ptrResult, temp
        clb
        inc     ptrResult
        div     value, #10
        jne     loop
        cpy     *ptrResult, #0 // null-terminate
        ret

// Compute the remainder of the two args on the stack.
//  sp+4 return address
//  sp+6 b
//  sp+8 a
//  sp+10 a % b
//     a - (a / b * b)
// func remainder(a word, b word) (result word)
remainder:
    .temp = 2
    .a  = 6
    .b  = 8
    .result = 10

    psh [sp+a]                          // temp = a
    div [sp+temp], [sp+b]               // temp /= b
    mul [sp+temp], [sp+b]               // temp *= b
    cpy [sp+result], [sp+a]             // result = a
    sec
    sub [sp+result], [sp+temp]
    ret
