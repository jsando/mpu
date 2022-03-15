//-------------------------------------
// MPU PONG
//-------------------------------------
// Improvements:
//  Make y vector depend on where it hits paddle (-n if near top, 0 if middle, n if bottom)
//  Ensure dy=0 doesn't happen, otherwise ball just bounces back and forth with no play
//  Add sound
//  Increase velocity as game progresses
//  Use fixed point for ball for more realistic motion / bounce

            import "random"
            import "lcd"
            import "strconv"

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
LCD_CHAR_WIDTH  = 20
LCD_CHAR_SPACE  = 24
LCD_LINE_SPACE  = 48

SCREEN_WIDTH    = 640
SCREEN_HEIGHT   = 400
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
SDLK_SPACE      = 0x20
SDLK_1          = 0x31
SDLK_2          = 0x32

// Globals
quit:               dw 0

player1_score:      dw 23
player1_paddle_x:   dw 26
player1_paddle_y:   dw SCREEN_HEIGHT / 2
player1_paddle_w:   dw PADDLE_WIDTH
player1_paddle_h:   dw PADDLE_HEIGHT
player1_paddle_up:  dw 0
player1_paddle_dn:  dw 0

players:            dw 0

player2_score:      dw 21
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

game_over:          dw 1 // 1 if game over, 0 if playing

game_over_msg:      db "GAME OVER",0
press_space_msg:    db "1=1 PLAYER, 2=2 PLAYERS",0


//
// main is the main entry point.
//
main():
            // Open the main window
            jsr InitScreen
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
            jne checkl
            cpy quit, #1
            jmp done
.checkl
            cmp keycode, #SDLK_l
            jne checkComma
            cpy player2_paddle_up, #1
            jmp done
.checkComma
            cmp keycode, #SDLK_COMMA
            jne check1
            cpy player2_paddle_dn, #1
            jmp done
.check1
            cmp keycode, #SDLK_1
            jne check2
            cpy players, #1
            jsr NewGame
            jmp done
.check2
            cmp keycode, #SDLK_2
            jne done
            cpy players, #2
            jsr NewGame
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
            jne checkl
            cpy player1_paddle_dn, #0
.checkl
            cmp keycode, #SDLK_l
            jne checkComma
            cpy player2_paddle_up, #0
.checkComma
            cmp keycode, #SDLK_COMMA
            jne done
            cpy player2_paddle_dn, #0
.done            
            ret

//
// Start a new game.
//
NewGame():
            // Ignore if already in a game
            cmp game_over, #0
            jeq done

            cpy game_over, #0
            cpy player1_score, #0
            cpy player2_score, #0
            jsr InitBall
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
            jsr DrawScore
            cmp game_over, #0
            jne not_playing

            jsr DrawBall
            cmp players, #2
            jge done
            jsr Player2AI
            jmp done

.not_playing
            // Show GAME OVER
            cpy tx, #215
            cpy ty, #80
            psh #game_over_msg
            jsr DrawString
            pop #2
            cpy tx, #40
            cpy ty, #80+LCD_LINE_SPACE
            psh #press_space_msg
            jsr DrawString
            pop #2

            // Present what we've drawn and pause 16ms
.done            
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
            cmp player1_paddle_y, #SCREEN_HEIGHT - PADDLE_HEIGHT
            jlt move_done
            cpy player1_paddle_y, #SCREEN_HEIGHT - PADDLE_HEIGHT
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
            cmp player2_paddle_y, #SCREEN_HEIGHT - PADDLE_HEIGHT
            jlt move_done
            cpy player2_paddle_y, #SCREEN_HEIGHT - PADDLE_HEIGHT
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
    var isLeft word
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
    var trash word
            // Set draw color to white
            cpy REG_IO_REQ, #color

            // Draw the ball as a filled rectange
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
            jlt player2_scored
            cmp ball_x, #SCREEN_WIDTH
            jlt no_reset

            // player 1 scored
            inc player1_score
            jsr IsGameOver
            cmp game_over, #0
            jne exitGameOver
            jsr InitBall
            jmp no_reset
.player2_scored
            inc player2_score
            jsr IsGameOver
            cmp game_over, #0
            jne exitGameOver
            jsr InitBall                                    
.no_reset            
            // Move ball
            clc
            add ball_x, ball_xspeed
            clc
            add ball_y, ball_yspeed

            jsr BounceBall      // Bounce ball off player paddles
.exitGameOver
            ret

            // device request to set color
.color      dw 0x0205
            db 255,255,255,255

            // device request to fill rectangle
.rect       dw 0x0208
.rect_x     dw 100
.rect_y     dw 100
.rect_w     dw 50
.rect_h     dw 50

//
// Game over when one player hits at least 11 and is +2 over the other player.
//
IsGameOver():
    var t1 word
            cmp player1_score, player2_score
            jlt player2Higher

            cmp player1_score, #11
            jlt done
            cpy t1, player1_score
            sub t1, player2_score
            cmp t1, #2
            jlt done
            jmp over
.player2Higher
            cmp player2_score, #11
            jlt done
            cpy t1, player2_score
            sub t1, player1_score
            cmp t1, #2
            jlt done
.over            
            cpy game_over, #1
.done
            ret

BounceBall():
            var overlap word

            // See if the ball is hitting player1 paddle, but only if its going in that direction
            cmp ball_xspeed, #0
            jge check_player2

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
            // Ball collidding with paddle2?  Only check if the ball is going in that direction.
            cmp ball_xspeed, #0
            jlt done

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
    var middle word
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
            var right1      word
            var bottom1     word
            var right2      word
            var bottom2     word

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

//
// DrawScore
//
DrawScore():
    var score word
            cpy tx, #SCREEN_WIDTH/4
            cpy ty, #5
            psh player1_score
            jsr PrintInteger
            pop #2

            cpy tx, #SCREEN_WIDTH*3/4
            cpy ty, #5
            psh player2_score
            jsr PrintInteger
            pop #2

            ret

//
// Print the word passed on the stack to stdout in decimal.
//
PrintInteger(value word):
        psh value
        psh #buffer
        psh #10                 // buffer length = 10
        jsr Itoa
        pop #2
        jsr DrawString
        pop #2
        ret

.buffer ds 12       // max 10 digits + null ... plus 2 extra cuz I got a bug somewhere's
