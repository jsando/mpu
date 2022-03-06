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
LCD_CHAR_WIDTH  = 16
LCD_CHAR_SPACE  = 20
LCD_LINE_SPACE  = 40

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
NEXT_X          = BOARD_X + BOARD_WIDTH
NEXT_Y1         = PADDING + CELL_SIZE
NEXT_Y2         = NEXT_Y1 + 4 * CELL_SIZE
NEXT_Y3         = NEXT_Y2 + 4 * CELL_SIZE
COORD_SCALE     = 64

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
ColorBlockGrey: dw 0x0205
                db 170,170,170,255                
ColorWhite:     dw 0x0205
                db 255,255,255,255
ColorGrey:      dw 0x0205
                db 50,50,50,255
ColorScoreBG:   dw 0x0205
                db 0,0,50,255
ColorBG:        dw 0x0205
                db 0,0,80,255
ColorScoreFG:   dw 0x0205
                db 0,255,0,255

BlockColors:    dw ColorBlack
                dw ColorBlue
                dw ColorGreen
                dw ColorCyan
                dw ColorRed
                dw ColorPink
                dw ColorOrange
                dw ColorBlockGrey

Score:          dw 0            
Lines:          dw 0
Level:          dw 1

Piece:          dw 255          // Current piece number 0-6, 255 = game over
PieceX:         dw 0            // Current piece board position 1-10 (0 and 11 are borders) * 256 (8.8 fixed point)
PieceY:         dw 0            // Current piece board position 0-23 (25 is border) * 256 (8.8 fixed point)
PieceDX:        dw 0            // Current piece delta to actual position
Rotation:       dw 0            // Current piece rotation 0-3

NextPiece1:     dw 0
NextPiece2:     dw 0
NextPiece3:     dw 0

SpaceWait:      dw 0

// Game board ... 12x24 color # of each cell, but left/bottom/right are solid and not drawn
GameBoard:      ds 12*25*2

GameOverX       = SCREEN_WIDTH / 2 - 4 * LCD_CHAR_SPACE
GameOverY       = SCREEN_HEIGHT / 2
GameOverMessage:    db "GAME OVER",0

DebugHere:  dw 0x0101
            dw buffer
.buffer     db "here\n",0


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
            cpy REG_IO_REQ, #ColorBG
            cpy REG_IO_REQ, #clear

            cpy REG_IO_REQ, #ColorBlack
            cpy REG_IO_REQ, #boardRect
            cpy REG_IO_REQ, #ColorWhite
            cpy REG_IO_REQ, #boardRect2

            jsr DrawBoard
            jsr DrawScore

            cmp Piece, #255
            jne noNewPiece
            jsr CreatePiece
            cmp Piece, #255
            jeq showGameOver
.noNewPiece
            jsr DrawPieces
            jsr UpdateGame
            jmp done
.showGameOver
            // In game-over mode
            cpy REG_IO_REQ, #ColorWhite
            cpy tx, #GameOverX
            cpy ty, #GameOverY
            psh #GameOverMessage
            jsr DrawString
            pop #2
            cmp KeySpaceDown, #1
            jne done
            jsr NewGame
.done            
            cpy REG_IO_REQ, #present
            ret

            // device request to clear screen
.clear      dw 0x0204

            // device request to present backbuffer to screen
.present    dw 0x0203
            dw 10               // delay ms

.boardRect  dw 0x0208
            dw BOARD_X + CELL_SIZE / 2
            dw BOARD_Y + CELL_SIZE / 2
            dw BOARD_WIDTH - CELL_SIZE - 1
            dw BOARD_HEIGHT - CELL_SIZE - 1
.boardRect2  dw 0x0207
            dw BOARD_X + CELL_SIZE / 2
            dw BOARD_Y + CELL_SIZE / 2
            dw BOARD_WIDTH - CELL_SIZE - 1
            dw BOARD_HEIGHT - CELL_SIZE - 1

