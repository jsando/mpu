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
	"bufio"
	"fmt"
	"os"
)

// SourceReader reads source lines from files for display in test output.
type SourceReader struct {
	cache map[string][]string
}

// NewSourceReader creates a new source reader.
func NewSourceReader() *SourceReader {
	return &SourceReader{
		cache: make(map[string][]string),
	}
}

// GetLine reads a specific line from a file (1-indexed).
func (s *SourceReader) GetLine(filename string, lineNum int) (string, error) {
	lines, err := s.getLines(filename)
	if err != nil {
		return "", err
	}

	if lineNum < 1 || lineNum > len(lines) {
		return "", fmt.Errorf("line %d out of range (file has %d lines)", lineNum, len(lines))
	}

	return lines[lineNum-1], nil
}

// GetContext returns lines around a specific line with context.
func (s *SourceReader) GetContext(filename string, lineNum int, before, after int) ([]string, int, error) {
	lines, err := s.getLines(filename)
	if err != nil {
		return nil, 0, err
	}

	if lineNum < 1 || lineNum > len(lines) {
		return nil, 0, fmt.Errorf("line %d out of range", lineNum)
	}

	start := lineNum - before - 1
	if start < 0 {
		start = 0
	}

	end := lineNum + after
	if end > len(lines) {
		end = len(lines)
	}

	return lines[start:end], start + 1, nil
}

// getLines reads all lines from a file, using cache if available.
func (s *SourceReader) getLines(filename string) ([]string, error) {
	if lines, ok := s.cache[filename]; ok {
		return lines, nil
	}

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	s.cache[filename] = lines
	return lines, nil
}

// FormatSourceLine formats a source line for display with line number.
func FormatSourceLine(lineNum int, line string, highlight bool) string {
	if highlight {
		return fmt.Sprintf("  %4d | %s", lineNum, line)
	}
	return fmt.Sprintf("  %4d | %s", lineNum, line)
}
