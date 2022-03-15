// After seeing how horrible the pong game is using integer dx/dy
// this was a quick test at using 12.4 fixed point values to represent
// dx/dy as vectors with a separate speed, and normalizing
// the vector.

            import "random"
            import "stdio"
            import "sqrt"

            org 0
REG_PC:     dw main

REG_IO_REQ  =   6

            org 0x10
SCREEN_WIDTH    = 640
SCREEN_HEIGHT   = 400
BALL_RADIUS     = 20
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

ball_x:             dw 0
ball_y:             dw 0
ball_fx:            dw 0        // fixed point x position
ball_fy:            dw 0        // fixed point y position
ball_dx:            dw 0        // fixed point x direction [-16,16]
ball_dy:            dw 0        // fixed point y direction [-16,16]
ball_speed:         dw 3

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
.title      db "BALL", 0

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
            cmp keycode, #SDLK_ESCAPE
            jne done
            cpy quit, #1
.done            
            ret

//
// Handle keyup events.
//
onKeyUp(keycode word):
            ret

//
// Redraw the screen.
//
DrawScreen():
            // Clear the screen to black
            cpy REG_IO_REQ, #color
            cpy REG_IO_REQ, #clear
            jsr DrawBall
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
// InitBall
//
InitBall():
    var isLeft word
            // Start in center of screen
            cpy ball_x, #SCREEN_WIDTH / 2
            cpy ball_y, #SCREEN_HEIGHT / 2
            cpy ball_fx, ball_x
            mul ball_fx, #16
            cpy ball_fy, ball_y
            mul ball_fy, #16

            // Choose random dx between -16,16
            psh #0
            psh #32
            jsr Random
            pop #2
            pop ball_dx
            sec
            sub ball_dx, #16

            // Choose random dy between -16,16
            psh #0
            psh #32
            jsr Random
            pop #2
            pop ball_dy
            sec
            sub ball_dy, #16

            ret

// Normalize the ball dx/dy vector to be 1
//
//  length = sqrt((x * x) + (y * y));
//  if length != 0 {
//      length = 1 / length
//      x *= length
//      y *= length
//  }
BallVectNormalize():
            var t1 word
            var t2 word
            var length word

            cpy t1, ball_dx
            mul t1, t1
            div t1, #16         // move decimal point back

            cpy t2, ball_dy
            mul t2, t2
            div t2, #16

            add t1, t2
            psh #0              // sqrt result
            mul t1, #16
            psh t1
            jsr sqrt
            pop #2
            pop length
            jeq done

            cpy t1, #0b_1_0000
            div t1, length
            mul t1, #16
            cpy length, t1
            mul ball_dx, length
            mul ball_dy, length
.done
            ret

//
// Draw ball.
//
DrawBall():
            var t1 word

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
            mul ball_dy, #-1
.y_no_bounce
            // If ball off screen, re-initialize
            cmp ball_x, #-BALL_RADIUS
            jlt x_bounce
            cmp ball_x, #SCREEN_WIDTH
            jlt no_reset
.x_bounce
            mul ball_dx, #-1
.no_reset            
            // ball_x += ball_dx * ball_speed
            cpy t1, ball_dx
            mul t1, ball_speed            
            clc
            add ball_fx, t1
            cpy ball_x, ball_fx
            div ball_x, #16

            // ball_y += ball_dy * ball_speed
            cpy t1, ball_dy
            mul t1, ball_speed            
            clc
            add ball_fy, t1
            cpy ball_y, ball_fy
            div ball_y, #16

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