DrawScore():
            // Level   000000
            cpy REG_IO_REQ, #ColorScoreBG
            cpy REG_IO_REQ, #level_rect
            cpy REG_IO_REQ, #ColorWhite
            cpy tx, #text_x
            cpy ty, #level_y
            psh #level_string
            jsr DrawString
            pop #2
            //add tx, #2*LCD_CHAR_SPACE
            psh Level
            jsr Draw5Digit
            pop #2

            // Lines   000000
            cpy REG_IO_REQ, #ColorScoreBG
            cpy REG_IO_REQ, #lines_rect
            cpy REG_IO_REQ, #ColorWhite
            cpy tx, #text_x
            cpy ty, #lines_y
            psh #lines_string
            jsr DrawString
            pop #2
            //add tx, #2*LCD_CHAR_SPACE
            psh Lines
            jsr Draw5Digit
            pop #2

            // Score   000000
            cpy REG_IO_REQ, #ColorScoreBG
            cpy REG_IO_REQ, #score_rect
            cpy REG_IO_REQ, #ColorWhite
            cpy tx, #text_x
            cpy ty, #score_y
            psh #score_string
            jsr DrawString
            pop #2
            //add tx, #2*LCD_CHAR_SPACE
            psh Score
            jsr Draw5Digit
            pop #2

            ret

.text_x      = 30
.level_y     = 2 * (SCREEN_HEIGHT / 3)
.lines_y     = level_y + LCD_LINE_SPACE + 5
.score_y     = lines_y + LCD_LINE_SPACE + 5

.level_string   db  "LEVEL ",0
.lines_string   db  "LINES ",0
.score_string   db  "SCORE ",0

.level_rect dw 0x0208
            dw text_x - 10
            dw level_y - 3
            dw 12 * LCD_CHAR_SPACE
            dw LCD_LINE_SPACE
.lines_rect dw 0x0208
            dw text_x - 10
            dw lines_y - 3
            dw 12 * LCD_CHAR_SPACE
            dw LCD_LINE_SPACE
.score_rect dw 0x0208
            dw text_x - 10
            dw score_y - 3
            dw 12 * LCD_CHAR_SPACE
            dw LCD_LINE_SPACE

Draw5Digit(value word):
            .t1 local word
            .t2 local word
            .divisor local word
            .first local word

            cpy first, #0
            cpy divisor, #10000
.loop
            cpy t1, value
            div t1, divisor
            cpy t2, t1
            mul t2, divisor
            sub value, t2

            cmp t1, #0
            jne fgColor
            cmp divisor, #1
            jeq fgColor
            cmp first, #0
            jne fgColor
            cpy REG_IO_REQ, #ColorGrey
            jmp printChar
.fgColor
            cpy first, #1
            cpy REG_IO_REQ, #ColorScoreFG
.printChar
            add t1, #0x10
            psh t1
            jsr DrawCharacter
            pop #2

            cmp divisor, #1
            jeq done
            div divisor, #10
            jne loop
.done
            ret

