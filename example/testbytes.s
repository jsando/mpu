    dw main

    org 0x20
s1: db "hello",0
s2: ds 20

    org 0x100    
main():
    var p1 word
    var p2 word

    cpy p1, #s1
    cpy p2, #s2
.loop
    seb
    cpy *p2, *p1
    jeq done
    clb
    inc p1
    inc p2
    jmp loop
.done
    clb
    hlt
