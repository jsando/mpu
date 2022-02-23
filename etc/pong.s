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
SCREEN_WIDTH    = 640
SCREEN_HEIGHT   = 480
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
player1_paddle_w:   dw 20
player1_paddle_h:   dw 80
player1_paddle_up:  dw 0
player1_paddle_dn:  dw 0

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

.init       dw 2,1              // graphics, initialize
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

.poll       dw 2,2              // graphics, poll        
.poll_event dw 0                // space for response event type id
.poll_time  dw 0                // space for response event timestamp (1/4 second since init)
.keycode
.poll_data  ds 8                // space for response, structure depends on event type

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

            jsr DrawPlayer1Paddle

            // Present what we've drawn and pause 16ms
            cpy REG_IO_REQ, #present
            ret

            // device request to set color
.color      dw 2,5
.color_r    db 0
.color_g    db 0
.color_b    db 0
.color_a    db 255

            // device request to clear screen
.clear
            dw 2,4              // graphics, clear

            // device request to present backbuffer to screen
.present
            dw 2,3              // graphics, present
            dw 16               // delay ms

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
.color      dw 2,5
            db 255,255,255,255

            // device request to fill rectangle
.rect       dw 2,8
.rect_x     dw 0
.rect_y     dw 0
.rect_w     dw 0
.rect_h     dw 0
