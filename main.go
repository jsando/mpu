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
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/jsando/mpu/asm"
	"github.com/jsando/mpu/machine"
)

func main() {
	//var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	buildCmd := flag.NewFlagSet("build", flag.ExitOnError)
	outputFile := buildCmd.String("o", "", "Output file")

	runCmd := flag.NewFlagSet("run", flag.ExitOnError)
	sysmon := runCmd.Bool("m", false, "Open system monitor")

	fmtCmd := flag.NewFlagSet("fmt", flag.ExitOnError)
	rewrite := fmtCmd.Bool("w", false, "Rewrite original file")

	//if *cpuprofile != "" {
	//	f, err := os.Create(*cpuprofile)
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//	pprof.StartCPUProfile(f)
	//	defer pprof.StopCPUProfile()
	//}
	if len(os.Args) <= 1 {
		flag.Usage()
		os.Exit(1)
	}
	switch os.Args[1] {
	case "build":
		buildCmd.Parse(os.Args[2:])
		inputs := getInputs(buildCmd)
		var output = *outputFile
		if output == "" {
			file := inputs[0]
			output = file.Name()[:len(file.Name())-2] + ".bin"
		}
		build(inputs, output)
	case "run":
		runCmd.Parse(os.Args[2:])
		inputs := getInputs(runCmd)
		run(inputs, *sysmon)
	case "fmt":
		fmtCmd.Parse(os.Args[2:])
		inputs := getInputs(fmtCmd)
		format(inputs, *rewrite)
	default:
		fmt.Println("expected 'build' or 'run' command")
		os.Exit(1)
	}
}

func format(inputs []*os.File, rewrite bool) {
	if len(inputs) != 1 {
		fmt.Printf("only handling 1 file right now\n")
		os.Exit(1)
	}
	lexer := asm.NewLexer(inputs[0].Name(), inputs[0])
	parser := asm.NewParser(asm.NewInput([]asm.TokenReader{lexer}))
	parser.SetProcessImport(false)
	parser.Parse()
	parser.PrintErrors()
	if parser.HasErrors() {
		os.Exit(1)
	}
	p := asm.NewPrinter(os.Stdout)
	p.Print(parser.Statements())
}

func getInputs(flagSet *flag.FlagSet) []*os.File {
	var inputs []*os.File
	for _, name := range flagSet.Args() {
		f, err := os.Open(name)
		if err != nil {
			fmt.Printf("error opening file '%s': %s\n", name, err.Error())
			os.Exit(1)
		}
		inputs = append(inputs, f)
	}
	if len(inputs) == 0 {
		fmt.Printf("at least one input file is required\n")
		os.Exit(1)
	}
	return inputs
}

// Run can be invoked with 1 file that doesn't end with .s, or a list
// of files ending with .s
func run(inputs []*os.File, monitor bool) {
	bin := false
	src := false
	for _, f := range inputs {
		ext := filepath.Ext(f.Name())
		if ext == ".s" {
			src = true
		} else {
			bin = true
		}
	}
	if bin && src {
		fmt.Printf("can't mix both binary and source files\n")
		os.Exit(1)
	}
	if bin && len(inputs) > 1 {
		fmt.Printf("can't have more than one binary file\n")
		os.Exit(1)
	}

	var bytes []byte
	var err error
	setBaseDirFromInputFile(inputs[0].Name())
	if src {
		linker, _ := compile(inputs)
		bytes = linker.Code()
	} else {
		// run program
		bytes, err = ioutil.ReadAll(inputs[0])
		if err != nil {
			panic(err)
		}
	}
	m := machine.NewMachine(bytes)
	if monitor {
		monitor := &Monitor{machine: m, memory: m.Memory()}
		monitor.Run()
	} else {
		m.Run()
		//fmt.Printf("Program completed, memory dump:\n")
		//m.Dump(os.Stdout, 0, 65535)
	}
}

func setBaseDirFromInputFile(path string) {
	dir := filepath.Dir(path)
	os.Setenv(machine.BaseDirEnv, dir)
}

func compile(inputs []*os.File) (*asm.Linker, []string) {
	parser := asm.NewParser(newTokenReader(inputs))
	parser.Parse()
	parser.PrintErrors()
	if parser.HasErrors() {
		os.Exit(1)
	}
	linker := asm.NewLinker(parser.Statements())
	linker.Link()
	linker.PrintMessages()
	if linker.HasErrors() {
		os.Exit(1)
	}
	return linker, parser.Files()
}

func build(inputs []*os.File, outputName string) {
	linker, files := compile(inputs)
	asm.WriteListing(files, linker)
	code := linker.Code()
	err := ioutil.WriteFile(outputName, code, 0644)
	if err != nil {
		fmt.Printf("error writing object file '%s': %s\n", outputName, err)
	} else {
		fmt.Printf("wrote %d bytes to %s\n", len(code), outputName)
	}
}

func newTokenReader(inputs []*os.File) *asm.Input {
	var tr []asm.TokenReader
	for _, file := range inputs {
		tr = append(tr, asm.NewLexer(file.Name(), file))
	}
	return asm.NewInput(tr)
}
