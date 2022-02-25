'-----------------------------------------------------------------------------
' Bobtris
'-----------------------------------------------------------------------------

const int FRAME_MILLIS = 10         ' Animation loop rate

int bloc[7, 4]
int bonus
int cx = 80
int cy = 20
int dx
int dy
int game[12,25]
int lcd[46]
int level
int line_score[5]
int lines
int mult
int next_piece1
int next_piece2
int next_piece3
int oldrot
int oldx
int oldy
int ox
int oy
int piece
int px, py
int rot
int score
int space_wait
int temp_piece
int temp_rot
int tx
int ty
int valid
int x
int y
long current_time
long fall_delay
long start_time
long loop_time = System.Time ()

Sound.Initialize ()
Display.SetMode (0)
InitLCDArray ()
InitLineScore ()
InitGameArray ()
InitBlockArray ()

' Main loop start
loop
    Display.ClearScreen ()
    DrawDigits (50, 160, level)
    DrawDigits (50, 170, lines)
    DrawDigits (50, 180, score)
    tx = 20: ty = 160: DrawDigit (21): DrawDigit (14): DrawDigit (31): DrawDigit (14): DrawDigit (21)
    tx = 20: ty = 170: DrawDigit (21): DrawDigit (18): DrawDigit (23): DrawDigit (14): DrawDigit (28)
    tx = 20: ty = 180: DrawDigit (28): DrawDigit (12): DrawDigit (24): DrawDigit (27): DrawDigit (14)

    for i=1 to 10
        for j=4 to 23
            x = i*8 + cx
            y = j*8 + cy
            Display.DrawRect (x, y, x+6, y+6, 1, game[i,j])
        next
    next

    if piece = 255 then CreatePiece ()
    if piece = 255 then
        ' Gameover mode. Wait for a key to restart
        Display.DrawRect (0,0,255,223,0,15)
        if Keyboard.KeyDown(Keyboard.VK_SPACE) then InitGameArray ()
    else
        if dx > (px / 256 * 256) then dx = dx - 32
        if dx < (px / 256 * 256) then dx = dx + 32
        x = ((dx*8) / 256) + cx : y = (py*8 / 256) + cy - 7 : temp_piece = piece : temp_rot = rot : DrawBlock ()
        x = 9*8 + cx+3 : y = cy-2 : temp_piece = next_piece1 : temp_rot = 0 : DrawBlock ()
        x = 13*8 + cx+3 : y = cy-2 : temp_piece = next_piece2 : temp_rot = 0 : DrawBlock ()
        x = 17*8 + cx+3 : y = cy-2 : temp_piece = next_piece3 : temp_rot = 0 : DrawBlock ()

        oldx = px : oldy = py : oldrot = rot
        current_time = System.Time ()
        py = py + 2 + 3 * level

        if Keyboard.KeyDown(Keyboard.VK_LEFT) <> 0 and px > 32 then px = px - 32
        if Keyboard.KeyDown(Keyboard.VK_RIGHT) then px = px + 32
        if Keyboard.KeyDown(Keyboard.VK_DOWN) then py = py + 54

        if Keyboard.KeyDown(Keyboard.VK_SPACE) <> 0 and space_wait = 0 then space_wait = 1 : if rot = 3 then rot = 0 else rot = rot + 1
        if Keyboard.KeyDown(Keyboard.VK_SPACE) = 0 then space_wait = 0

        ' Evaluate current position validity
        loop
            if IsValid() then break

            ' reject the last move or stamp the block at the last valid position (if only the y has moved)
            if px = oldx and rot = oldrot and oldy <> py then 
                py = oldy
                StampPiece ()
                break
            end
            if px <> oldx and rot <> oldrot then 
                px = oldx
            else
                if rot <> oldrot then Sound.PlaySound (3,0)                    
                px = oldx
                rot = oldrot
                if oldy = py then break
            end
        end
    end

    ' rect cx+8,cy,cx+11*8-1,cy-1+4*8,1,0
    Display.DrawRect (cx-1+8,cy-1+4*8,cx+11*8-1,cy+24*8-1,0,15)
    Display.ShowPage ()
    if Keyboard.KeyDown(Keyboard.VK_ESCAPE) then break

    ' frame timer (slow it down on fast systems)
    long now = System.Time ()
    int diff = now - loop_time
    if diff < FRAME_MILLIS then
        long sleep = FRAME_MILLIS - diff
        System.Sleep (sleep)
        now = System.Time ()
    end
    loop_time = now

end ' main loop

function IsValid () boolean
    mult = 1
    ox = px / 256
    oy = py / 256
    for i = 1 to 14
        if (bloc[piece,rot] and mult) <> 0 then
            if game[ox,oy] <> 0 then 
                return false
            end
        end
        if i = 4 or i = 8 or i = 12 then 
            ox = ox - 3
            oy = oy + 1 
        else 
            ox = ox + 1
        end
        mult = mult * 2
    next
    return true
end

