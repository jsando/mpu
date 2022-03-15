                dw main          // Set initial PC to start at main
                org 0x10
main:
                cpy 0x06,#myreq // 0x06 is the IO request register
                hlt 

myreq:          dw 0x0101       // stdout / putchars
                dw hello        // pointer to zero terminated string
hello:          db "Hello, world!",0x0a,0x00
