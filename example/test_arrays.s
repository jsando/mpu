// Unit tests for array operations
// Run with: mpu test test_arrays.s

        org 0x100

//
// Array sum function
//
// Calculates sum of array elements
// Input: array_ptr - pointer to array
//        array_len - number of elements
// Output: sum - sum of all elements
//
array_sum:
        cpy sum, zero       // Initialize sum
        cpy i, zero         // Initialize counter
sum_loop:
        cmp i, array_len
        jeq sum_done        // Exit if i == len
        
        // Add array[i] to sum
        cpy temp, i
        add temp, temp      // temp = i * 2 (word offset)
        add temp, array_ptr // temp = array_ptr + offset
        add sum, *temp      // sum += array[i]
        
        inc i
        jmp sum_loop
sum_done:
        ret

//
// Find maximum element in array
//
array_max:
        cpy temp, array_ptr
        cpy max_val, *temp  // max = array[0]
        cpy i, one          // Start from index 1
max_loop:
        cmp i, array_len
        jeq max_done
        
        cpy temp, i
        add temp, temp      // temp = i * 2
        add temp, array_ptr
        cmp *temp, max_val
        jlt skip_update
        cpy max_val, *temp  // Update max if current > max
skip_update:
        inc i
        jmp max_loop
max_done:
        ret

//
// Tests
//

test TestArraySum():
        // Test sum of [1, 2, 3, 4, 5] = 15
        cpy array_ptr, test_array1_addr
        cpy array_len, five
        jsr array_sum
        sea
        cmp sum, #15
        ret

test TestArraySumEmpty():
        // Test empty array sum = 0
        cpy array_ptr, test_array1_addr
        cpy array_len, zero
        jsr array_sum
        sea
        cmp sum, #0
        ret

test TestArraySumSingle():
        // Test single element [42] = 42
        cpy array_ptr, single_array_addr
        cpy array_len, one
        jsr array_sum
        sea
        cmp sum, #42
        ret

test TestArrayMax():
        // Test max of [3, 7, 2, 9, 1] = 9
        cpy array_ptr, test_array2_addr
        cpy array_len, five
        jsr array_max
        sea
        cmp max_val, #9
        ret

test TestArrayMaxNegative():
        // Test max of [-5, -2, -8, -1] = -1
        cpy array_ptr, neg_array_addr
        cpy array_len, four
        jsr array_max
        sea
        cmp max_val, #-1
        ret

//
// Test data
//
array_ptr:  dw 0
array_len:  dw 0
sum:        dw 0
max_val:    dw 0
i:          dw 0
temp:       dw 0

// Constants
zero: dw 0
one: dw 1
four: dw 4
five: dw 5

// Array addresses
test_array1_addr: dw test_array1
test_array2_addr: dw test_array2
single_array_addr: dw single_array
neg_array_addr: dw neg_array

// Arrays
test_array1: dw 1, 2, 3, 4, 5
test_array2: dw 3, 7, 2, 9, 1
single_array: dw 42
neg_array:   dw -5, -2, -8, -1