' Generic number display routine
function DrawDigits (int nx, int ny, long text_value)
    tx = nx: ty = ny
    Display.SetColor (15)
    long divisor = 100000
    int first_number = 0
    loop
        int temp = text_value / divisor
        text_value = text_value - temp * divisor
        if temp = 0 and divisor <> 1 and first_number = 0 then
            tx = tx + 5
        else
            DrawDigit (temp)
            first_number = 1
        end
        if divisor = 1 then break
        divisor = divisor / 10
    end
end

function DrawDigit (int digit)
    int dra = lcd[digit]
    if dra and 1 then  Display.DrawLine (tx+1,ty+0,tx+2,ty+0)
    if dra and 2 then  Display.DrawLine (tx+0,ty+1,tx+0,ty+2)
    if dra and 4 then  Display.DrawLine (tx+3,ty+1,tx+3,ty+2)
    if dra and 8 then  Display.DrawLine (tx+1,ty+3,tx+2,ty+3)
    if dra and 16 then Display.DrawLine (tx+0,ty+4,tx+0,ty+5)
    if dra and 32 then Display.DrawLine (tx+3,ty+4,tx+3,ty+5)
    if dra and 64 then Display.DrawLine (tx+1,ty+6,tx+2,ty+6)
    if dra and 128 then Display.DrawLine (tx+1,ty+2,tx+2,ty+1)
    if dra and 256 then Display.DrawLine (tx+1,ty+4,tx+2,ty+5)
    if dra and 512 then Display.DrawLine (tx+0,ty+1,tx+1,ty+2)
    if dra and 1024 then Display.DrawLine (tx+1,ty+1,tx+1,ty+3)
    if dra and 2048 then Display.DrawLine (tx+2,ty+2,tx+3,ty+1)
    if dra and 4096 then Display.DrawLine (tx+0,ty+5,tx+1,ty+4)
    if dra and 8192 then Display.DrawLine (tx+1,ty+3,tx+1,ty+5)
    if dra and 16384 then Display.DrawLine (tx+2,ty+4,tx+3,ty+5)
    tx = tx + 5
end

function InitBlockArray ()
' Initialise static block array. bloc(# piece number, rotation from 0 to 3)

    bloc[0,0] = 2+32+512+1024
    bloc[0,1] = 16+32+64+256
    bloc[0,2] = 1+2+32+512
    bloc[0,3] = 4+16+32+64

    bloc[1,0] = 2+32+256+512
    bloc[1,1] = 1+16+32+64
    bloc[1,2] = 2+4+32+512
    bloc[1,3] = 16+32+64+1024

    bloc[2,0] = 16+32+64+512
    bloc[2,1] = 2+32+64+512
    bloc[2,2] = 2+16+32+64
    bloc[2,3] = 2+16+32+512

    bloc[3,0] = 32+64+512+1024
    bloc[3,1] = 32+64+512+1024
    bloc[3,2] = 32+64+512+1024
    bloc[3,3] = 32+64+512+1024

    bloc[4,0] = 2+32+64+1024
    bloc[4,1] = 2+4+16+32
    bloc[4,2] = 2+32+64+1024
    bloc[4,3] = 2+4+16+32

    bloc[5,0] = 16+32+512+1024
    bloc[5,1] = 4+64+32+512
    bloc[5,2] = 16+32+512+1024
    bloc[5,3] = 4+64+32+512

    bloc[6,0] = 2+32+512+8192
    bloc[6,1] = 16+32+64+128
    bloc[6,2] = 2+32+512+8192
    bloc[6,3] = 16+32+64+128
end

function StampPiece ()
    ' Stamp the current piece in the game array
    mult = 1
    ox = px / 256
    oy = py / 256
    for i = 1 to 14
        if (bloc[piece,rot] and mult) <> 0 then game[ox,oy] = piece+1
        if i = 4 or i = 8 or i = 12 then ox = ox - 3 : oy = oy + 1 else ox = ox + 1
        mult = mult * 2
    next
    current_time = System.Time () : piece = 255
    bonus = 30 - (current_time-start_time) / 60
    if bonus > 0 then score = score + bonus
    Sound.PlaySound (3,0)

    oy = py / 256
    int total_line = 0
    for j = oy to oy + 3
        if j = 24 then break
        int count = 0
        for i = 1 to 10
            if game[i,j] <> 0 then count = count + 1
        next i
        if count = 10 then
            total_line = total_line + 1
            for j2 = j-1 to 2 step -1
                for i2 = 1 to 10
                    game[i2,j2+1] = game[i2,j2]
                next
            next
        end
    next

    ' Add score according to number of line, at maybe a flashing visual effect.
    score = score + line_score[total_line]
    if total_line >= 1 and total_line <= 3 then Sound.PlaySound (4,0)
    if total_line = 4 then Sound.PlaySound (5,0)
    lines = lines + total_line
    if lines >= level*15 then 
        level = level + 1
        Sound.PlaySound (11,0)
    end
