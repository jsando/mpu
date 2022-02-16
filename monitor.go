package main

import (
	"bufio"
	"fmt"
	"github.com/jsando/lilac/machine"
	"os"
	"strconv"
	"strings"
)

type Monitor struct {
	machine *machine.Machine
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
	m.machine.Dump(os.Stdout, start, end)
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
	m.next = m.machine.List(os.Stdout, start, 20)
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
	m.machine.RunAt(addr)
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
	m.next = m.machine.Step(addr)
}

func parseInt(s string) (int, error) {
	if strings.HasPrefix(s, "0x") {
		i, err := strconv.ParseInt(s[2:], 16, 16)
		return int(i), err
	}
	i, err := strconv.ParseInt(s, 10, 16)
	return int(i), err
}
