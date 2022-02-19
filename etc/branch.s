    dw main
    org 0x100
main:
    psh #0x1000
    psh #0x8000
    psh #0x1111
    jsr fill_memory
    pop #6
    db 0

fill_memory(start word, end word, value word):
    .ptr local word

    cpy ptr, start
.loop
    cpy *ptr, value
    inc ptr
    cmp ptr, end
    jlt loop
    jge skip
    db 0
    db 0
    db 0
.skip
    ret
