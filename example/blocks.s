// Falling blocks game
//
// Copyright 2022 Jason Sando <jason.sando.lv@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

            import "random"
            import "lcd"
            import "stdio"
            import "strconv"

            org 0
            dw Main

            org 0x10
IO_REQUEST  = 6

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

// Common color commands ... copy these addresses to IO_REQUEST to set color
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

// Table of block colors, indexed by piece # + 1.  These are pointers
// to the IO commands to copy to IO_REQUEST to set that as the current
// color.
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
PieceX:         dw 0            // Current piece board position 1-10 (0 and 11 are borders) * COORD_SCALE
PieceY:         dw 0            // Current piece board position 0-23 (25 is border) * COORD_SCALE
PieceDX:        dw 0            // Current piece delta to actual position (to animate smooth as it moves)
Rotation:       dw 0            // Current piece rotation 0-3

NextPiece1:     dw 0
NextPiece2:     dw 0
NextPiece3:     dw 0

SpaceWait:      dw 0

OutroFlag:      dw 0            // if != 0, play the game over sequence

// Game board ... 12x24 color # of each cell, but left/bottom/right are solid and not drawn
GameBoard:      ds 12*25*2

GameOverX       = SCREEN_WIDTH / 2 - 4 * LCD_CHAR_SPACE
GameOverY       = SCREEN_HEIGHT / 2
GameOverMessage:    db "GAME OVER",0

WavDropBlock:   db "wav/shoot-01.wav",0
WavLineScore:   db "wav/collect-point-00.wav",0
WavLevelUp:     db "wav/achievement-01.wav",0
WavGameOver:    db "wav/lose-00.wav",0
WavNewGame:     db "wav/win-00.wav",0

DebugHere:  dw 0x0101           // Quick way to trace execution by printing stuff :)
            dw buffer
.buffer     db "here\n",0

LineScoreTable:
            dw 0, 100, 200, 500, 2000

// BlockTable defines the block shapes using bitmasks, bit 0 is the top left 
// of a 4x4 grid, bit 1 is the second on the top, and so on.  It only uses
// 14 bits because the bottom right 2 are not used by any shape.
// Each row defines one block, with entries for rotation=0,1,2,3
// Index: block_num * 8 + rotation * 2
//
// Bit positions:
//      0  1  2  3
//      4  5  6  7
//      8  9 10 11
//     12 13 14
//
// Decimal:
//      1       2       4       8
//      16      32      64      128
//      256     512     1024    2048 
//      4096    8192
//
BlockTable:
            dw 2+32+512+1024, 16+32+64+256, 1+2+32+512, 4+16+32+64
            dw 2+32+256+512, 1+16+32+64, 2+4+32+512, 16+32+64+1024
            dw 16+32+64+512, 2+32+64+512, 2+16+32+64, 2+16+32+512
            dw 32+64+512+1024, 32+64+512+1024, 32+64+512+1024, 32+64+512+1024
            dw 2+32+64+1024, 2+4+16+32, 2+32+64+1024, 2+4+16+32
            dw 16+32+512+1024, 4+64+32+512, 16+32+512+1024, 4+64+32+512
            dw 2+32+512+8192, 16+32+64+128, 2+32+512+8192, 16+32+64+128

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
            cpy IO_REQUEST, #init
            cpy IO_REQUEST, #initAudio
            cpy wavptr, #WavDropBlock
            cpy IO_REQUEST, #loadWav
            cpy wavptr, #WavLineScore
            cpy IO_REQUEST, #loadWav
            cpy wavptr, #WavLevelUp
            cpy IO_REQUEST, #loadWav
            cpy wavptr, #WavGameOver
            cpy IO_REQUEST, #loadWav
            cpy wavptr, #WavNewGame
            cpy IO_REQUEST, #loadWav
            ret

.init       dw 0x0201
            dw SCREEN_WIDTH
            dw SCREEN_HEIGHT
            dw title
.title      db "MPU Blocks", 0
.initAudio  dw 0x020a
.loadWav    dw 0x020b
.wavptr     dw 0

PollEvents():
.loop
            cpy IO_REQUEST, #poll
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
            var tabptr word

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
            var tabptr word

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
            cpy IO_REQUEST, #ColorBG
            cpy IO_REQUEST, #clear

            cpy IO_REQUEST, #ColorBlack
            cpy IO_REQUEST, #boardRect
            cpy IO_REQUEST, #ColorWhite
            cpy IO_REQUEST, #boardRect2

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
            cmp OutroFlag, #0
            jeq drawGameOver
            cpy OutroFlag, #0
            cpy IO_REQUEST, #gameOverSound
.drawGameOver
            cpy IO_REQUEST, #ColorWhite
            cpy tx, #GameOverX
            cpy ty, #GameOverY
            psh #GameOverMessage
            jsr DrawString
            pop #2
            cmp KeySpaceDown, #1
            jne done
            jsr NewGame
.done            
            cpy IO_REQUEST, #present
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
.gameOverSound  dw 0x020c
            dw WavGameOver

DrawScore():
            // Level   000000
            cpy IO_REQUEST, #ColorScoreBG
            cpy IO_REQUEST, #level_rect
            cpy IO_REQUEST, #ColorWhite
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
            cpy IO_REQUEST, #ColorScoreBG
            cpy IO_REQUEST, #lines_rect
            cpy IO_REQUEST, #ColorWhite
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
            cpy IO_REQUEST, #ColorScoreBG
            cpy IO_REQUEST, #score_rect
            cpy IO_REQUEST, #ColorWhite
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
            var t1  word
            var t2 word
            var divisor word
            var first word

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
            cpy IO_REQUEST, #ColorGrey
            jmp printChar
.fgColor
            cpy first, #1
            cpy IO_REQUEST, #ColorScoreFG
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
            var i word
            var j word
            var x word
            var y word
            var boardPtr word
            var color word
            var t1 word

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
            cpy IO_REQUEST, *color
            cpy IO_REQUEST, #fillRect

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

NewGame():
            var from word
            var to word

            cpy OutroFlag, #1
            cpy IO_REQUEST, #newGameSound

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

.newGameSound  dw 0x020c
            dw WavNewGame

CreatePiece():
            var valid word

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
            var i word
            var mask word
            var bx word
            var by word
            var block word
            var t1 word
            var ptr word

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
            var t1 word
            var oldx word
            var oldy word
            var oldr word

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
            var i word
            var mask word
            var bx word
            var by word
            var block word
            var t1 word
            var ptr word

            cpy IO_REQUEST, #dropSound

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

.dropSound  dw 0x020c
            dw WavDropBlock

CollapseRows():
var by          word
var total_line  word
var i           word
var j           word
var k           word
var ptr         word
var t1          word

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
            cmp total_line, #1
            jlt noscore
            cpy IO_REQUEST, #lineSound
.noscore
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
            cpy IO_REQUEST, #levelUpSound
.done            
            ret

.lineSound dw 0x020c
            dw WavLineScore
.levelUpSound dw 0x020c
            dw WavLevelUp

DrawPiece(x word, y word, piece word, rotation word):
            var i word
            var mask word
            var block word
            var t1 word
            var ptr word

            cpy ptr, piece
            inc ptr
            mul ptr, #2
            add ptr, #BlockColors
            cpy IO_REQUEST, *ptr

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
            cpy IO_REQUEST, #fillRect
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
            var t1 word

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
