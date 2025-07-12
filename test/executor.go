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

package test

import (
	"fmt"

	"github.com/jsando/mpu/asm"
	"github.com/jsando/mpu/machine"
)

// TestResult holds the result of a single test execution.
type TestResult struct {
	Name    string
	Passed  bool
	Message string
	// For assertion failures
	FailureDetails []AssertionDetail
}

// AssertionDetail holds details about a single assertion failure.
type AssertionDetail struct {
	PC       uint16
	Expected int
	Actual   int
	File     string
	Line     int
}

// TestExecutor runs a test suite and collects results.
type TestExecutor struct {
	machine   *machine.Machine
	suite     *TestSuite
	symbols   *asm.SymbolTable
	debugInfo []asm.DebugInfo
	results   []TestResult
	lastError *TestResult
}

// NewTestExecutor creates a new test executor.
func NewTestExecutor(m *machine.Machine, suite *TestSuite, symbols *asm.SymbolTable, debugInfo []asm.DebugInfo) *TestExecutor {
	return &TestExecutor{
		machine:   m,
		suite:     suite,
		symbols:   symbols,
		debugInfo: debugInfo,
		results:   []TestResult{},
	}
}

// Run executes all tests in the suite.
func (e *TestExecutor) Run() error {
	// Enable test mode on the machine
	e.machine.EnableTestMode()

	for _, test := range e.suite.Tests {
		result := e.runTest(test)
		e.results = append(e.results, result)
	}

	return nil
}

// runTest executes a single test.
func (e *TestExecutor) runTest(test TestInfo) TestResult {
	// Find the test function address
	symbol := e.symbols.GetSymbol(test.Function)
	if symbol == nil {
		return TestResult{
			Name:    test.Name,
			Passed:  false,
			Message: fmt.Sprintf("test function '%s' not found in symbol table", test.Function),
		}
	}

	// Reset machine state
	e.resetMachineState()

	// Call setup if exists
	if e.suite.SetupFn != "" {
		if err := e.callFunction(e.suite.SetupFn); err != nil {
			return TestResult{
				Name:    test.Name,
				Passed:  false,
				Message: fmt.Sprintf("setup failed: %v", err),
			}
		}
	}

	// Save initial assertion count
	initialFailures := e.machine.AssertionFailures()

	// Call the test function with panic recovery
	testAddr := uint16(symbol.Value())

	// Recover from panics during test execution
	func() {
		defer func() {
			if r := recover(); r != nil {
				// Find the PC where the error occurred
				pc := e.machine.Memory().GetWord(machine.PCAddr)
				file, line := e.findSourceLocation(pc)

				var msg string
				switch err := r.(type) {
				case error:
					msg = err.Error()
				default:
					msg = fmt.Sprintf("%v", r)
				}

				// Store the error for later retrieval
				e.lastError = &TestResult{
					Name:    test.Name,
					Passed:  false,
					Message: fmt.Sprintf("runtime error: %s", msg),
					FailureDetails: []AssertionDetail{{
						PC:   pc,
						File: file,
						Line: line,
					}},
				}
			}
		}()

		e.callAddress(testAddr)
	}()

	// If we caught a panic, return the error result
	if e.lastError != nil {
		result := *e.lastError
		e.lastError = nil
		return result
	}

	// Call teardown if exists
	if e.suite.TeardownFn != "" {
		e.callFunction(e.suite.TeardownFn)
	}

	// Check if test passed
	failures := e.machine.AssertionFailures() - initialFailures
	if failures > 0 {
		// Get failure details
		var failureDetails []AssertionDetail
		if lastFailure := e.machine.LastAssertionFailure(); lastFailure != nil {
			// Find source location from debug info
			file, line := e.findSourceLocation(lastFailure.PC)
			failureDetails = append(failureDetails, AssertionDetail{
				PC:       lastFailure.PC,
				Expected: lastFailure.Expected,
				Actual:   lastFailure.Actual,
				File:     file,
				Line:     line,
			})
		}

		return TestResult{
			Name:           test.Name,
			Passed:         false,
			Message:        fmt.Sprintf("%d assertion(s) failed", failures),
			FailureDetails: failureDetails,
		}
	}

	return TestResult{
		Name:   test.Name,
		Passed: true,
	}
}

// resetMachineState resets the machine to a clean state for each test.
func (e *TestExecutor) resetMachineState() {
	// Reset SP to top of memory
	e.machine.Memory().PutWord(machine.SPAddr, 0xFFFF)
	// Reset FP
	e.machine.Memory().PutWord(machine.FPAddr, 0)
	// Clear carry and other flags by executing CLC, CLB
	// This is a simplified reset - might need more comprehensive reset later
}

// callFunction calls a function by name.
func (e *TestExecutor) callFunction(name string) error {
	symbol := e.symbols.GetSymbol(name)
	if symbol == nil {
		return fmt.Errorf("function '%s' not found", name)
	}

	addr := uint16(symbol.Value())
	e.callAddress(addr)
	return nil
}

// callAddress calls a function at the given address using JSR/RET convention.
func (e *TestExecutor) callAddress(addr uint16) {
	// Get current PC
	pc := e.machine.Memory().GetWord(machine.PCAddr)

	// Push return address (current PC + 3 for JSR instruction)
	sp := e.machine.Memory().GetWord(machine.SPAddr)
	sp -= 2
	e.machine.Memory().PutWord(sp, pc+3)
	e.machine.Memory().PutWord(machine.SPAddr, sp)

	// Jump to test function
	e.machine.Memory().PutWord(machine.PCAddr, addr)

	// Run until RET
	e.machine.Run()
}

// Results returns the test results.
func (e *TestExecutor) Results() []TestResult {
	return e.results
}

// Summary returns a summary of test results.
func (e *TestExecutor) Summary() (passed, failed int) {
	for _, result := range e.results {
		if result.Passed {
			passed++
		} else {
			failed++
		}
	}
	return
}

// findSourceLocation finds the source file and line for a given PC.
func (e *TestExecutor) findSourceLocation(pc uint16) (string, int) {
	// Search debug info for the closest PC match
	for i := len(e.debugInfo) - 1; i >= 0; i-- {
		if e.debugInfo[i].PC <= pc {
			return e.debugInfo[i].File, e.debugInfo[i].Line
		}
	}
	return "", 0
}
