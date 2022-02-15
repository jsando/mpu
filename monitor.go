package main

import (
	"bufio"
	"fmt"
	"github.com/jsando/lilac/machine2"
	"os"
	"strconv"
	"strings"
)

const welcome = `
 ------------------------------------------
|   M P U   S y s t e m   M o n i t o r    |
 ------------------------------------------
`

/*

d/dump [start [end]]
l/list [start]
set [address value [value]*]
run address
s/step [address]

*/
func RunMonitor(machine *machine2.Machine) {
	next := 0
	fmt.Printf(welcome)
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Printf("> ")
		if !scanner.Scan() {
			break
		}
		cmd := strings.TrimSpace(scanner.Text())
		if len(cmd) == 0 {
			continue
		}
		s := strings.Split(cmd, " ")
		switch s[0] {
		case "dump", "d":
			start := next
			end := start + 160 - 1
			if len(s) > 1 {
				i, err := parseInt(s[1])
				if err != nil {
					fmt.Printf("invalid addr (%s)\n", err)
					continue
				}
				start = i
				end = start + 160 - 1
			}
			if len(s) > 2 {
				i, err := parseInt(s[2])
				if err != nil {
					fmt.Printf("invalid addr (%s)\n", err)
					continue
				}
				end = i
			}
			if end < start {
				end = start + 160 - 1
			}
			machine.Dump(os.Stdout, start, end)
			next = end + 1
		case "list", "l":
			start := next
			if len(s) > 1 {
				i, err := parseInt(s[1])
				if err != nil {
					fmt.Printf("invalid addr (%s)\n", err)
					continue
				}
				start = i
			}
			next = machine.List(os.Stdout, start, 20)
		}
	}
}

func parseInt(s string) (int, error) {
	if strings.HasPrefix(s, "0x") {
		i, err := strconv.ParseInt(s[2:], 16, 16)
		return int(i), err
	}
	i, err := strconv.ParseInt(s, 10, 16)
	return int(i), err
}
