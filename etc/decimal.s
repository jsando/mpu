pc:	dw main
sp:	dw 0xffff

	    org 0x100

value:  dw     12345
result: ds     100 // space for output

main:
        psh     value
        psh     #result
        jsr     atoi
        db 0

atoi:
ptrResult:   =   2
temp:        =   4

        cpy     ptrResult, result
loop:
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

