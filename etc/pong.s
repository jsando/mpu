//-------------------------------------
// MPU PONG
//-------------------------------------

// The initial program counter needs to point to the entry point at startup.
// The rest of the special mem area ($00-$0f) can be left zero on startup.
            org 0
REG_PC:     dw main
REG_SP:     dw 0
REG_FP:     dw 0
REG_IO_REQ: dw 0
REG_IO_RES: dw 0
REG_RAND:   dw 0

            org 0x10
// Constants
SCREEN_WIDTH    = 624
SCREEN_HEIGHT   = 352
BALL_RADIUS     = 20
PADDLE_WIDTH    = 20
PADDLE_HEIGHT   = 80
SDL_QUIT        = 0x100
SDL_KEYDOWN     = 0x300
SDL_KEYUP       = 0x301
SDLK_a          = 0x61
SDLK_z          = 0x7a
SDLK_l          = 0x6c
SDLK_COMMA      = 0x2c
SDLK_ESCAPE     = 0x1b

// Globals
quit:               dw 0

player1_score:      dw 0
player1_paddle_x:   dw 26
player1_paddle_y:   dw SCREEN_HEIGHT / 2
player1_paddle_w:   dw PADDLE_WIDTH
player1_paddle_h:   dw PADDLE_HEIGHT
player1_paddle_up:  dw 0
player1_paddle_dn:  dw 0

player2_score:      dw 0
player2_paddle_x:   dw SCREEN_WIDTH - 48
player2_paddle_y:   dw SCREEN_HEIGHT / 2
player2_paddle_w:   dw PADDLE_WIDTH
player2_paddle_h:   dw PADDLE_HEIGHT
player2_paddle_up:  dw 0
player2_paddle_dn:  dw 0

ball_x:             dw 0
ball_y:             dw 0
ball_xspeed:        dw 0
ball_yspeed:        dw 0

//
// main is the main entry point.
//
main():
            // Open the main window
            jsr InitScreen
            jsr InitBall
.loop
            jsr PollEvents
            jsr DrawScreen
            cmp quit, #0
            jeq loop
.exit        
            // Main doesn't return, it just halts.
            hlt

//
// Initialize the display.
//
InitScreen():
            cpy REG_IO_REQ, #init
            ret

.init       dw 0x0201
            dw SCREEN_WIDTH
            dw SCREEN_HEIGHT
            dw title
.title      db "MPU PONG", 0

//
// Poll and handle all pending graphics events, return 1 if time to exit.
//
PollEvents():
.loop
            cpy REG_IO_REQ, #poll
            cmp poll_event, #SDL_QUIT
            jne isKeyDown
            cpy quit, #1
            jmp exit
.isKeyDown
            cmp poll_event, #SDL_KEYDOWN
            jne isKeyUp
            psh keycode
            jsr onKeyDown
            pop #2
            jmp loop
.isKeyUp
            cmp poll_event, #SDL_KEYUP
            jne isNoMore
            psh keycode
            jsr onKeyUp
            pop #2
            jmp loop
.isNoMore
            cmp poll_event, #0
            jne loop
.exit            
            ret

.poll       dw 0x0202           // graphics, poll        
.poll_event dw 0                // space for response event type id
.poll_time  dw 0                // space for response event timestamp (1/4 second since init)
.keycode
.poll_data  ds 8                // space for response, structure depends on event type

//
// Handle keydown events.
//
onKeyDown(keycode word):
            cmp keycode, #SDLK_a
            jne checkZ
            cpy player1_paddle_up, #1
            jmp done
.checkZ
            cmp keycode, #SDLK_z
            jne checkEsc
            cpy player1_paddle_dn, #1
            jmp done            
.checkEsc  
            cmp keycode, #SDLK_ESCAPE
            jne done
            cpy quit, #1
.done            
            ret

//
// Handle keyup events.
//
onKeyUp(keycode word):
            cmp keycode, #SDLK_a
            jne checkZ

            cpy player1_paddle_up, #0
            jmp done
.checkZ
            cmp keycode, #SDLK_z
            jne done

            cpy player1_paddle_dn, #0
.done            
            ret

//
// Redraw the screen.
//
DrawScreen():
            // Clear the screen to black
            cpy REG_IO_REQ, #color
            cpy REG_IO_REQ, #clear

            cpy REG_IO_REQ, #white
            cpy REG_IO_REQ, #line

            jsr DrawPlayer1Paddle
            jsr DrawPlayer2Paddle
            jsr DrawBall
            jsr Player2AI

            // temp to test my drawstring
            cpy tx, #100
            cpy ty, #SCREEN_HEIGHT / 2
            psh #game_over
            jsr DrawString
            pop #2

            // Present what we've drawn and pause 16ms
            cpy REG_IO_REQ, #present
            ret

            // device request to set color
