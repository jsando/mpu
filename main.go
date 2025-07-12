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
	"strings"

	"github.com/jsando/mpu/asm"
	"github.com/jsando/mpu/machine"
	"github.com/jsando/mpu/test"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// Handle version and help flags before subcommands
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--version", "-v":
			fmt.Printf("mpu version %s\n", version)
			fmt.Printf("  commit: %s\n", commit)
			fmt.Printf("  built:  %s\n", date)
			os.Exit(0)
		case "--help", "-h", "help":
			printUsage()
			os.Exit(0)
		}
	}

	buildCmd := flag.NewFlagSet("build", flag.ContinueOnError)
	outputFile := buildCmd.String("o", "", "output file (default: input.bin)")
	buildHelp := buildCmd.Bool("help", false, "show help for build command")

	runCmd := flag.NewFlagSet("run", flag.ContinueOnError)
	sysmon := runCmd.Bool("m", false, "open system monitor/debugger")
	runHelp := runCmd.Bool("help", false, "show help for run command")

	fmtCmd := flag.NewFlagSet("fmt", flag.ContinueOnError)
	rewrite := fmtCmd.Bool("w", false, "rewrite original file (not yet implemented)")
	fmtHelp := fmtCmd.Bool("help", false, "show help for fmt command")

	testCmd := flag.NewFlagSet("test", flag.ContinueOnError)
	testVerbose := testCmd.Bool("v", false, "verbose output")
	testColor := testCmd.Bool("color", true, "colorize output")
	testHelp := testCmd.Bool("help", false, "show help for test command")

	// Custom usage for subcommands
	buildCmd.Usage = func() { printBuildUsage() }
	runCmd.Usage = func() { printRunUsage() }
	fmtCmd.Usage = func() { printFmtUsage() }
	testCmd.Usage = func() { printTestUsage() }

	if len(os.Args) <= 1 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "build":
		if err := buildCmd.Parse(os.Args[2:]); err != nil {
			os.Exit(1)
		}
		if *buildHelp {
			printBuildUsage()
			os.Exit(0)
		}
		inputs := getInputs(buildCmd)
		var output = *outputFile
		if output == "" {
			file := inputs[0]
			output = strings.TrimSuffix(file.Name(), filepath.Ext(file.Name())) + ".bin"
		}
		build(inputs, output)
	case "run":
		if err := runCmd.Parse(os.Args[2:]); err != nil {
			os.Exit(1)
		}
		if *runHelp {
			printRunUsage()
			os.Exit(0)
		}
		inputs := getInputs(runCmd)
		run(inputs, *sysmon)
	case "fmt":
		if err := fmtCmd.Parse(os.Args[2:]); err != nil {
			os.Exit(1)
		}
		if *fmtHelp {
			printFmtUsage()
			os.Exit(0)
		}
		inputs := getInputs(fmtCmd)
		format(inputs, *rewrite)
	case "test":
		if err := testCmd.Parse(os.Args[2:]); err != nil {
			os.Exit(1)
		}
		if *testHelp {
			printTestUsage()
			os.Exit(0)
		}
		inputs := getInputs(testCmd)
		runTests(inputs, *testVerbose, *testColor)
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown command '%s'\n\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("MPU - Memory Processing Unit")
	fmt.Println("A 16-bit virtual computer system with assembler")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  mpu <command> [arguments]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  build    Assemble source files into binary")
	fmt.Println("  run      Execute a program")
	fmt.Println("  fmt      Format assembly source code")
	fmt.Println("  test     Run unit tests in assembly files")
	fmt.Println()
	fmt.Println("Global Options:")
	fmt.Println("  --help, -h     Show this help message")
	fmt.Println("  --version, -v  Show version information")
	fmt.Println()
	fmt.Println("Use 'mpu <command> --help' for more information about a command.")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  mpu run example/hello.s")
	fmt.Println("  mpu build -o game.bin game.s")
	fmt.Println("  mpu fmt mycode.s")
}

func printBuildUsage() {
	fmt.Println("Usage: mpu build [options] <file.s> [<file2.s> ...]")
	fmt.Println()
	fmt.Println("Assembles one or more assembly source files (.s) into a binary (.bin) file.")
	fmt.Println("Produces an assembly listing to stdout showing addresses and generated code.")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -o <file>    Output filename (default: first_input.bin)")
	fmt.Println("  --help       Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  mpu build program.s")
	fmt.Println("  mpu build -o game.bin main.s graphics.s sound.s")
}

func printRunUsage() {
	fmt.Println("Usage: mpu run [options] <file>")
	fmt.Println()
	fmt.Println("Executes a program file. Can run either:")
	fmt.Println("  - Binary files (.bin) - previously compiled programs")
	fmt.Println("  - Assembly files (.s) - compiles in memory then runs")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -m         Open system monitor/debugger for single-stepping")
	fmt.Println("  --help     Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  mpu run example/hello.s")
	fmt.Println("  mpu run game.bin")
	fmt.Println("  mpu run -m debug_this.s")
	fmt.Println()
	fmt.Println("Graphics programs:")
	fmt.Println("  - Press ESC to quit")
	fmt.Println("  - Press Ctrl-C in terminal for non-graphics programs")
}