DrawBoard():
            .i local word
            .j local word
            .x local word
            .y local word
            .boardPtr local word
            .color local word
            .t1 local word

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
            mul boardPtr, #24           // 12 entries * 2 bytes per entry
            cpy t1, i
            mul t1, #2
            add boardPtr, t1
            add boardPtr, #GameBoard
            cpy color, *boardPtr

            mul color, #2
            add color, #BlockColors
            cpy REG_IO_REQ, *color
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
            dw 1,0,0,0,0,0,0,0,0,0,0,1
            dw 1,0,0,0,0,0,0,0,0,0,0,1
            dw 1,0,0,0,0,0,0,0,0,0,0,1
            dw 1,0,0,0,0,0,0,0,0,0,0,1
            dw 1,0,0,0,0,0,0,0,0,0,0,1
            dw 1,0,0,0,0,0,0,0,0,0,0,1
            dw 1,0,0,0,0,0,0,0,0,0,0,1
            dw 1,0,0,0,0,0,0,0,0,0,0,1
            dw 1,0,0,0,0,0,0,0,0,0,0,1
            dw 1,0,0,0,0,0,0,0,0,0,0,1
            dw 1,0,0,0,0,0,0,0,0,0,0,1
            dw 1,0,0,0,0,0,0,0,0,0,0,1
            dw 1,0,0,0,0,0,0,0,0,0,0,1
            dw 1,0,0,0,0,0,0,0,0,0,0,1
            dw 1,0,0,0,0,0,0,0,0,0,0,1
            dw 1,0,0,0,0,0,0,0,0,0,0,1
            dw 1,0,0,0,0,0,0,0,0,0,0,1
            dw 1,0,0,0,0,0,0,0,0,0,0,1
            dw 1,0,0,0,0,0,0,0,0,0,0,1
            dw 1,0,0,0,0,0,0,0,0,0,0,1
            dw 1,0,0,0,0,0,0,0,0,0,0,1
            dw 1,0,0,0,0,0,0,0,0,0,0,1
            dw 1,0,0,0,0,0,0,0,0,0,0,1
            dw 1,0,0,0,0,0,0,0,0,0,0,1
            dw 1,1,1,1,1,1,1,1,1,1,1,1
ResetEnd:
            dw 0

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

CreatePiece():
            .valid local word

            cpy Piece, NextPiece1
            cpy NextPiece1, NextPiece2
            cpy NextPiece2, NextPiece3

            // NextPiece3 = random(7)
            psh #0
            psh #7
            jsr Random
            pop #2
            pop NextPiece3

            cpy Rotation, #0
            cpy PieceX, #(4*COORD_SCALE)
            cpy PieceY, #(2*COORD_SCALE)
            cpy PieceDX, PieceX

            psh #0
            jsr IsValid
            pop valid
            jne done
            cpy Piece, #255
.done            
            ret

IsValid(valid word):
            .i local word
            .mask local word
            .bx local word
            .by local word
            .block local word
            .t1 local word
            .ptr local word

            cpy bx, PieceX
            div bx, #COORD_SCALE
            cpy by, PieceY
            div by, #COORD_SCALE

            // Lookup block mask
            cpy ptr, Piece
            mul ptr, #8
            cpy t1, Rotation
            mul t1, #2
            add ptr, t1
            add ptr, #BlockTable
            cpy block, *ptr

            cpy i, #1
            cpy mask, #1
.loop           
            cpy t1, block
            and t1, mask
            jeq row

            // Lookup what's on the game board in this position
            cpy ptr, by
            mul ptr, #24
            cpy t1, bx
            mul t1, #2
            add ptr, t1
            add ptr, #GameBoard
            cmp *ptr, #0
            jeq row
            cpy valid, #0
            ret
.row
            cmp i, #4
            jeq horiz
            cmp i, #8
            jeq horiz
            cmp i, #12
            jne vert
.horiz
            sub bx, #3
            inc by
            jmp endloop
.vert
            inc bx
.endloop
            mul mask, #2
            inc i
            cmp i, #15
            jlt loop
.done
            cpy valid, #1
            ret

UpdateGame():
            .t1 local word
            .oldx local word
            .oldy local word
            .oldr local word

            cpy oldx, PieceX
            cpy oldy, PieceY
            cpy oldr, Rotation

            cmp KeyJDown, #1
            jne checkRight
            cmp PieceX, #33
            jlt checkRight
            sub PieceX, #8
.checkRight
            cmp KeyLDown, #1
            jne bumpY
            add PieceX, #8
.bumpY
            cpy t1, Level
            mul t1, #3
            add t1, #2
            add PieceY, t1

            cmp KeyKDown, #1
            jne checkSpace
            add PieceY, #54
.checkSpace            
            cmp KeySpaceDown, #0
            jne rotate
            cpy SpaceWait, #0
            jmp checkValid