.color      dw 0x0205
.color_r    db 0
.color_g    db 0
.color_b    db 0
.color_a    db 255

            // device request to clear screen
.clear      dw 0x0204

            // device request to present backbuffer to screen
.present    dw 0x0203
            dw 10               // delay ms
.white      dw 0x0205
            db 255,255,255,255
.line       dw 0x0206
            dw SCREEN_WIDTH / 2
            dw 0
            dw SCREEN_WIDTH / 2
            dw SCREEN_HEIGHT

.game_over  db "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ",0

//
// Draw player 1 paddle.
//
DrawPlayer1Paddle():
            // Set draw color to white
            cpy REG_IO_REQ, #color

            // Draw the paddle as a filled rectange
            cpy rect_x, player1_paddle_x
            cpy rect_y, player1_paddle_y
            cpy rect_w, player1_paddle_w
            cpy rect_h, player1_paddle_h
            cpy REG_IO_REQ, #rect

            // Check the up/down flags, move if set (but keep on screen)
            cmp player1_paddle_up, #0
            jeq check_down
            sec
            sub player1_paddle_y, #2
            jge move_done
            cpy player1_paddle_y, #0
            jmp move_done
.check_down
            cmp player1_paddle_dn, #0            
            jeq move_done
            clc
            add player1_paddle_y, #2
            cmp player1_paddle_y, #SCREEN_HEIGHT - 10
            jlt move_done
            cpy player1_paddle_y, #SCREEN_HEIGHT - 10
.move_done                        
            ret

            // device request to set color
.color      dw 0x0205
            db 255,255,255,255

            // device request to fill rectangle
.rect       dw 0x0208
.rect_x     dw 0
.rect_y     dw 0
.rect_w     dw 0
.rect_h     dw 0

//
// Draw player 2 paddle.
//
DrawPlayer2Paddle():
            // Set draw color to white
            cpy REG_IO_REQ, #color

            // Draw the paddle as a filled rectange
            cpy rect_x, player2_paddle_x
            cpy rect_y, player2_paddle_y
            cpy rect_w, player2_paddle_w
            cpy rect_h, player2_paddle_h
            cpy REG_IO_REQ, #rect

            // Check the up/down flags, move if set (but keep on screen)
            cmp player2_paddle_up, #0
            jeq check_down
            sec
            sub player2_paddle_y, #2
            jge move_done
            cpy player2_paddle_y, #0
            jmp move_done
.check_down
            cmp player2_paddle_dn, #0            
            jeq move_done
            clc
            add player2_paddle_y, #2
            cmp player2_paddle_y, #SCREEN_HEIGHT - 10
            jlt move_done
            cpy player2_paddle_y, #SCREEN_HEIGHT - 10
.move_done                        
            ret

            // device request to set color
.color      dw 0x0205
            db 255,255,255,255

            // device request to fill rectangle
.rect       dw 0x0208
.rect_x     dw 0
.rect_y     dw 0
.rect_w     dw 0
.rect_h     dw 0

//
// InitBall
//
InitBall():
    .isLeft local word
            // Start in center of screen
            cpy ball_x, #SCREEN_WIDTH / 2
            cpy ball_y, #SCREEN_HEIGHT / 2

            // Set xspeed to either 3 or 4
            psh #0
            psh #2
            jsr Random
            pop #2
            pop ball_xspeed
            clc
            add ball_xspeed, #3

            // Half the time, have the ball going left instead of right
            psh #0
            psh #100
            jsr Random
            pop #2
            pop isLeft
            and isLeft, #1
            jeq setYSpeed
            mul ball_xspeed, #-1
.setYSpeed
            // Set yspeed between -3...3
            psh #0
            psh #7
            jsr Random
            pop #2
            pop ball_yspeed
            sec
            sub ball_yspeed, #3
            ret

//
// Draw ball.
//
DrawBall():
            // Set draw color to white
            cpy REG_IO_REQ, #color

            // Draw the paddle as a filled rectange
            cpy rect_x, ball_x
            cpy rect_y, ball_y
            cpy rect_w, #BALL_RADIUS
            cpy rect_h, #BALL_RADIUS
            cpy REG_IO_REQ, #rect

            // Bounce if hit top or bottom
            cmp ball_y, #0
            jlt y_bounce
            cmp ball_y, #(SCREEN_HEIGHT - BALL_RADIUS)
            jlt y_no_bounce
.y_bounce
            mul ball_yspeed, #-1
