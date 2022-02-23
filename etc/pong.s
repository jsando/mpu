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

// Globals

// main is the main entry point.
main():
    .pollQuit local word
            // Open the main window
            jsr InitScreen
.loop
            // Process any mouse or keyboard events, returns != 0 if should exit
            psh #0              // push space for return value
            jsr PollEvents
            pop pollQuit        // pop the value to set the zero flag
            jne exit            // if not zero, exit

            // Redraw the whole screen
            jsr DrawScreen
            jmp loop
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
PollEvents(quit word):
            cpy quit, #0
.loop
            cpy REG_IO_REQ, #poll
            cmp poll_event, #SDL_QUIT
            jne next
            cpy quit, #1
            jne exit
.next            
            // TODO: handle additional event types here
            cmp poll_event, #0
            jne loop
.exit            
            ret

.poll       dw 2,2              // graphics, poll        
.poll_event dw 0                // space for response event type id
.poll_time  dw 0                // space for response event timestamp (1/4 second since init)

//
// Redraw the screen.
//
DrawScreen():
            // Clear the screen to black
            cpy REG_IO_REQ, #color
            cpy REG_IO_REQ, #clear

            // TODO: draw more stuff here!

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

