RNG_ADDR    = 10

//
// Generate a random word into 'result', in the range from [0, range).
//		
Random(result word, range word):
    var i word
    var j word
    cpy i, RNG_ADDR             // get a random number in range 0-65535
    cpy j, i                    // value / range * range
    div j, range
    mul j, range
    cpy result, i
    sec
    sub result, j
    ret