.y_no_bounce
            // If ball off screen, re-initialize
            cmp ball_x, #-BALL_RADIUS
            jlt reset_ball
            cmp ball_x, #SCREEN_WIDTH
            jlt no_reset
.reset_ball
            jsr InitBall
.no_reset            
            // Move ball
            clc
            add ball_x, ball_xspeed
            clc
            add ball_y, ball_yspeed

            jsr BounceBall      // Bounce ball off player paddles

            ret

            // device request to set color
.color      dw 0x0205
            db 0,0,255,255

            // device request to fill rectangle
.rect       dw 0x0208
.rect_x     dw 100
.rect_y     dw 100
.rect_w     dw 50
.rect_h     dw 50

BounceBall():
            .overlap local word

            // See if the ball is hitting player1 paddle
            psh #0              // overlap if != 0
            psh player1_paddle_x
            psh player1_paddle_y
            psh player1_paddle_w
            psh player1_paddle_h
            psh ball_x
            psh ball_y
            psh #BALL_RADIUS
            psh #BALL_RADIUS
            jsr Overlap
            pop #16
            pop overlap
            jeq check_player2
            mul ball_xspeed, #-1
            jmp done
.check_player2
            psh #0              // overlap if != 0
            psh player2_paddle_x
            psh player2_paddle_y
            psh player2_paddle_w
            psh player2_paddle_h
            psh ball_x
            psh ball_y
            psh #BALL_RADIUS
            psh #BALL_RADIUS
            jsr Overlap
            pop #16
            pop overlap
            jeq done
            mul ball_xspeed, #-1
.done
            ret

//
// Player 2 is automatic.
//
Player2AI():
    .middle local word
            cpy middle, player2_paddle_y
            clc
            add middle, #(PADDLE_HEIGHT / 2)
            cmp ball_y, middle
            jlt move_up
            jeq no_move
.move_dn
            cpy player2_paddle_dn, #1
            cpy player2_paddle_up, #0
            ret
.move_up
            cpy player2_paddle_dn, #0
            cpy player2_paddle_up, #1
            ret
.no_move       
            cpy player2_paddle_dn, #0
            cpy player2_paddle_up, #0
            ret

//
// Determine if two rectanges are overlapping.
//
Overlap(overlap word, x1 word, y1 word, w1 word, h1 word, x2 word, y2 word, w2 word, h2 word):
            .right1     local word
            .bottom1    local word
            .right2     local word
            .bottom2    local word

            cpy overlap, #0     // default to no overlap
            clc                 // compute bottom/right edges
            cpy right1, x1
            add right1, w1
            clc
            cpy bottom1, y1
            add bottom1, h1
            clc
            cpy right2, x2
            add right2, w2
            clc
            cpy bottom2, y2
            add bottom2, h2

            cmp y1, bottom2     // r1 completely below r2?
            jge done
            cmp bottom1, y2     // r1 completely above r2?
            jlt done
            cmp x1, right2      // r1 to right of r2?
            jge done
            cmp right1, x2      // r1 tot left of r2?
            jlt done
            cpy overlap, #1     // If its not all of the above, its overlapping
.done
            ret

// Generate a random number in the range (0, range]
//
Random(result word, range word):
    .i local word
    .j local word
            cpy i, 10           // get a random number in range 0-65535
            cpy j, i            // value / range * range
            div j, range
            mul j, range
            cpy result, i
            sec
            sub result, j
            ret

tx:         dw 0                // Text next print x coordinate
ty:         dw 0                // Text next print y coordinate

DrawString(pstring word):
    .ch local word
            cpy ch, #0
.loop
            seb
            cpy ch, *pstring
            jeq done
            clb

            cmp ch, #' '
            jne is_digit
            add tx, #5
            jmp next
.is_digit
            cmp ch, #'0' // skip anything less than '0'
            jlt next
            cmp ch, #'9'+1
            jlt number
            cmp ch, #'A'
            jlt next
            cmp ch, #'Z'+1
            jge next

            sub ch, #'A'
            add ch, #10
            jmp drawchar
.number            
            sub ch, #'0'
.drawchar
            psh ch
            jsr DrawCharacter            
            pop #2
.next
            inc pstring
            jmp loop
.done
            clb
            ret

DrawCharacter(char word):
            .mask local word
            .test local word
            .lcd local word
            .segptr local word

            // Set drawcolor to white
            cpy REG_IO_REQ, #white

            // Lookup which segments to draw for the requested character
            cpy lcd, char
            mul lcd, #2
            add lcd, #CharacterTable
            cpy mask, *lcd
            cpy segptr, #CharacterSegmentTable
