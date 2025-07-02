// Unit tests for basic arithmetic operations
// Run with: mpu test test_arithmetic.s

        org 0x100

//
// Addition tests
//
test TestAddition():
        // Test 5 + 3 = 8
        cpy five, a
        add three, a
        sea
        cmp a, #8
        
        // Test 10 + 10 = 20
        cpy ten, a
        add ten, a
        sea
        cmp a, #20
        
        // Test 0 + 5 = 5
        cpy zero, a
        add five, a
        sea
        cmp a, #5
        ret

//
// Subtraction tests
//
test TestSubtraction():
        // Test 10 - 3 = 7
        cpy ten, a
        sub three, a
        sea
        cmp a, #7
        
        // Test 5 - 5 = 0
        cpy five, a
        sub five, a
        sea
        cmp a, #0
        
        // Test 3 - 5 = -2
        cpy three, a
        sub five, a
        sea
        cmp a, #-2
        ret

//
// Multiplication tests
//
test TestMultiplication():
        // Test 3 * 4 = 12
        cpy three, a
        mul four, a
        sea
        cmp a, #12
        
        // Test 5 * 0 = 0
        cpy five, a
        mul zero, a
        sea
        cmp a, #0
        
        // Test 7 * 1 = 7
        cpy seven, a
        mul one, a
        sea
        cmp a, #7
        ret

//
// Division tests
//
test TestDivision():
        // Test 12 / 3 = 4
        cpy twelve, a
        div three, a
        sea
        cmp a, #4
        
        // Test 7 / 1 = 7
        cpy seven, a
        div one, a
        sea
        cmp a, #7
        
        // Test 15 / 5 = 3
        cpy fifteen, a
        div five, a
        sea
        cmp a, #3
        ret

//
// Increment/Decrement tests
//
test TestIncDec():
        // Test increment
        cpy five, a
        inc a
        sea
        cmp a, #6
        
        // Test decrement
        cpy ten, a
        dec a
        sea
        cmp a, #9
        
        // Test increment zero
        cpy zero, a
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