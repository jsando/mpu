package main

import (
	"flag"
	"fmt"
	"github.com/jsando/mpu/asm"
	"github.com/jsando/mpu/machine"
	"io/ioutil"
	"os"
	"path/filepath"
)

var inputPath = flag.String("i", "", "Input file")
var sysmon = flag.Bool("m", false, "Open system monitor")
var runFlag = flag.Bool("r", false, "assemble and run")

func main() {
	flag.Parse()
	if len(*inputPath) == 0 {
		flag.PrintDefaults()
		os.Exit(1)
	}
	file, err := os.Open(*inputPath)
	if err != nil {
		panic(err)
	}
	ext := filepath.Ext(file.Name())
	if ext == ".s" {
		// assemble to bin file
		lexer := asm.NewLexer(file.Name(), file)
		parser := asm.NewParser(lexer)
		parser.Parse()
		parser.PrintErrors()
		if !parser.HasErrors() {
			linker := asm.NewLinker(parser.Fragments())
			linker.Link()
			linker.PrintMessages()
			if !linker.HasErrors() {
				if *runFlag {
					m := machine.NewMachine(linker.Code())
					m.Run()
				} else {
					file, err := os.Open(*inputPath)
					if err != nil {
						panic(err)
					}
					asm.WriteListing(file, os.Stdout, linker)
					outFileName := file.Name()[:len(file.Name())-2] + ".bin"
					code := linker.Code()
					err = ioutil.WriteFile(outFileName, code, 0644)
					if err != nil {
						fmt.Printf("error writing object file '%s': %s\n", outFileName, err)
					} else {
						fmt.Printf("wrote %d bytes to %s\n", len(code), outFileName)
					}
				}
			}
		}
	} else if ext == ".bin" {
		// run program
		bytes, err := ioutil.ReadAll(file)
		if err != nil {
			panic(err)
		}
		m := machine.NewMachine(bytes)
		if *sysmon {
			monitor := &Monitor{machine: m}
			monitor.Run()
		} else {
			m.Run()
			//fmt.Printf("Program completed, memory dump:\n")
			//m.Dump(os.Stdout, 0, 65535)
		}
	}
}
