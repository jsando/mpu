//
// Sqrt(n)
//
// I literally just coded this from the algorthim on wikipedia :) 
// https://en.wikipedia.org/wiki/Methods_of_computing_square_roots#Binary_numeral_system_(base_2)
//            
sqrt(res word, n word):                
            .x local word
            .c local word
            .d local word
            .t1 local word

            cpy x, n
            cpy c, #0
            cpy d, #0b0100_0000_0000_0000
.loop1
            cmp d, n
            jlt loop2
            jeq loop2
            div d, #4
            jmp loop1
.loop2
            cmp d, #0
            jeq done
            cpy t1, c
            add t1, d
            cmp x, t1
            jlt else
            sub x, t1
            div c, #2
            add c, d
            jmp next
.else
            div c, #2
.next
            div d, #4
            jmp loop2            
.done
            cpy res, c
            ret                        

