// Test driver for lcd

            include "lcd.s"

            dw main

// Define font size, these are referenced from lcd.s
LCD_CHAR_WIDTH  = 16
LCD_CHAR_SPACE  = 20
LCD_LINE_SPACE  = 40

// Constants
REG_IO_REQ      = 6
SCREEN_WIDTH    = 800
SCREEN_HEIGHT   = 600
SDL_QUIT        = 0x100
SDL_KEYDOWN     = 0x300
SDL_KEYUP       = 0x301
SDLK_ESCAPE     = 0x1b 

// Globals
                    org 0x10

quit:               dw 0

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

.init       dw 0x0201              // graphics, initialize
            dw SCREEN_WIDTH
            dw SCREEN_HEIGHT
            dw title
.title      db "MPU LCD", 0

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

            // temp to test my drawstring
            cpy REG_IO_REQ, #white
            cpy tx, #0
            cpy ty, #0
            psh #sample
            jsr DrawString
            pop #2
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
            db 255,0,0,255
.line       dw 0x0206
            dw SCREEN_WIDTH / 2
            dw 0
            dw SCREEN_WIDTH / 2
            dw SCREEN_HEIGHT

.sample  db " !\"#$%&'()*+,-./",10
         db "0123456789:;<=>?@",10
         db "ABCDEFGHIJKLMNOPQRSTUVWXYZ",10
         db "[\\]^_`",10
         db "abcdefghijklmnopqrstuvwxyz",10
         db "{|}~",0

