        include "stdio.s"
        include "strconv.s"

//---------------------------------------------------------
//  Print the numbers from 1 to 100 to the console.
//---------------------------------------------------------
main():
        var i word          // Declare variable 'i' as a word (on the stack)
        cpy i, #1           // i = 1
.loop:
        psh i               // push i on stack
        jsr PrintInteger    // print i
        pop #2              // remove 2 bytes (i) from stack
        jsr Println         // print newline
        inc i               // i = i + 1
        cmp i, #101         // i >= 100?
        jlt loop            // jump if less than to 'loop'
        hlt                 // halt
