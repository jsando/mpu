// Unit tests for common utility functions
// Run with: mpu test test_stdlib.s
        
        org 0x100

//
// Math functions
//

// abs - returns absolute value
abs:
        cmp arg1, #0
        jge abs_done    // If positive, return as is
        sub #0, arg1, result  // Negate if negative
        ret
abs_done:
        cpy arg1, result
        ret

// min - returns minimum of two values
min:
        cmp arg1, arg2
        jlt min_arg1
        cpy arg2, result
        ret
min_arg1:
        cpy arg1, result
        ret

// max - returns maximum of two values
max:
        cmp arg1, arg2
        jge max_arg1
        cpy arg2, result
        ret
max_arg1:
        cpy arg1, result
        ret

//
// Math tests
//

test TestAbs():
        // Test positive number
        cpy #42, arg1
        jsr abs
        sea
        cmp result, #42
        
        // Test negative number
        cpy #-42, arg1
        jsr abs
        sea
        cmp result, #42
        
        // Test zero
        cpy #0, arg1
        jsr abs
        sea
        cmp result, #0
        ret

test TestMin():
        // Test a < b
        cpy #5, arg1
        cpy #10, arg2
        jsr min
        sea
        cmp result, #5
        
        // Test a > b
        cpy #10, arg1
        cpy #5, arg2
        jsr min
        sea
        cmp result, #5
        
        // Test a == b
        cpy #7, arg1
        cpy #7, arg2
        jsr min
        sea
        cmp result, #7
        ret

test TestMax():
        // Test a < b
        cpy #5, arg1
        cpy #10, arg2
        jsr max
        sea
        cmp result, #10
        
        // Test a > b
        cpy #10, arg1
        cpy #5, arg2
        jsr max
        sea
        cmp result, #10
        
        // Test a == b
        cpy #7, arg1
        cpy #7, arg2
        jsr max
        sea
        cmp result, #7
        ret

//
// String functions
//

// strlen - returns length of null-terminated string
strlen:
        cpy #0, result
        cpy arg1, temp
strlen_loop:
        seb             // Byte mode
        cmp *temp, #0
        clb             // Word mode
        jeq strlen_done
        inc result
        inc temp
        jmp strlen_loop
strlen_done:
        ret

//
// String tests
//

test TestStrlen():
        // Test empty string
        cpy #empty_str, arg1
        jsr strlen
        sea
        cmp result, #0
        
        // Test "hello"
        cpy #hello_str, arg1
        jsr strlen
        sea
        cmp result, #5
        
        // Test single character
        cpy #single_char, arg1
        jsr strlen
        sea
        cmp result, #1
        ret

//
// Test data
//
arg1:   dw 0
arg2:   dw 0
result: dw 0
temp:   dw 0

empty_str:  db 0
hello_str:  db "hello", 0
single_char: db "x", 0