.rotate
            cmp SpaceWait, #0
            jne checkValid
            cpy SpaceWait, #1
            inc Rotation
            cmp Rotation, #4
            jlt checkValid
            cpy Rotation, #0
.checkValid 
            psh #0
            jsr IsValid
            pop t1
            jne done

            // current state isn't valid, reset to last state
            cmp PieceX, oldx
            jne invalid2
            cmp Rotation, oldr
            jne invalid2
            cmp PieceY, oldy
            jeq invalid2
            cpy PieceY, oldy
            jsr StampyTown
            jsr CollapseRows
            jmp done
.invalid2
            cmp PieceX, oldx
            jeq else
            cmp Rotation, oldr
            jeq else
            cpy PieceX, oldx
            jmp checkValid
.else
            cmp Rotation, oldr
            jeq movingOn
            // todo: play sound then fall through
.movingOn
            cpy PieceX, oldx
            cpy Rotation, oldr
            cmp PieceY, oldy
            jne checkValid
.done            
            ret

StampyTown():
            .i local word
            .mask local word
            .bx local word
            .by local word
            .block local word
            .t1 local word
            .ptr local word

            cpy bx, PieceX
            div bx, #COORD_SCALE
            cpy by, PieceY
            div by, #COORD_SCALE

            // Lookup block mask
            cpy ptr, Piece
            mul ptr, #8
            cpy t1, Rotation
            mul t1, #2
            add ptr, t1
            add ptr, #BlockTable
            cpy block, *ptr

            cpy i, #1
            cpy mask, #1
.loop           
            cpy t1, block
            and t1, mask
            jeq row

            // Mark game board
            cpy ptr, by
            mul ptr, #24
            cpy t1, bx
            mul t1, #2
            add ptr, t1
            add ptr, #GameBoard
            cpy *ptr, Piece
            inc *ptr
.row
            cmp i, #4
            jeq horiz
            cmp i, #8
            jeq horiz
            cmp i, #12
            jne vert
.horiz
            sub bx, #3
            inc by
            jmp endloop
.vert
            inc bx
.endloop
            mul mask, #2
            inc i
            cmp i, #15
            jlt loop
            cpy Piece, #255
            // todo bonus time
            ret

CollapseRows():
.by         local word
.total_line local word
.i          local word
.j          local word
.k          local word
.ptr        local word
.t1         local word

            cpy total_line, #0
            cpy by, PieceY
            div by, #COORD_SCALE
            cpy j, #0
.loop           
            cmp by, #24
            jeq updateScore

            // count how many in a row are occupied            
            cpy ptr, by

            mul ptr, #24        // 12 across times 2 bytes/cell
            add ptr, #GameBoard
            add ptr, #2         // skip left column
            cpy i, #10
.countLoop  
            cmp *ptr, #0
            jeq nextRow
            add ptr, #2          
            dec i
            jne countLoop

            inc total_line      // all 10 were lit up, collapse this row
            cpy ptr, by
            mul ptr, #24        // 12 across times 2 bytes/cell
            add ptr, #GameBoard
            add ptr, #2         // skip left column
            cpy t1, by
            dec t1
            mul t1, #24        
            add t1, #GameBoard
            add t1, #2       
            cpy k, by           // loop from by to 2 step -1  
.collapseLoop
            cpy i, #10
.copyRow
            cpy *ptr, *t1
            add ptr, #2
            add t1, #2
            dec i
            jne copyRow
            sub ptr, #20+24
            sub t1, #20+24
            dec k
            cmp k, #2
            jge collapseLoop
.nextRow
            inc by
            inc j
            cmp j, #4
            jge updateScore
            jmp loop
.updateScore
            cpy ptr, total_line
            mul ptr, #2
            add ptr, #LineScoreTable
            add Score, *ptr
            add Lines, total_line
            cpy t1, Level
            mul t1, #15
            cmp Lines, t1
            jlt done
            inc Level
.done            
            ret