.loop
            cpy test, #1
            and test, mask
            jne draw
            add segptr, #8
            jmp next
.draw
            cpy line_x, tx
            add line_x, *segptr
            add segptr, #2

            cpy line_y, ty
            add line_y, *segptr
            add segptr, #2

            cpy line_x2, tx
            add line_x2, *segptr
            add segptr, #2

            cpy line_y2, ty
            add line_y2, *segptr
            add segptr, #2
            cpy REG_IO_REQ, #line
.next
            div mask, #2        // shift right
            jne loop            // if result is zero there's nothing more to draw

            add tx, #CHAR_SPACE
            ret

.white      dw 0x0205
            db 255,255,255,255
.line       dw 0x0206
.line_x     dw 0
.line_y     dw 0
.line_x2    dw 0
.line_y2    dw 0

// A 16-segment lcd font.  Its a 14-segment display with 2 custom segments for K and R.
//
//    aaaaaaaa     h   n    i          l                  
//   b        c     h  n   i         l                    
//   b        c      h n  i        l                     
//   b        c       hn i       l                      
//    dddddddd                                      
//   e        f       jo k       m                   
//   e        f      j o  k        m                 
//   e        f     j  o   k         m                  
//    gggggggg     j   o    k          m                
//
// Mapping to bitmask values:
//  1 = a
//  2 = b
//  4 = c
//  8 = d
// 16 = e
// 32 = f
// 64 = g
// 128 = l
// 256 = m
// 512 = h
// 1024 = n
// 2048 = i
// 4096 = j
// 8192 = o
// 16384 = k

CHAR_SPACE  = 14
CWIDTH      = 12
CLEFT       = 0
CRIGHT      = CWIDTH - 1
CTOP        = 0
CBOTTOM     = 2 * CWIDTH - 1
CMIDY       = CWIDTH
CMIDX       = CWIDTH / 2

CharacterSegmentTable:
    dw CLEFT+1,CTOP,CRIGHT-1,CTOP
    dw CLEFT,CTOP+1,CLEFT,CMIDY-1
    dw CRIGHT,CTOP+1,CRIGHT,CMIDY-1
    dw CLEFT+1,CMIDY,CRIGHT-1,CMIDY
    dw CLEFT,CMIDY+1,CLEFT,CBOTTOM-1
    dw CRIGHT,CMIDY+1,CRIGHT,CBOTTOM-1
    dw CLEFT+1,CBOTTOM,CRIGHT-1,CBOTTOM
    dw CLEFT+1,CMIDY,CRIGHT-1,CTOP
    dw CLEFT+1,CMIDY,CRIGHT-1,CBOTTOM
    dw CLEFT,CTOP,CMIDX,CMIDY
    dw CMIDX,CTOP,CMIDX,CMIDY
    dw CMIDX,CMIDY,CRIGHT,CTOP
    dw CMIDX,CMIDY,CLEFT,CBOTTOM
    dw CMIDX,CMIDY,CMIDX,CBOTTOM
    dw CMIDX,CMIDY,CRIGHT,CBOTTOM


CharacterTable:
    dw 1+2+4+16+32+64           // 0
    dw 4+32                     // 1
    dw 1+4+8+16+64              // 2
    dw 1+4+8+32+64              // 3
    dw 2+4+8+32                 // 4
    dw 1+2+8+32+64              // 5
    dw 2+8+16+32+64             // 6
    dw 1+4+32                   // 7
    dw 1+2+4+8+16+32+64         // 8
    dw 1+2+4+8+32               // 9
    dw 16+2+1+4+32+8            // A
    dw 2+1+4+16+32+64+8         // B
    dw 1+2+16+64                // C
    dw 64+16+8+32+4             // D
    dw 64+16+8+2+1              // E
    dw 16+8+2+1                 // F
    dw 1+2+16+64+32             // G
    dw 2+16+32+4+8              // H
    dw 4+32                     // I
    dw 4+32+64                  // J
    dw 16+256+128+2             // K
    dw 2+16+64                  // L
    dw 16+2+512+2048+4+32       // M
    dw 16+2+512+16384+32+4      // N
    dw 1+2+4+16+32+64           // O
    dw 2+1+4+8+16               // P
    dw 1+2+4+16+32+64           // Q
    dw 16+2+1+4+8+256           // R
    dw 1+2+8+32+64              // S
    dw 8192+1024+1              // T
    dw 16+64+32+2+4             // U
    dw 16+64+32+2+4             // V
    dw 16+2+32+4+4096+16384     // W
    dw 512+16384+2048+4096      // X
    dw 512+2048+8192            // Y
    dw 1+2048+4096+64           // Z

