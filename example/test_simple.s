// Simple unit tests demonstrating MPU test framework
// Run with: mpu test test_simple.s

        org 0x100

//
// Basic value tests
//
test TestConstants():
        // Test that constants have expected values
        sea
        cmp one, #1
        
        sea
        cmp five, #5
        
        sea
        cmp neg_one, #-1
        ret

//
// Basic arithmetic
//
test TestSimpleArithmetic():
        // Test 2 + 3 = 5
        cpy temp, two
        add temp, three
        sea
        cmp temp, #5
        
        // Test 10 - 3 = 7
        cpy temp, ten
        sub temp, three
        sea
        cmp temp, #7
        ret

//
// Memory access
//
test TestMemoryAccess():
        // Test direct memory write and read
        cpy value, forty_two
        sea
        cmp value, #42
        
        // Test indirect access
        cpy ptr, value_ptr
        cpy *ptr, ninety_nine
        sea
        cmp value, #99
        ret

//
// Flag tests
//
test TestFlags():
        // Test zero flag
        cpy temp, zero
        cmp temp, #0
        jeq zero_ok
        hlt  // Should not reach here
zero_ok:

        // Test negative flag
        cpy temp, neg_one
        cmp temp, #0
        jlt neg_ok
        hlt  // Should not reach here
neg_ok:
        ret

//
// Test data
//
zero:     dw 0
one:      dw 1
two:      dw 2
three:    dw 3
five:     dw 5
ten:      dw 10
neg_one:  dw -1
temp:     dw 0
value:    dw 0
ptr:      dw 0
value_ptr: dw value
forty_two: dw 42
ninety_nine: dw 99