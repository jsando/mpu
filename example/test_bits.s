// Unit tests for bit manipulation operations
// Run with: mpu test test_bits.s

        org 0x100

//
// Count set bits in a word
//
count_bits:
        cpy bit_count, zero
        cpy temp, value
count_loop:
        cmp temp, #0
        jeq count_done
        
        // Check lowest bit
        cpy test_bit, temp
        and test_bit, one
        add bit_count, test_bit
        
        // Shift right by dividing by 2
        div temp, two
        jmp count_loop
count_done:
        ret

//
// Check if number is power of 2
//
is_power_of_2:
        cpy is_pow2, zero   // Default to false
        
        // Zero is not a power of 2
        cmp value, #0
        jeq pow2_done
        
        // n & (n-1) == 0 for powers of 2
        cpy temp, value
        dec temp            // temp = n - 1
        and temp, value
        cmp temp, #0
        jne pow2_done
        cpy is_pow2, one    // It's a power of 2
pow2_done:
        ret

//
// Reverse bits (simplified for 8-bit)
//
reverse_bits:
        seb                 // Set to byte mode
        cpy reversed, zero
        cpy counter, eight  // 8 bits to process
        cpy temp, value
reverse_loop:
        cmp counter, #0
        jeq reverse_done
        
        // Shift reversed left
        mul reversed, two
        
        // Add lowest bit of temp
        cpy test_bit, temp
        and test_bit, one
        or reversed, test_bit
        
        // Shift temp right
        div temp, two
        
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
        cpy value, ten
        jsr count_bits
        sea
        cmp bit_count, #2
        
        // Test 0b1111 (15) = 4 bits
        cpy value, fifteen
        jsr count_bits
        sea
        cmp bit_count, #4
        
        // Test 0 = 0 bits
        cpy value, zero
        jsr count_bits
        sea
        cmp bit_count, #0
        
        // Test 0b10000000 (128) = 1 bit
        cpy value, val_128
        jsr count_bits
        sea
        cmp bit_count, #1
        ret

test TestIsPowerOf2():
        // Test 8 (power of 2)
        cpy value, eight
        jsr is_power_of_2
        sea
        cmp is_pow2, #1
        
        // Test 16 (power of 2)
        cpy value, sixteen
        jsr is_power_of_2
        sea
        cmp is_pow2, #1
        
        // Test 10 (not power of 2)
        cpy value, ten
        jsr is_power_of_2
        sea
        cmp is_pow2, #0
        
        // Test 0 (not power of 2)
        cpy value, zero
        jsr is_power_of_2
        sea
        cmp is_pow2, #0
        
        // Test 1 (power of 2)
        cpy value, one
        jsr is_power_of_2
        sea
        cmp is_pow2, #1
        ret

test TestReverseBits():
        // Test 0b10110000 -> 0b00001101
        cpy value, val_b0   // 176
        jsr reverse_bits
        sea
        cmp reversed, #0x0D // 13
        
        // Test 0b11111111 -> 0b11111111
        cpy value, val_ff
        jsr reverse_bits
        sea
        cmp reversed, #0xFF
        
        // Test 0b00000001 -> 0b10000000
        cpy value, one
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

// Constants
zero:       dw 0
one:        dw 1
two:        dw 2
eight:      dw 8
ten:        dw 10
fifteen:    dw 15
sixteen:    dw 16
val_128:    dw 128
val_b0:     dw 0xB0
val_ff:     dw 0xFF