// All passing tests example
        org 0x100
        
value:  dw 42

test TestPass1():
        sea
        cmp value, #42
        ret

test TestPass2():
        sea
        cmp value, #42
        ret