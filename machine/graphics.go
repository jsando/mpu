package machine

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/veandco/go-sdl2/gfx"
	"github.com/veandco/go-sdl2/sdl"
	"os"
)

// SDLDevice provides a graphics adapter via go-sdl2.
type SDLDevice struct {
	machine  *Machine
	handlers []SDLHandler
	window   *sdl.Window
	renderer *sdl.Renderer
}

type SDLHandler interface {
	Exec(d *SDLDevice, addr uint16) uint16
}

func NewSDLDevice(machine *Machine) *SDLDevice {
	d := &SDLDevice{
		machine: machine,
	}
	d.handlers = []SDLHandler{
		1: &CmdSDLInit{},
		&CmdSDLPollEvents{},
		&CmdSDLPresent{},
		&CmdSDLClear{},
		&CmdSDLSetColor{},
		&CmdSDLDrawLine{},
		&CmdSDLDrawRect{},
		&CmdSDLFillRect{},
	}
	return d
}

func (d *SDLDevice) Invoke(machine *Machine, addr uint16) uint16 {
	cmd := int(machine.readUint16(addr + 2))
	if cmd >= len(d.handlers) || d.handlers[cmd] == nil {
		return ErrInvalidCommand
	}
	handler := d.handlers[cmd]
	err := binary.Read(bytes.NewReader(d.machine.memory[addr:]), binary.LittleEndian, handler)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error decoding io request command=%d, error=%s\n", cmd, err.Error())
		return ErrIOError
	}
	return handler.Exec(d, addr)
}

type CmdSDLInit struct {
	Device  uint16
	Command uint16
	Width   uint16
	Height  uint16
	Title   uint16 // Pointer to zstring
}

func (c *CmdSDLInit) Exec(d *SDLDevice, addr uint16) uint16 {
	winTitle := d.machine.ReadString(c.Title)
	//fmt.Printf("execInit: %v, title: '%s'\n", c, winTitle)
	window, err := sdl.CreateWindow(winTitle, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		int32(c.Width), int32(c.Height), sdl.WINDOW_SHOWN)
	if err != nil {
		fmt.Printf("error creating SDL window: %s", err.Error())
		return ErrIOError
	}
	d.window = window
	d.renderer, err = sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		fmt.Printf("Failed to create renderer: %s\n", err)
		return ErrIOError
	}
	return ErrNoErr
}

type CmdSDLPollEvents struct {
	Device    uint16
	Command   uint16
	EventType uint16 // space for response
	Timestamp uint16 // space for response
}

func (c *CmdSDLPollEvents) Exec(d *SDLDevice, addr uint16) uint16 {
	if d.window == nil {
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
	d.machine.writeUint16(addr+4, eventType)
	d.machine.writeUint16(addr+6, timestamp)

	switch t := event.(type) {
	case *sdl.KeyboardEvent:
		keyCode := uint16(t.Keysym.Sym)
		d.machine.writeUint16(addr+8, keyCode)
	}
	return ErrNoErr
}

type CmdSDLSetColor struct {
	Device     uint16
	Command    uint16
	R, G, B, A uint8
}

func (c CmdSDLSetColor) Exec(d *SDLDevice, addr uint16) uint16 {
	if d.renderer == nil {
		fmt.Printf("error in setcolor, sdl not initialized\n")
		return ErrIOError
	}
	//fmt.Printf("setcolor %v\n", c)
	err := d.renderer.SetDrawColor(c.R, c.G, c.B, c.A)
	if err != nil {
		fmt.Printf("setcolor error: %s\n", err.Error())
		return ErrIOError
	}
	return ErrNoErr
}

type CmdSDLClear struct {
	Device  uint16
	Command uint16
}

func (c *CmdSDLClear) Exec(d *SDLDevice, addr uint16) uint16 {
	if d.renderer == nil {
		fmt.Printf("error in clear, sdl not initialized\n")
		return ErrIOError
	}
	d.renderer.Clear()
	return ErrNoErr
}

type CmdSDLDrawLine struct {
	Device         uint16
	Command        uint16
	X1, Y1, X2, Y2 uint16
}

func (c CmdSDLDrawLine) Exec(d *SDLDevice, addr uint16) uint16 {
	if d.renderer == nil {
		fmt.Printf("error in drawline, sdl not initialized\n")
		return ErrIOError
	}
	//fmt.Printf("drawline: %v\n", c)
	err := d.renderer.DrawLine(int32(c.X1), int32(c.Y1), int32(c.X2), int32(c.Y2))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error in drawline: %s\n", err.Error())
		return ErrIOError
	}
	return ErrNoErr
}

type CmdSDLDrawRect struct {
	Device     uint16
	Command    uint16
	X, Y, W, H uint16
}

func (c *CmdSDLDrawRect) Exec(d *SDLDevice, addr uint16) uint16 {
	if d.renderer == nil {
		fmt.Printf("error in drawline, sdl not initialized\n")
		return ErrIOError
	}
	var rect sdl.Rect
	rect.X = int32(c.X)
	rect.Y = int32(c.Y)
	rect.W = int32(c.W)
	rect.H = int32(c.H)
	err := d.renderer.DrawRect(&rect)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error in drawrect: %s\n", err.Error())
		return ErrIOError
	}
	return ErrNoErr
}

type CmdSDLFillRect struct {
	Device     uint16
	Command    uint16
	X, Y, W, H uint16
}

func (c CmdSDLFillRect) Exec(d *SDLDevice, addr uint16) uint16 {
	if d.renderer == nil {
		fmt.Printf("error in drawline, sdl not initialized\n")
		return ErrIOError
	}
	var rect sdl.Rect
	rect.X = int32(c.X)
	rect.Y = int32(c.Y)
	rect.W = int32(c.W)
	rect.H = int32(c.H)
	//fmt.Printf("fillrect %v\n", rect)
	err := d.renderer.FillRect(&rect)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error in drawrect: %s\n", err.Error())
		return ErrIOError
	}
	return ErrNoErr
}

type CmdSDLPresent struct {
	Device  uint16
	Command uint16
	DelayMS uint16
}

func (c *CmdSDLPresent) Exec(d *SDLDevice, addr uint16) uint16 {
	if d.renderer == nil {
		fmt.Fprintf(os.Stderr, "SDL error, not initialized\n")
		return ErrIOError
	}
	gfx.StringColor(d.renderer, 16, 16, "GFX Demo", sdl.Color{0, 255, 0, 255})

	d.renderer.Present()
	if c.DelayMS > 0 {
		sdl.Delay(uint32(c.DelayMS))
	}
	return ErrNoErr
}