LineScoreTable:
            dw 0, 100, 200, 500, 2000

// BlockTable defines the block shapes using bitmasks.
// Each row defines one block, with masks for rotation=0,1,2,3
// Index: block_num * 8 + rotation * 2
BlockTable:
            dw 2+32+512+1024, 16+32+64+256, 1+2+32+512, 4+16+32+64
            dw 2+32+256+512, 1+16+32+64, 2+4+32+512, 16+32+64+1024
            dw 16+32+64+512, 2+32+64+512, 2+16+32+64, 2+16+32+512
            dw 32+64+512+1024, 32+64+512+1024, 32+64+512+1024, 32+64+512+1024
            dw 2+32+64+1024, 2+4+16+32, 2+32+64+1024, 2+4+16+32
            dw 16+32+512+1024, 4+64+32+512, 16+32+512+1024, 4+64+32+512
            dw 2+32+512+8192, 16+32+64+128, 2+32+512+8192, 16+32+64+128

DrawPiece(x word, y word, piece word, rotation word):
            .i local word
            .mask local word
            .block local word
            .t1 local word
            .ptr local word

            cpy ptr, piece
            inc ptr
            mul ptr, #2
            add ptr, #BlockColors
            cpy REG_IO_REQ, *ptr

            // Lookup block mask
            cpy ptr, piece
            mul ptr, #8
            cpy t1, rotation
            mul t1, #2
            add ptr, t1
            add ptr, #BlockTable
            cpy block, *ptr

            cpy i, #1
            cpy mask, #1
.loop           
            cpy t1, block
            and t1, mask
            jeq row

            cpy fillX, x
            cpy fillY, y
            cpy REG_IO_REQ, #fillRect
.row
            cmp i, #4
            jeq horiz
            cmp i, #8
            jeq horiz
            cmp i, #12
            jne vert
.horiz
            sub x, #CELL_SIZE*3
            add y, #CELL_SIZE
            jmp endloop
.vert
            add x, #CELL_SIZE
.endloop
            mul mask, #2
            inc i
            cmp i, #15
            jlt loop
.done
            ret

.fillRect   dw 0x0208
.fillX      dw 0
.fillY      dw 0
            dw CELL_SIZE-2
            dw CELL_SIZE-2

// Draw the current and next pieces on the screen.
DrawPieces():
            .t1 local word

            // next 1
            psh #NEXT_X
            psh #NEXT_Y1
            psh NextPiece1
            psh #0
            jsr DrawPiece
            pop #8

            // next 2
            psh #NEXT_X
            psh #NEXT_Y2
            psh NextPiece2
            psh #0
            jsr DrawPiece
            pop #8

            // next 3
            psh #NEXT_X
            psh #NEXT_Y3
            psh NextPiece3
            psh #0
            jsr DrawPiece
            pop #8

            // if dx > (px / COORD_SCALE * COORD_SCALE) then dx = dx - 32
            cpy t1, PieceX
            div t1, #COORD_SCALE
            mul t1, #COORD_SCALE
            cmp t1, PieceDX
            jge incdx
            sub PieceDX, #8
.incdx            
            // if dx < (px / COORD_SCALE * COORD_SCALE) then dx = dx + 32
            cmp PieceDX, t1
            jge draw1
            add PieceDX, #8
.draw1
            // x = dx * CELL_SIZE / COORD_SCALE + BOARD_LEFT
            // y = py * CELL_SIZE / COORD_SCALE + BOARD_TOP - CELL_SIZE - 1
            cpy t1, PieceDX
            mul t1, #CELL_SIZE
            div t1, #COORD_SCALE
            add t1, #BOARD_X
            psh t1
            cpy t1, PieceY
            mul t1, #CELL_SIZE
            div t1, #COORD_SCALE
            add t1, #BOARD_Y
            sub t1, #CELL_SIZE - 1
            psh t1
            psh Piece
            psh Rotation
            jsr DrawPiece
            pop #8

            ret
