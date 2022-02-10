//-------------------------------------
// First ever lilac test program!
//-------------------------------------
pc:	    dw main		// entry point
sp:	    dw 0xffff	// stack pointer (grows down)

	    org 0x0100
ptr:    dw 0

main:   cpy ptr, #0x2000
loop:   cpy *ptr, #0xd0d0
        add ptr, 2
        cmp ptr, #0x3000
        jlt loop
        db 0 // halt

