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
        cpy result, zero    // result = 0
        sub result, arg1    // result = 0 - arg1
        ret
abs_done:
        cpy result, arg1
        ret

// min - returns minimum of two values
min:
        cmp arg1, arg2
        jlt min_arg1
        cpy result, arg2
        ret
min_arg1:
        cpy result, arg1
        ret

// max - returns maximum of two values
max:
        cmp arg1, arg2
        jge max_arg1
        cpy result, arg2
        ret
max_arg1:
        cpy result, arg1
        ret

//
// Math tests
//

test TestAbs():
        // Test positive number
        cpy arg1, forty_two
        jsr abs
        sea
        cmp result, #42
        
        // Test negative number
        cpy arg1, neg_forty_two
        jsr abs
        sea
        cmp result, #42
        
        // Test zero
        cpy arg1, zero
        jsr abs
        sea
        cmp result, #0
        ret

test TestMin():
        // Test a < b
        cpy arg1, five
        cpy arg2, ten
        jsr min
        sea
        cmp result, #5
        
        // Test a > b
        cpy arg1, ten
        cpy arg2, five
        jsr min
        sea
        cmp result, #5
        
        // Test a == b
        cpy arg1, seven
        cpy arg2, seven
        jsr min
        sea
        cmp result, #7
        ret

test TestMax():
        // Test a < b
        cpy arg1, five
        cpy arg2, ten
        jsr max
        sea
        cmp result, #10
        
        // Test a > b
        cpy arg1, ten
        cpy arg2, five
        jsr max
        sea
        cmp result, #10
        
        // Test a == b
        cpy arg1, seven
        cpy arg2, seven
        jsr max
        sea
        cmp result, #7
        ret

//
// String functions
//

// strlen - returns length of null-terminated string
strlen:
        cpy result, zero
        cpy temp, arg1
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
        cpy arg1, empty_str_ptr
        jsr strlen
        sea
        cmp result, #0
        
        // Test "hello"
        cpy arg1, hello_str_ptr
        jsr strlen
        sea
        cmp result, #5
        
        // Test single character
        cpy arg1, single_char_ptr
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

// Constants
zero: dw 0
five: dw 5
seven: dw 7
ten: dw 10
forty_two: dw 42
neg_forty_two: dw -42

// String pointers
empty_str_ptr: dw empty_str
hello_str_ptr: dw hello_str
single_char_ptr: dw single_char

// Strings
empty_str:  db 0
hello_str:  db "hello", 0
single_char: db "x", 0