// Simple graphics tests.
main():
    var i word

    // Initialize main window
    req #io_window_req

.loop:
    // poll for events until no more
    req #io_poll_req
    cmp io_poll_event, #0
    jne loop

    req #io_sdl_setcolor
    req #io_sdl_clear

    // Draw random colored rectangles
    cpy i, #20
.draw_rects:
    jsr RandomFilledRect

    dec i
    jne draw_rects

    req #io_present_req
    jmp loop

.io_window_req:
    dw 0x0201, 0
    dw 640  // width
    dw 480  // height
    dw window_title
.window_title:
    db "Hello World, from MPU!", 0

.io_poll_req:
    dw 0x0202, 0
.io_poll_event:
    dw 2
.io_poll_time:
    dw 2
    ds 8 // space for event data

.io_present_req:
    dw 0x0203, 0
    dw 16 // delay ms

.io_sdl_clear:
    dw 0x0204, 0

.io_sdl_setcolor:
    dw 0x0205, 0
    db 0,0,0,255

// result = value - (value / range * range)
Random(result word, range word):
    var i word
    var j word
    req #GetRandom
    cpy i, RandomResult
    cpy j, i
    div j, range
    mul j, range
    cpy result, i
    sec
    sub result, j
    ret

.GetRandom:
    dw 0x0102, 0
.RandomResult:
    dw 0

RandomFilledRect():
    // .rect_x = random(640)
    psh #0
    psh #640
    jsr Random
    pop #2
    pop rect_x

    // .rect_y = random(480)
    psh #0
    psh #480
    jsr Random
    pop #2
    pop rect_y

    // .rect_w = .rect_h = random(100)
    psh #0
    psh #100
    jsr Random
    pop #2
    pop rect_w
    cpy rect_h, rect_w

    // .color_{r,g,b} = random(255)
    psh #0
    psh #255
    jsr Random
    pop #2
    seb
    pop color_r
    pop #1
    clb

    psh #0
    psh #255
    jsr Random
    pop #2
    seb
    pop color_g
    pop #1
    clb

    psh #0
    psh #255
    jsr Random
    pop #2
    seb
    pop color_b
    pop #1
    clb

    req #color
    req #rect
    ret

.color:      dw 0x0205, 0
.color_r:    db 0
.color_g:    db 0
.color_b:    db 0
.color_a:    db 255

.rect:       dw 0x0208, 0
.rect_x:     dw 0
.rect_y:     dw 0
.rect_w:     dw 0
.rect_h:     dw 0

