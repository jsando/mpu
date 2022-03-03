//---------------------------------------------------------
//  Print a random number between 0 and 100.
//---------------------------------------------------------
        import "stdio"
        import "strconv"
        import "random"
        
		dw main

IOREQ   = 6
IORES   = 8
				
        org 0x10
main():
        psh #0
        psh #100
        jsr Random
        pop #2
        jsr PrintInteger
        pop #2
        jsr Println
        hlt
