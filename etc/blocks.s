/*
    game board is a 12x25 grid
    pieces are on a 4x4 grid
    window is 800x600

    let's say window height = wh, leaving some margin at top and bottom and room
    for a border around the game board.

    cellSize = (window_height - padding) / 25
    board_x = (window_width / 2) - 6 * cellSize
    board_y = top padding

*/
            import "random"
            import "lcd"
            import "stdio"
            import "strconv"

            org 0
            dw Main

            org 0x10
REG_IO_REQ  = 6

            org 0x10
// Constants
LCD_CHAR_WIDTH  = 20
LCD_CHAR_SPACE  = 24
LCD_LINE_SPACE  = 48

SCREEN_WIDTH    = 800
SCREEN_HEIGHT   = 600
SDL_QUIT        = 0x100
SDL_KEYDOWN     = 0x300
SDL_KEYUP       = 0x301
SDLK_ESCAPE     = 0x1b
SDLK_SPACE      = 0x20
SDLK_i          = 0x69
SDLK_j          = 0x6a
SDLK_k          = 0x6b
SDLK_l          = 0x6c

PADDING         = 25
BOARD_HEIGHT    = SCREEN_HEIGHT - 2*PADDING
CELL_SIZE       = BOARD_HEIGHT /25
BOARD_X         = (SCREEN_WIDTH / 2 - 6 * CELL_SIZE)
BOARD_Y         = PADDING
BOARD_WIDTH     = 12 * CELL_SIZE

// Globals
QuitFlag:       dw 0

// Key Status Table
// OnKeyDown/OnKeyUp scan this table of keys we are interested in.
// Its a keycode, followed by a word for it to write 1 if down, 0 if not.
// Table ends with an extra zero.
KeyTable:       dw SDLK_ESCAPE
KeyEscDown:     dw 0
                dw SDLK_SPACE
KeySpaceDown:   dw 0
                dw SDLK_i
KeyIDown:       dw 0
                dw SDLK_j
KeyJDown:       dw 0
                dw SDLK_k
KeyKDown:       dw 0
                dw SDLK_l
KeyLDown:       dw 0
                dw 0    // end of list

// Common color commands ... copy these addresses to REG_IO_REQ to set color
ColorBlack:     dw 0x0205
                db 0,0,0,255
ColorBlue:      dw 0x0205
                db 0,0,170,255
ColorGreen:     dw 0x0205
                db 0,170,0,255
ColorCyan:      dw 0x0205
                db 0,170,170,255                
ColorRed:       dw 0x0205
                db 170,0,0,255
ColorPink:      dw 0x0205
                db 170,0,170,255                            
ColorOrange:    dw 0x0205
                db 170,85,0,255
ColorWhite:     dw 0x0205
                db 255,255,255,255

BlockColors:    dw ColorBlack
                dw ColorBlue
                dw ColorGreen
                dw ColorCyan
                dw ColorRed
                dw ColorPink
                dw ColorOrange

Score:          dw 0            
Lines:          dw 0
Level:          dw 1
Piece:          dw 255
NextPiece1:     dw 0
NextPiece2:     dw 0
NextPiece3:     dw 0

// Game board ... 12x24 color # of each cell, but left/bottom/right are solid and not drawn
GameBoard:      ds 12*25*2

Main():
            jsr InitScreen
            jsr NewGame
.loop
            jsr PollEvents
            jsr DrawScreen
            cmp QuitFlag, #0
            jeq loop
            hlt

InitScreen():
            cpy REG_IO_REQ, #init
            ret

.init       dw 0x0201
            dw SCREEN_WIDTH
            dw SCREEN_HEIGHT
            dw title
.title      db "MPU Blocks", 0

PollEvents():
.loop
            cpy REG_IO_REQ, #poll
            cmp poll_event, #SDL_QUIT
            jne isKeyDown
            cpy QuitFlag, #1
            jmp exit
.isKeyDown
            cmp poll_event, #SDL_KEYDOWN
            jne isKeyUp
            psh keycode
            jsr OnKeyDown
            pop #2
            jmp loop
.isKeyUp
            cmp poll_event, #SDL_KEYUP
            jne isNoMore
            psh keycode
            jsr OnKeyUp
            pop #2
            jmp loop
.isNoMore
            cmp poll_event, #0
            jne loop

            cmp KeyEscDown, #1
            jne exit
            cpy QuitFlag, #1
.exit            
            ret

.poll       dw 0x0202           
.poll_event dw 0                
.poll_time  dw 0                
.keycode
.poll_data  ds 8                

OnKeyDown(keycode word):
            .tabptr local word

            cpy tabptr, #KeyTable
.loop            
            cmp *tabptr, #0
            jeq done
            cmp *tabptr, keycode
            jeq keyFound
            add tabptr, #4
            jmp loop
.keyFound
            add tabptr, #2
            cpy *tabptr, #1
.done
            ret

OnKeyUp(keycode word):
            .tabptr local word

            cpy tabptr, #KeyTable
