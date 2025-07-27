                req #myreq // 0x06 is the IO request register
                hlt 

myreq:          dw 0x0101       // stdout / putchars
                dw 0            // space for an error return code
                dw hello        // pointer to zero terminated string
hello:          db "Hello, world!",0x0a,0x00
