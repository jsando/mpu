// Unit tests for bit manipulation operations
// Run with: mpu test test_bits.s

        org 0x100

//
// Count set bits in a word
//
count_bits:
        cpy #0, bit_count
        cpy value, temp
count_loop:
        cmp temp, #0
        jeq count_done
        
        // Check lowest bit
        and temp, #1, test_bit
        add test_bit, bit_count
        
        // Shift right by dividing by 2
        div temp, #2, temp
        jmp count_loop
count_done:
        ret

//
// Check if number is power of 2
//
is_power_of_2:
        cpy #0, is_pow2     // Default to false
        
        // Zero is not a power of 2
        cmp value, #0
        jeq pow2_done
        
        // n & (n-1) == 0 for powers of 2
        cpy value, temp
        dec temp            // temp = n - 1
        and value, temp, temp
        cmp temp, #0
        jne pow2_done
        cpy #1, is_pow2     // It's a power of 2
pow2_done:
        ret

//
// Reverse bits (simplified for 8-bit)
//
reverse_bits:
        seb                 // Set to byte mode
        cpy #0, reversed
        cpy #8, counter     // 8 bits to process
        cpy value, temp
reverse_loop:
        cmp counter, #0
        jeq reverse_done
        
        // Shift reversed left
        mul reversed, #2, reversed
        
        // Add lowest bit of temp
        and temp, #1, test_bit
        or reversed, test_bit, reversed
        
        // Shift temp right
        div temp, #2, temp
        
        dec counter
        jmp reverse_loop
reverse_done:
        clb                 // Clear byte mode
        ret

//
// Tests
//

test TestCountBits():
        // Test 0b1010 (10) = 2 bits
        cpy #10, value
        jsr count_bits
        sea
        cmp bit_count, #2
        
        // Test 0b1111 (15) = 4 bits
        cpy #15, value
        jsr count_bits
        sea
        cmp bit_count, #4
        
        // Test 0 = 0 bits
        cpy #0, value
        jsr count_bits
        sea
        cmp bit_count, #0
        
        // Test 0b10000000 (128) = 1 bit
        cpy #128, value
        jsr count_bits
        sea
        cmp bit_count, #1
        ret

test TestIsPowerOf2():
        // Test 8 (power of 2)
        cpy #8, value
        jsr is_power_of_2
        sea
        cmp is_pow2, #1
        
        // Test 16 (power of 2)
        cpy #16, value
        jsr is_power_of_2
        sea
        cmp is_pow2, #1
        
        // Test 10 (not power of 2)
        cpy #10, value
        jsr is_power_of_2
        sea
        cmp is_pow2, #0
        
        // Test 0 (not power of 2)
        cpy #0, value
        jsr is_power_of_2
        sea
        cmp is_pow2, #0
        
        // Test 1 (power of 2)
        cpy #1, value
        jsr is_power_of_2
        sea
        cmp is_pow2, #1
        ret

test TestReverseBits():
        // Test 0b10110000 -> 0b00001101
        cpy #0xB0, value    // 176
        jsr reverse_bits
        sea
        cmp reversed, #0x0D // 13
        
        // Test 0b11111111 -> 0b11111111
        cpy #0xFF, value
        jsr reverse_bits
        sea
        cmp reversed, #0xFF
        
        // Test 0b00000001 -> 0b10000000
        cpy #0x01, value
        jsr reverse_bits
        sea
        cmp reversed, #0x80
        ret

//
// Variables
//
value:      dw 0
temp:       dw 0
bit_count:  dw 0
is_pow2:    dw 0
reversed:   dw 0
counter:    dw 0
test_bit:   dw 0