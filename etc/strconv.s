// Itoa convert the integer value passed in into ASCII decimal representation
// in the buffer pointed to by 'buffer'.  On exit, buffer is left
// pointing to the first character (no longer the start of the buffer,
// as it generates chars right to left).
//
// Params:
//      value       The value to convert to decimal ASCII
//      buffer      Pointer to buffer to leave the ASCII string in (with null at end)
//      bsize       Size of buffer in bytes.  Value is right-aligned in buffer.
Itoa(value word, buffer word, bsize word):
    .next local word
    .t1 local word
    .t2 local word
        clc
        add buffer, bsize // start at right side of buffer
        dec buffer
        cpy *buffer, #0
.loop
        cmp value, #10
        jlt last
        dec buffer
        cpy next, value
        div next, #10
        cpy t1, next
        mul t1, #10
        clc
        cpy t2, #'0'
        add t2, value
        sec
        sub t2, t1
        seb
        cpy *buffer, t2
        clb
        cpy value, next
        jmp loop
.last
        dec buffer
        clc
        cpy t2, #'0'
        add t2, value
        seb
        cpy *buffer, t2
        clb
        ret