func printFmtUsage() {
	fmt.Println("Usage: mpu fmt [options] <file.s>")
	fmt.Println()
	fmt.Println("Formats assembly source code with consistent style.")
	fmt.Println("Currently outputs to stdout - redirect to save formatted output.")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -w         Rewrite file in place (not yet implemented)")
	fmt.Println("  --help     Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  mpu fmt program.s")
	fmt.Println("  mpu fmt program.s > formatted.s")
}

func printTestUsage() {
	fmt.Println("Usage: mpu test [options] <file.s> [<file2.s> ...]")
	fmt.Println()
	fmt.Println("Runs unit tests found in assembly source files.")
	fmt.Println("Tests are defined with 'test FunctionName():' declarations.")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -v         Show verbose output (display all test names)")
	fmt.Println("  -color     Colorize output (default: true)")
	fmt.Println("  --help     Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  mpu test tests.s")
	fmt.Println("  mpu test -v tests/*.s")
	fmt.Println("  mpu test -color=false tests.s > results.txt")
}

func format(inputs []*os.File, rewrite bool) {
	if len(inputs) != 1 {
		fmt.Fprintf(os.Stderr, "Error: fmt command requires exactly one file\n")
		fmt.Fprintf(os.Stderr, "Usage: mpu fmt <file.s>\n")
		os.Exit(1)
	}
	if rewrite {
		fmt.Fprintf(os.Stderr, "Warning: -w flag is not yet implemented, outputting to stdout\n")
	}
	lexer := asm.NewLexer(inputs[0].Name(), inputs[0])
	parser := asm.NewParser(asm.NewInput([]asm.TokenReader{lexer}))
	parser.SetProcessInclude(false)
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
	args := flagSet.Args()

	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Error: no input files specified\n\n")
		flagSet.Usage()
		os.Exit(1)
	}

	for _, name := range args {
		f, err := os.Open(name)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "Error: file '%s' not found\n", name)
			} else if os.IsPermission(err) {
				fmt.Fprintf(os.Stderr, "Error: permission denied reading '%s'\n", name)
			} else {
				fmt.Fprintf(os.Stderr, "Error: cannot open '%s': %s\n", name, err.Error())
			}
			os.Exit(1)
		}
		inputs = append(inputs, f)
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
		fmt.Fprintf(os.Stderr, "Error: cannot mix binary and source files\n")
		fmt.Fprintf(os.Stderr, "Use either .bin files or .s files, not both\n")
		os.Exit(1)
	}
	if bin && len(inputs) > 1 {
		fmt.Fprintf(os.Stderr, "Error: only one binary file can be run at a time\n")
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
		fmt.Fprintf(os.Stderr, "Error: failed to write output file '%s': %s\n", outputName, err)
		os.Exit(1)
	}
	fmt.Printf("Successfully wrote %d bytes to %s\n", len(code), outputName)
}

func newTokenReader(inputs []*os.File) *asm.Input {
	var tr []asm.TokenReader
	for _, file := range inputs {
		tr = append(tr, asm.NewLexer(file.Name(), file))
	}
	return asm.NewInput(tr)
}

func runTests(inputs []*os.File, verbose bool, color bool) {
	// Parse all files
	parser := asm.NewParser(newTokenReader(inputs))
	parser.Parse()
	parser.PrintErrors()
	if parser.HasErrors() {
		os.Exit(1)
	}

	// Discover tests
	suite, err := test.DiscoverTests(parser.Statements())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error discovering tests: %v\n", err)
		os.Exit(1)
	}

	if len(suite.Tests) == 0 {
		fmt.Println("No tests found.")
		return
	}

	// Link the code
	linker := asm.NewLinker(parser.Statements())
	linker.Link()
	linker.PrintMessages()
	if linker.HasErrors() {
		os.Exit(1)
	}

	// Prepare code for execution
	code := linker.Code()
	if len(code) < 65536 {
		padded := make([]byte, 65536)
		copy(padded, code)
		code = padded
	}

	// Create machine and executor
	m := machine.NewMachine(code)
	executor := test.NewTestExecutor(m, suite, linker.Symbols(), linker.DebugInfo())

	// Run tests
	err = executor.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running tests: %v\n", err)
		os.Exit(1)
	}

	// Format and display results
	formatter := test.NewTerminalFormatter(verbose, color)
	formatter.Format(executor.Results(), os.Stdout)

	// Exit with appropriate code
	_, failed := executor.Summary()
	if failed > 0 {
		os.Exit(1)
	}
}
