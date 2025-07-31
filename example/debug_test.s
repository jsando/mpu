// Simple test program for debugger

main:
    cpy r1,#10      // r1 = 10
    cpy r2,#20      // r2 = 20
    add r1,r2       // r1 = r1 + r2 (30)
    cpy r3,r1       // r3 = r1
    cmp r3,#30      // check if r3 == 30
    jeq success     // jump to success if equal
    hlt             // halt if not equal
    
success:
    cpy 0x2000,#1   // write 1 to memory
    hlt             // halt
    
r1: dw 0
r2: dw 0  
r3: dw 0