end

' Build the game array
function InitGameArray ()
    for i = 0 to 11
        for j = 0 to 24
            game[i, j] = 0
        end
    end
    for i = 0 to 11
        game[i, 24] = 1
    end
    for j = 0 to 24
        game[0, j] = 1
        game[11, j] = 1
    end
    score = 0
    piece = 255
    lines = 0
    level = 1
    next_piece1 = System.Random(0,6)
    next_piece2 = System.Random(0,6)
    next_piece3 = System.Random (0,6)
end

' initialise the 'lcd' array to contain a font
function InitLCDArray ()
    lcd[0] = 1+2+4+16+32+64 : lcd[1] = 4+32 : lcd[2] = 1+4+8+16+64 : lcd[3] = 1+4+8+32+64 : lcd[4] = 2+4+8+32
    lcd[5] = 1+2+8+32+64 : lcd[6] = 2+8+16+32+64 : lcd[7] = 1+4+32 : lcd[8] = 1+2+4+8+16+32+64 : lcd[9] = 1+2+4+8+32
    lcd[10] = 16+2+1+4+32+8 'a
    lcd[11] = 2+1+4+16+32+64+8 'b
    lcd[12] = 1+2+16+64 'c
    lcd[13] = 64+16+8+32+4 'd
    lcd[14] = 64+16+8+2+1 'e
    lcd[15] = 16+8+2+1 'f
    lcd[16] = 1+2+16+64+32 'g
    lcd[17] = 2+16+32+4+8 'h
    lcd[18] = 4+32 'i
    lcd[19] = 4+32+64 'j
    lcd[20] = 16+256+128+2 'k
    lcd[21] = 2+16+64 'L
    lcd[22] = 16+2+512+2048+4+32 'M
    lcd[23] = 16+2+512+16384+32+4 'N
    lcd[24] = 1+2+4+16+32+64 'O
    lcd[25] = 2+1+4+8+16 'P
    lcd[26] = 1+2+4+16+32+64 'Q
    lcd[27] = 16+2+1+4+8+256 'R
    lcd[28] = 1+2+8+32+64 'S
    lcd[29] = 8192+1024+1 'T
    lcd[30] = 16+64+32+2+4 'U
    lcd[31] = 16+64+32+2+4 'V
    lcd[32] = 16+2+32+4+4096+16384 'W
    lcd[33] = 512+16384+2048+4096 'X
    lcd[34] = 512+2048+8192 'Y
    lcd[35] = 1+2048+4096+64 'Z
end

' Initialise score obtain when clearing lines
function InitLineScore ()
    line_score[0] = 0
    line_score[1] = 100
    line_score[2] = 200
    line_score[3] = 500
    line_score[4] = 2000
end

' Create a new piece, and test if it is valid
function CreatePiece ()
    piece = next_piece1
    next_piece1 = next_piece2
    next_piece2 = next_piece3
    next_piece3 = System.Random(0,6)
    rot = 0
    px = 4 * 256
    dx = px
    py = 2 * 256
    fall_delay = System.Time ()
    start_time = System.Time ()
    IsValid() ' test if block can spawn or not
    if valid = 1 then return ' start the game normally
    piece = 255 ' cancel block, gameover.
end

' Draw a block on the screen at pos 'x','y', piece # 'temp_piece', rotation # 'temp_rot'
function DrawBlock ()
    local i

    lda #1
    sta mult
    lda x
    sta ox
    lda y
    sta oy

    lda #1
    sta i
loop:
    lda i
    cmp #15
    bge endloop

    ldx temp_piece
    ldy temp_rot
    lda bloc[x,y]
    beq endif1
    lda mult
    beq endif1

    'call Display.DrawRect (ox,oy,ox+6,oy+6,1,temp_piece+1)
    lda ox
    pha
    lda oy
    pha
    clc
    lda ox
    adc #6
    pha
    clc
    lda oy
    adc #6
    pha
    lda #1
    pha
    clc
    lda temp_piece
    adc #1
    jsr Display.DrawRect
endif1:
    lda i
    cmp #4
    beq doit2
    cmp #8
    beq doit2
    cmp #12
    bne else2
doit2:
    clc
    lda ox
    sbc #24
    sta ox
    lda oy
    adc 8
    sta oy
    jmp endif2
else2:
    lda ox
    adc #8
    sta ox

endif2:
    asl mult
    inc i
    jmp loop
endloop:
    return
end

' Draw a block on the screen at pos 'x','y', piece # 'temp_piece', rotation # 'temp_rot'
function DrawBlock()
    mult = 1
    ox = x
    oy = y
    for i = 1 to 14
        if bloc[temp_piece,temp_rot] and mult then
            Display.DrawRect (ox,oy,ox+6,oy+6,1,temp_piece+1)
        end
        if i = 4 or i = 8 or i = 12 then
            ox = ox - 24
            oy = oy + 8 
        else
            ox = ox + 8
        end
        mult = mult * 2
    end
end