.loop            
            cmp *tabptr, #0
            jeq done
            cmp *tabptr, keycode
            jeq keyFound
            add tabptr, #4
            jmp loop
.keyFound
            add tabptr, #2
            cpy *tabptr, #0
.done
            ret

DrawScreen():
            // Clear the screen to black
            cpy REG_IO_REQ, #ColorBlack
            cpy REG_IO_REQ, #clear

            jsr DrawBoard

            cpy REG_IO_REQ, #ColorWhite
            cpy REG_IO_REQ, #boardRect

            cpy REG_IO_REQ, #present
            ret

            // device request to clear screen
.clear      dw 0x0204

            // device request to present backbuffer to screen
.present    dw 0x0203
            dw 10               // delay ms

.boardRect  dw 0x0207
            dw BOARD_X + CELL_SIZE / 2
            dw BOARD_Y + CELL_SIZE / 2
            dw BOARD_WIDTH - CELL_SIZE - 1
            dw BOARD_HEIGHT - CELL_SIZE - 1

/*
    for i=1 to 10
        for j=4 to 23
            x = i*8 + cx
            y = j*8 + cy
            Display.DrawRect (x, y, x+6, y+6, 1, game[i,j])
        next
    next
*/
DrawBoard():
            .i local word
            .j local word
            .x local word
            .y local word
            .boardPtr local word
            .color local word

            cpy color, #0
            cpy i, #1
.iLoop
            cpy j, #4
.jLoop
            cpy x, i
            mul x, #CELL_SIZE
            add x, #BOARD_X
            cpy fillX, x

            cpy y, j
            mul y, #CELL_SIZE
            add y, #BOARD_Y
            cpy fillY, y

            cpy boardPtr, j
            mul boardPtr, #12
            add boardPtr, i
            add boardPtr, #GameBoard
            seb
            cpy color, *boardPtr
            clb

            //psh color
            //jsr PrintHex
            //pop #2
            //jsr Println

            mul color, #2
            add color, #BlockColors
            cpy REG_IO_REQ, *color

//            cpy REG_IO_REQ, #ColorRed            
            cpy REG_IO_REQ, #fillRect

            // inner loop
            inc j
            cmp j, #24
            jlt jLoop

            // outer loop
            inc i
            cmp i, #11
            jlt iLoop
            ret

.fillRect   dw 0x0208
.fillX      dw 0
.fillY      dw 0
            dw CELL_SIZE-2
            dw CELL_SIZE-2

ResetBoard:
            db 1,0,0,0,0,0,0,0,0,0,0,1
            db 1,0,0,0,0,0,0,0,0,0,0,1
            db 1,0,0,0,0,0,0,0,0,0,0,1
            db 1,0,0,0,0,0,0,0,0,0,0,1
            db 1,0,0,0,0,0,0,0,0,0,0,1
            db 1,0,0,0,0,0,0,0,0,0,0,1
            db 1,0,0,0,0,0,0,0,0,0,0,1
            db 1,0,0,0,0,0,0,0,0,0,0,1
            db 1,0,0,0,0,0,0,0,0,0,0,1
            db 1,0,0,0,0,0,0,0,0,0,0,1
            db 1,0,0,0,0,0,0,0,0,0,0,1
            db 1,0,0,0,0,0,0,0,0,0,0,1
            db 1,0,0,0,0,0,0,0,0,0,0,1
            db 1,0,0,0,0,0,0,0,0,0,0,1
            db 1,0,0,0,0,0,0,0,0,0,0,1
            db 1,0,0,0,0,0,0,0,0,0,0,1
            db 1,0,0,0,0,0,0,0,0,0,0,1
            db 1,0,0,0,0,0,0,0,0,0,0,1
            db 1,0,0,0,0,0,0,0,0,0,0,1
            db 1,0,0,0,0,0,0,0,0,0,0,1
            db 1,0,0,0,0,0,0,0,0,0,0,1
            db 1,0,0,0,0,0,0,0,0,0,0,1
            db 1,0,0,0,0,0,0,0,0,0,0,1
            db 1,0,0,0,0,0,0,0,0,0,0,1
            db 1,1,1,1,1,1,1,1,1,1,1,1
ResetEnd:
            db 0

NewGame():
            .from local word
            .to   local word

            // Clear the game board
            cpy from, #ResetBoard
            cpy to, #GameBoard
.initBoardLoop    
            cpy *to, *from
            add to, #2
            add from, #2
            cmp from, #ResetEnd
            jlt initBoardLoop
            
            cpy Score, #0
            cpy Piece, #255
            cpy Lines, #0
            cpy Level, #1

            // NextPiece1 = Random(7)
            psh #0
            psh #7
            jsr Random
            pop #2
            pop NextPiece1

            // NextPiece2 = Random(7)
            psh #0
            psh #7
            jsr Random
            pop #2
            pop NextPiece2

            // NextPiece3 = Random(7)
            psh #0
            psh #7
            jsr Random
            pop #2
            pop NextPiece3
            ret
