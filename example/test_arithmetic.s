// Unit tests for basic arithmetic operations
// Run with: mpu test test_arithmetic.s

        org 0x100

//
// Addition tests
//
test TestAddition():
        // Test 5 + 3 = 8
        cpy a, five
        add a, three
        sea
        cmp a, #8
        
        // Test 10 + 10 = 20
        cpy a, ten
        add a, ten
        sea
        cmp a, #20
        
        // Test 0 + 5 = 5
        cpy a, zero
        add a, five
        sea
        cmp a, #5
        ret

//
// Subtraction tests
//
test TestSubtraction():
        // Test 10 - 3 = 7
        cpy a, ten
        sub a, three
        sea
        cmp a, #7
        
        // Test 5 - 5 = 0
        cpy a, five
        sub a, five
        sea
        cmp a, #0
        
        // Test 3 - 5 = -2
        cpy a, three
        sub a, five
        sea
        cmp a, #-2
        ret

//
// Multiplication tests
//
test TestMultiplication():
        // Test 3 * 4 = 12
        cpy a, three
        mul a, four
        sea
        cmp a, #12
        
        // Test 5 * 0 = 0
        cpy a, five
        mul a, zero
        sea
        cmp a, #0
        
        // Test 7 * 1 = 7
        cpy a, seven
        mul a, one
        sea
        cmp a, #7
        ret

//
// Division tests
//
test TestDivision():
        // Test 12 / 3 = 4
        cpy a, twelve
        div a, three
        sea
        cmp a, #4
        
        // Test 7 / 1 = 7
        cpy a, seven
        div a, one
        sea
        cmp a, #7
        
        // Test 15 / 5 = 3
        cpy a, fifteen
        div a, five
        sea
        cmp a, #3
        ret

//
// Increment/Decrement tests
//
test TestIncDec():
        // Test increment
        cpy a, five
        inc a
        sea
        cmp a, #6
        
        // Test decrement
        cpy a, ten
        dec a
        sea
        cmp a, #9
        
        // Test increment zero
        cpy a, zero
        inc a
        sea
        cmp a, #1
        ret

//
// Test data
//
a:       dw 0
zero:    dw 0
one:     dw 1
three:   dw 3
four:    dw 4
five:    dw 5
seven:   dw 7
ten:     dw 10
twelve:  dw 12
fifteen: dw 15