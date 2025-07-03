// Demonstrates runtime error handling in tests
// Run with: mpu test test_errors.s

        org 0x100

test TestDivideByZero():
        // This will cause a runtime error
        cpy a, ten
        div a, zero     // Divide by zero!
        ret

test TestValidTest():
        // This should pass
        cpy a, ten
        div a, two
        sea
        cmp a, #5
        ret

// Test data
a:    dw 0
zero: dw 0
two:  dw 2
ten:  dw 10