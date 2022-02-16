//-------------------------------------
// First ever mpu test program!
//-------------------------------------
pc:	    dw main		// entry point
sp:	    dw 0xfffe	// stack pointer (grows down)

	    org 0x0100
main:
.ptr 	= 2
	psh #0x2000
.loop
    cpy *[sp+ptr], #0xd0d0
    add [sp+ptr], #2
    cmp [sp+ptr], #0x3000
    jlt loop
    db 0 // halt
