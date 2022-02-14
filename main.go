package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/jsando/lilac/asm"
	machine2 "github.com/jsando/lilac/machine"
	"io/ioutil"
	"os"
	"path/filepath"
)

var inputPath = flag.String("i", "", "Input file")

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
	} else if ext == ".bin" {
		// run program
		bytes, err := ioutil.ReadAll(file)
		if err != nil {
			panic(err)
		}
		m := machine2.NewMachineFromSlice(bytes)
		m.Run()
		fmt.Printf("Program completed, memory dump:\n")
		fmt.Println(hex.Dump(m.Snapshot()))
	}
}
