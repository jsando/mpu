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
	"io"
	"os"
	"strings"
)

// Formatter is an interface for formatting test results.
type Formatter interface {
	Format(results []TestResult, w io.Writer) error
}

// TerminalFormatter formats results for terminal output.
type TerminalFormatter struct {
	Verbose      bool
	Color        bool
	sourceReader *SourceReader
}

// NewTerminalFormatter creates a new terminal formatter.
func NewTerminalFormatter(verbose, color bool) *TerminalFormatter {
	return &TerminalFormatter{
		Verbose:      verbose,
		Color:        color,
		sourceReader: NewSourceReader(),
	}
}

// Format writes formatted test results to the writer.
func (f *TerminalFormatter) Format(results []TestResult, w io.Writer) error {
	if w == nil {
		w = os.Stdout
	}

	for _, result := range results {
		if result.Passed {
			if f.Color {
				fmt.Fprintf(w, "\033[32m✓\033[0m %s\n", result.Name)
			} else {
				fmt.Fprintf(w, "✓ %s\n", result.Name)
			}
		} else {
			if f.Color {
				fmt.Fprintf(w, "\033[31m✗\033[0m %s\n", result.Name)
			} else {
				fmt.Fprintf(w, "✗ %s\n", result.Name)
			}
			if result.Message != "" && (f.Verbose || strings.Contains(result.Message, "runtime error")) {
				fmt.Fprintf(w, "  %s\n", result.Message)
			}

			// Show failure details with source
			for _, detail := range result.FailureDetails {
				if detail.File != "" && detail.Line > 0 {
					fmt.Fprintf(w, "\n  at %s:%d\n", detail.File, detail.Line)

					// Get source lines with context
					lines, startLine, err := f.sourceReader.GetContext(detail.File, detail.Line, 1, 1)
					if err == nil {
						for i, line := range lines {
							lineNum := startLine + i
							if lineNum == detail.Line {
								// Highlight the failing line
								if f.Color {
									fmt.Fprintf(w, "\033[31m%s\033[0m\n", FormatSourceLine(lineNum, line, true))
								} else {
									fmt.Fprintf(w, "%s  <-- assertion failed\n", FormatSourceLine(lineNum, line, true))
								}
							} else {
								fmt.Fprintf(w, "%s\n", FormatSourceLine(lineNum, line, false))
							}
						}
					}

					// Only show expected/actual for assertion failures
					if !strings.Contains(result.Message, "runtime error") {
						fmt.Fprintf(w, "\n  Expected: %d\n", detail.Expected)
						fmt.Fprintf(w, "  Actual:   %d\n", detail.Actual)
					}
				}
			}
		}
	}

	// Summary
	passed, failed := 0, 0
	for _, r := range results {
		if r.Passed {
			passed++
		} else {
			failed++
		}
	}

	fmt.Fprintf(w, "\nTests: ")
	if passed > 0 {
		if f.Color {
			fmt.Fprintf(w, "\033[32m%d passed\033[0m", passed)
		} else {
			fmt.Fprintf(w, "%d passed", passed)
		}
	}
	if failed > 0 {
		if passed > 0 {
			fmt.Fprintf(w, ", ")
		}
		if f.Color {
			fmt.Fprintf(w, "\033[31m%d failed\033[0m", failed)
		} else {
			fmt.Fprintf(w, "%d failed", failed)
		}
	}
	fmt.Fprintf(w, ", %d total\n", passed+failed)

	return nil
}
