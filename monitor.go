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

package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/jsando/mpu/machine"
)

type Monitor struct {
	machine *machine.Machine
	memory  machine.Memory
	next    int // implied address if no address is provided
}

const welcome = `
 ------------------------------------------
|   M P U   S y s t e m   M o n i t o r    |
 ------------------------------------------
`

/*
d/dump [start [end]]
l/list [start]
run address
s/step [address]
set [address value [value]*]
*/
func (m *Monitor) Run() {
	fmt.Printf(welcome)
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Printf("> ")
		if !scanner.Scan() {
			break
		}
		line := strings.TrimSpace(scanner.Text())
		if len(line) == 0 {
			continue
		}
		cmd := strings.Split(line, " ")

		switch cmd[0] {
		case "dump", "d":
			m.dump(cmd)
		case "list", "l":
			m.list(cmd)
		case "run", "r":
			m.run(cmd)
		case "step", "s":
			m.step(cmd)
		}
	}
}

func (m *Monitor) dump(cmd []string) {
	start := m.next
	end := start + 160 - 1
	if len(cmd) > 1 {
		i, err := parseInt(cmd[1])
		if err != nil {
			fmt.Printf("invalid addr (%cmd)\n", err)
			return
		}
		start = i
		end = start + 160 - 1
	}
	if len(cmd) > 2 {
		i, err := parseInt(cmd[2])
		if err != nil {
			fmt.Printf("invalid addr (%cmd)\n", err)
			return
		}
		end = i
	}
	if end < start {
		end = start + 160 - 1
	}
	m.Dump(os.Stdout, start, end)
	m.next = end + 1
}

func (m *Monitor) list(cmd []string) {
	start := m.next
	if len(cmd) > 1 {
		i, err := parseInt(cmd[1])
		if err != nil {
			fmt.Printf("invalid addr (%s)\n", err)
		}
		start = i
	}
	m.next = m.List(os.Stdout, start, 20)
}

func (m *Monitor) run(cmd []string) {
	addr := m.next
	if len(cmd) > 1 {
		i, err := parseInt(cmd[1])
		if err != nil {
			fmt.Printf("invalid addr (%s)\n", err)
			return
		}
		addr = i
	}
	m.RunAt(addr)
}

func (m *Monitor) step(cmd []string) {
	addr := m.next
	if len(cmd) > 1 {
		i, err := parseInt(cmd[1])
		if err != nil {
			fmt.Printf("invalid addr (%s)\n", err)
			return
		}
		addr = i
	}
	m.next = m.Step(addr)
}

func parseInt(s string) (int, error) {
	if strings.HasPrefix(s, "0x") {
		i, err := strconv.ParseInt(s[2:], 16, 16)
		return int(i), err
	}
	i, err := strconv.ParseInt(s, 10, 16)
	return int(i), err
}

func (m *Monitor) Dump(w io.Writer, start int, end int) {
	ascii := make([]byte, 16)
	charIndex := 0
	flush := func() {
		for charIndex < 16 {
			ascii[charIndex] = ' '
			charIndex++
			fmt.Fprintf(w, "   ")
		}
		fmt.Fprintf(w, " |%s|\n", string(ascii))
		charIndex = 0
	}
	for addr := start; addr <= end; addr++ {
		if addr == start || charIndex == 16 {
			if addr != start {
				flush()
			}
			fmt.Fprintf(w, "%04x  ", addr)
		}
		ch := m.memory.GetByte(uint16(addr))
		fmt.Fprintf(w, "%02x ", ch)
		if ch >= 32 && ch <= 126 {
			ascii[charIndex] = ch
		} else {
			ascii[charIndex] = '.'
		}
		charIndex++
	}
	flush()
}

// List will disassemble n instructions starting at addr, and return the
// pc location following the last instruction.
func (m *Monitor) List(w io.Writer, addr int, n int) int {
	for i := 0; i < n; i++ {
		in := m.memory.GetByte(uint16(addr))
		op, m1, m2 := machine.DecodeOp(in)
		bytes := 0
		var args string
		if m1 != machine.Implied && m2 != machine.Implied {
			op1, n := m.formatOperand(m1, addr+1)
			bytes += n
			op2, n := m.formatOperand(m2, addr+1+n)
			bytes += n
			args = op1 + "," + op2
		} else if m1 != machine.Implied && m2 == machine.Implied {
			op1, n := m.formatOperand(m1, addr+1)
			bytes += n
			args = op1
		}
		fmt.Fprintf(w, "0x%04x  %02x ", addr, in)
		for j := 0; j < 4; j++ {
			if j < bytes {
				fmt.Fprintf(w, "%02x ", m.memory.GetByte(uint16(addr+j+1)))
			} else {
				fmt.Fprintf(w, "   ")
			}
		}
		fmt.Fprintf(w, "%s %s\n", op, args)
		addr = addr + 1 + bytes
	}
	return addr
}

func (m *Monitor) RunAt(addr int) {
	m.machine.RunAt(uint16(addr))
}

func (m *Monitor) Step(addr int) int {
	m.List(os.Stdout, addr, 1)
	next := m.machine.Step(uint16(addr))
	flags := m.machine.Flags()
	fmt.Printf("[status pc=%04x sp=%04x fp=%04x n=%d z=%d c=%d b=%d]\n",
		flags.PC, flags.SP, flags.FP, boolInt(flags.Negative), boolInt(flags.Zero),
		boolInt(flags.Carry), boolInt(flags.Bytes))
	return int(next)
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func (m *Monitor) formatOperand(mode machine.AddressMode, pc int) (op string, bytes int) {
	switch mode {
	case machine.Implied:
		// nothing to do
	case machine.Immediate:
		op = fmt.Sprintf("#0x%04x", m.memory.GetWord(uint16(pc)))
		bytes = 2
	case machine.ImmediateByte:
		op = fmt.Sprintf("#0x%02x", m.machine.ReadInt8(uint16(pc)))
		bytes = 1
	case machine.OffsetByte:
		value := m.machine.ReadInt8(uint16(pc))
		op = fmt.Sprintf("0x%04x (%d)", (pc-1)+value, value)
		bytes = 1
	case machine.Absolute:
		op = fmt.Sprintf("0x%04x", m.memory.GetWord(uint16(pc)))
		bytes = 2
	case machine.Indirect:
		op = fmt.Sprintf("*0x%04x", m.memory.GetWord(uint16(pc)))
		bytes = 2
	case machine.Relative:
		value := m.machine.ReadInt8(uint16(pc))
		op = fmt.Sprintf("fp%+d", value)
		bytes = 1
	case machine.RelativeIndirect:
		value := m.machine.ReadInt8(uint16(pc))
		op = fmt.Sprintf("*fp%+d", value)
		bytes = 1
	default:
		panic(fmt.Sprintf("illegal address mode: %d", mode))
	}
	return
}
