package machine

import (
	"fmt"
	"github.com/veandco/go-sdl2/sdl"
	"os"
)

const (
	SdlDeviceId = 0x0200
	SdlInit     = 1
	SdlPoll     = 2
	SdlPresent  = 3
	SdlClear    = 4
	SdlSetColor = 5
	SdlDrawLine = 6
	SdlDrawRect = 7
	SdlFillRect = 8
)

// I feel SO DIRTY having a global var but don't see how to encapsulate this better.
// The handlers can't have unexported fields because that breaks encoding/binary :(
var window *sdl.Window
var renderer *sdl.Renderer

func RegisterSDLHandlers(m *Machine) {
	m.RegisterIOHandler(SdlDeviceId|SdlInit, &SdlInitHandler{})
	m.RegisterIOHandler(SdlDeviceId|SdlPoll, &SdlPollHandler{})
	m.RegisterIOHandler(SdlDeviceId|SdlPresent, &SdlPresentHandler{})
	m.RegisterIOHandler(SdlDeviceId|SdlClear, &SdlClearHandler{})
	m.RegisterIOHandler(SdlDeviceId|SdlSetColor, &SdlSetColorHandler{})
	m.RegisterIOHandler(SdlDeviceId|SdlDrawLine, &SdlDrawLineHandler{})
	m.RegisterIOHandler(SdlDeviceId|SdlDrawRect, &SdlDrawRectHandler{})
	m.RegisterIOHandler(SdlDeviceId|SdlFillRect, &SdlFillRectHandler{})
}

type SdlInitHandler struct {
	Id     uint16
	Width  uint16
	Height uint16
	Title  uint16 // Pointer to zstring
}

func (c *SdlInitHandler) Handle(m *Machine, addr uint16) uint16 {
	winTitle := m.ReadString(c.Title)
	//fmt.Printf("execInit: %v, title: '%s'\n", c, winTitle)
	var err error
	window, err = sdl.CreateWindow(winTitle, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		int32(c.Width), int32(c.Height), sdl.WINDOW_SHOWN)
	if err != nil {
		fmt.Printf("error creating SDL window: %s", err.Error())
		return ErrIOError
	}
	renderer, err = sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		fmt.Printf("Failed to create renderer: %s\n", err)
		return ErrIOError
	}
	return ErrNoErr
}

type SdlPollHandler struct {
	Id        uint16
	EventType uint16 // space for response
	Timestamp uint16 // space for response
}

func (c *SdlPollHandler) Handle(m *Machine, addr uint16) uint16 {
	if window == nil {
		fmt.Printf("sdl not initialized, can't poll\n")
		return ErrIOError
	}
	event := sdl.PollEvent()
	eventType := uint16(0)
	timestamp := uint16(0)
	if event != nil {
		if event.GetType() > 65535 {
			fmt.Printf("error: event type %d exceeds uint16\n", event.GetType())
			return ErrIOError
		}
		eventType = uint16(event.GetType())
		timestamp = uint16(event.GetTimestamp() / 250)
		//fmt.Printf("sdl poll events (event=%d, time=%d)\n", eventType, timestamp)
	}
	m.writeUint16(addr+2, eventType)
	m.writeUint16(addr+4, timestamp)

	switch t := event.(type) {
	case *sdl.KeyboardEvent:
		keyCode := uint16(t.Keysym.Sym)
		m.writeUint16(addr+6, keyCode)
	}
	return ErrNoErr
}

type SdlSetColorHandler struct {
	Id         uint16
	R, G, B, A uint8
}

func (c SdlSetColorHandler) Handle(m *Machine, addr uint16) uint16 {
	if renderer == nil {
		fmt.Printf("error in setcolor, sdl not initialized\n")
		return ErrIOError
	}
	//fmt.Printf("setcolor %v\n", c)
	err := renderer.SetDrawColor(c.R, c.G, c.B, c.A)
	if err != nil {
		fmt.Printf("setcolor error: %s\n", err.Error())
		return ErrIOError
	}
	return ErrNoErr
}

type SdlClearHandler struct {
	Id uint16
}

func (c *SdlClearHandler) Handle(m *Machine, addr uint16) uint16 {
	if renderer == nil {
		fmt.Printf("error in clear, sdl not initialized\n")
		return ErrIOError
	}
	renderer.Clear()
	return ErrNoErr
}

type SdlDrawLineHandler struct {
	Id             uint16
	X1, Y1, X2, Y2 uint16
}

func (c SdlDrawLineHandler) Handle(m *Machine, addr uint16) uint16 {
	if renderer == nil {
		fmt.Printf("error in drawline, sdl not initialized\n")
		return ErrIOError
	}
	//fmt.Printf("drawline: %v\n", c)
	err := renderer.DrawLine(int32(c.X1), int32(c.Y1), int32(c.X2), int32(c.Y2))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error in drawline: %s\n", err.Error())
		return ErrIOError
	}
	return ErrNoErr
}

type SdlDrawRectHandler struct {
	Id         uint16
	X, Y, W, H uint16
}

func (c *SdlDrawRectHandler) Handle(m *Machine, addr uint16) uint16 {
	if renderer == nil {
		fmt.Printf("error in drawline, sdl not initialized\n")
		return ErrIOError
	}
	var rect sdl.Rect
	rect.X = int32(c.X)
	rect.Y = int32(c.Y)
	rect.W = int32(c.W)
	rect.H = int32(c.H)
	err := renderer.DrawRect(&rect)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error in drawrect: %s\n", err.Error())
		return ErrIOError
	}
	return ErrNoErr
}

type SdlFillRectHandler struct {
	Id         uint16
	X, Y, W, H uint16
}

func (c SdlFillRectHandler) Handle(m *Machine, addr uint16) uint16 {
	if renderer == nil {
		fmt.Printf("error in drawline, sdl not initialized\n")
		return ErrIOError
	}
	var rect sdl.Rect
	rect.X = int32(c.X)
	rect.Y = int32(c.Y)
	rect.W = int32(c.W)
	rect.H = int32(c.H)
	//fmt.Printf("fillrect %v\n", rect)
	err := renderer.FillRect(&rect)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error in drawrect: %s\n", err.Error())
		return ErrIOError
	}
	return ErrNoErr
}

type SdlPresentHandler struct {
	Id      uint16
	DelayMS uint16
}

func (c *SdlPresentHandler) Handle(m *Machine, addr uint16) uint16 {
	if renderer == nil {
		fmt.Fprintf(os.Stderr, "SDL error, not initialized\n")
		return ErrIOError
	}
	//gfx.StringColor(d.renderer, 16, 16, "GFX Demo", sdl.Color{0, 255, 0, 255})

	renderer.Present()
	if c.DelayMS > 0 {
		sdl.Delay(uint32(c.DelayMS))
	}
	return ErrNoErr
}
