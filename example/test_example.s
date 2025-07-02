// Test example for MPU unit testing
// Run with: mpu test test_example.s

        org 0x100
        
// Some data to test
value:  dw 42
array:  dw 1, 2, 3, 4, 5

// Test passing case
test TestValueEquals42():
        sea
        cmp value, #42
        ret

// Test array first element
test TestArrayFirstElement():
        sea
        cmp array, #1
        ret

// Test failing case (for demonstration)
test TestFailingExample():
        sea
        cmp value, #99  // This will fail since value is 42
        ret

// Test arithmetic
test TestAddition():
        cpy ten, result
        add twenty, result
        sea
        cmp result, #30
        ret

ten:    dw 10
twenty: dw 20

result: dw 0