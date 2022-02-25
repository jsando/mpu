// Simple graphics tests.
    dw main

IO_REQUEST = 6
IO_STATUS  = 8
RANDOM = 10

    org 0x100
main():
    .i local word

    // Initialize main window
    cpy IO_REQUEST, #io_window_req

.loop
    // poll for events until no more
    cpy IO_REQUEST, #io_poll_req
    cmp io_poll_event, #0
    jne loop

    cpy IO_REQUEST, #io_sdl_setcolor
    cpy IO_REQUEST, #io_sdl_clear

    // Draw random colored rectangles
    cpy i, #20
.draw_rects
    jsr RandomFilledRect
    dec i
    jne draw_rects

    cpy IO_REQUEST, #io_present_req
    jmp loop

.io_window_req
    dw 0x0201
    dw 640  // width
    dw 480  // height
    dw window_title
.window_title
    db "Hello World, from MPU!", 0

.io_poll_req
    dw 0x0202
.io_poll_event
    dw 2
.io_poll_time
    dw 2
    ds 8 // space for event data

.io_present_req
    dw 0x0203
    dw 16 // delay ms

.io_sdl_clear
    dw 0x0204

.io_sdl_setcolor
    dw 0x0205
    db 0,0,0,255

// result = value - (value / range * range)
Random(result word, range word):
    .i local word
    .j local word
    cpy i, RANDOM   // get a random number in range 0-65535
    cpy j, i        // value / range * range
    div j, range
    mul j, range
    cpy result, i
    sec
    sub result, j
    ret

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

    cpy IO_REQUEST, #color
    cpy IO_REQUEST, #rect
    ret

.color      dw 0x0205
.color_r    db 0
.color_g    db 0
.color_b    db 0
.color_a    db 255

.rect       dw 0x0208
.rect_x     dw 0
.rect_y     dw 0
.rect_w     dw 0
.rect_h     dw 0
