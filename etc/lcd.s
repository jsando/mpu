//-------------------------------------
// 14 Segment LCD Display
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
SCREEN_WIDTH    = 800
SCREEN_HEIGHT   = 600
SDL_QUIT        = 0x100
SDL_KEYDOWN     = 0x300
SDL_KEYUP       = 0x301
SDLK_ESCAPE     = 0x1b

// Globals
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
            cpy tx, left_margin
            cpy ty, #0
            psh #sample
            jsr DrawString
            pop #2

            add left_margin, margin_delta            
            cmp left_margin, #10
            jlt reverse
            cmp left_margin, #320
            jlt done
.reverse            
            mul margin_delta, #-1
.done            
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

margin_delta:   dw 2
left_margin:    dw 20
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

            cmp ch, #10         // linefeed
            jne validate
            cpy tx, left_margin
            add ty, #LINE_SPACE
            jmp next
.validate
            cmp ch, #32
            jlt next
            cmp ch, #128
            jge next

            sub ch, #32
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

.line       dw 0x0206
.line_x     dw 0
.line_y     dw 0
.line_x2    dw 0
.line_y2    dw 0

/*
     A 16-segment lcd font from https://github.com/dmadison/LED-Segment-ASCII

     Illustration here from https://github.com/MartyMacGyver/LCD_HT1622_16SegLcd

    /-----------\   /-----------\
   ||    'a'    || ||    'b'    ||
    \-----------/   \-----------/
   /-\ /--\      /-\      /--\ /-\
  |   |\   \    |   |    /   /|   |
  |   | \   \   |   |   /   / |   |
  |'h'|  \'k'\  |'m'|  /'n'/  |'c'|
  |   |   \   \ |   | /   /   |   |
  |   |    \   \|   |/   /    |   |
   \-/      \--/ \-/ \--/      \-/
    /-----------\   /-----------\
   ||    'u'    || ||    'p'    ||
    \-----------/   \-----------/
   /-\      /--\ /-\ /--\      /-\
  |   |    /   /|   |\   \    |   |
  |   |   /   / |   | \   \   |   |
  |'g'|  /'t'/  |'s'|  \'r'\  |'d'|
  |   | /   /   |   |   \   \ |   |
  |   |/   /    |   |    \   \|   |
   \-/ \--/      \-/      \--/ \-/
    /-----------\   /-----------\
   ||    'f'    || ||    'e'    ||  |DP|
    \-----------/   \-----------/

Order of bits within CharacterTable are as follows (A is bit 0):

    U-T-S-R-P-N-M-K-H-G-F-E-D-C-B-A

*/

CHAR_SPACE  = 36
LINE_SPACE  = 70

CharacterSegmentTable:
    .cw          = 32
    .x1          = 0
    .x2          = cw / 2
    .x3          = cw - 1
    .y1          = 0
    .y2          = cw
    .y3          = 2 * cw - 1
            dw x1+1,y1,x2-1,y1      // a
            dw x2+1,y1,x3-1,y1      // b
            dw x3,y1+1,x3,y2-1      // c
            dw x3,y2+1,x3,y3-1      // d
            dw x2+1,y3,x3-1,y3      // e
            dw x1+1,y3,x2-1,y3      // f
            dw x1,y2+1,x1,y3-1      // g
            dw x1,y1+1,x1,y2-1      // h
            dw x1,y1,x2,y2          // k
            dw x2,y1,x2,y2          // m
            dw x2,y2,x3,y1          // n
            dw x2,y2,x3,y2          // p
            dw x2,y2,x3,y3          // r
            dw x2,y2,x2,y3          // s
            dw x2,y2,x1,y3          // t
            dw x1+1,y2,x2-1,y2      // u

// ASCII printable characters 32-127, and which segments to light up for each.
// Bit 0 -> a (see comment above for map).
CharacterTable:
            dw 0x0000 /* (space) */
            dw 0x000C /* ! */
            dw 0x0204 /* " */
            dw 0xAA3C /* # */
            dw 0xAABB /* $ */
            dw 0xEE99 /* % */
            dw 0x9371 /* & */
            dw 0x0200 /* ' */
            dw 0x1400 /* ( */
            dw 0x4100 /* ) */
            dw 0xFF00 /* * */
            dw 0xAA00 /* + */
            dw 0x4000 /* , */
            dw 0x8800 /* - */
            dw 0x1000 /* . */
            dw 0x4400 /* / */
            dw 0x44FF /* 0 */
            dw 0x040C /* 1 */
            dw 0x8877 /* 2 */
            dw 0x083F /* 3 */
            dw 0x888C /* 4 */
            dw 0x90B3 /* 5 */
            dw 0x88FB /* 6 */
            dw 0x000F /* 7 */
            dw 0x88FF /* 8 */
            dw 0x88BF /* 9 */
            dw 0x2200 /* : */
            dw 0x4200 /* ; */
            dw 0x9400 /* < */
            dw 0x8830 /* = */
            dw 0x4900 /* > */
            dw 0x2807 /* ? */
            dw 0x0AF7 /* @ */
            dw 0x88CF /* A */
            dw 0x2A3F /* B */
            dw 0x00F3 /* C */
            dw 0x223F /* D */
            dw 0x80F3 /* E */
            dw 0x80C3 /* F */
            dw 0x08FB /* G */
            dw 0x88CC /* H */
            dw 0x2233 /* I */
            dw 0x007C /* J */
            dw 0x94C0 /* K */
            dw 0x00F0 /* L */
            dw 0x05CC /* M */
            dw 0x11CC /* N */
            dw 0x00FF /* O */
            dw 0x88C7 /* P */
            dw 0x10FF /* Q */
            dw 0x98C7 /* R */
            dw 0x88BB /* S */
            dw 0x2203 /* T */
            dw 0x00FC /* U */
            dw 0x44C0 /* V */
            dw 0x50CC /* W */
            dw 0x5500 /* X */
            dw 0x88BC /* Y */
            dw 0x4433 /* Z */
            dw 0x2212 /* [ */
            dw 0x1100 /* \ */
            dw 0x2221 /* ] */
            dw 0x5000 /* ^ */
            dw 0x0030 /* _ */
            dw 0x0100 /* ` */
            dw 0xA070 /* a */
            dw 0xA0E0 /* b */
            dw 0x8060 /* c */
            dw 0x281C /* d */
            dw 0xC060 /* e */
            dw 0xAA02 /* f */
            dw 0xA2A1 /* g */
            dw 0xA0C0 /* h */
            dw 0x2000 /* i */
            dw 0x2260 /* j */
            dw 0x3600 /* k */
            dw 0x00C0 /* l */
            dw 0xA848 /* m */
            dw 0xA040 /* n */
            dw 0xA060 /* o */
            dw 0x82C1 /* p */
            dw 0xA281 /* q */
            dw 0x8040 /* r */
            dw 0xA0A1 /* s */
            dw 0x80E0 /* t */
            dw 0x2060 /* u */
            dw 0x4040 /* v */
            dw 0x5048 /* w */
            dw 0x5500 /* x */
            dw 0x0A1C /* y */
            dw 0xC020 /* z */
            dw 0xA212 /* { */
            dw 0x2200 /* | */
            dw 0x2A21 /* } */
            dw 0xCC00 /* ~ */
            dw 0x0000 /* (del) */