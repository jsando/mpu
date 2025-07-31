Stdio = 0x0100
StdioGetRandom = Stdio | 0x02

//
// Generate a random word into 'result', in the range from [0, range).
//		
Random(result word, range word):
    var i word
    var j word
    req #GetRandomReq
    cpy i, RandomNumber         // get a random number in range 0-65535
    cpy j, i                    // value / range * range
    div j, range
    mul j, range
    cpy result, i
    sec
    sub result, j
    ret

.GetRandomReq:
    dw 0x0102           // StdioGetRandom
    dw 0                // Errcode
.RandomNumber:
    dw 0                // Random

