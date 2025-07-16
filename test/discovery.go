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
	"github.com/jsando/mpu/asm"
)

// TestInfo holds information about a single test function.
type TestInfo struct {
	Name     string
	Function string // Global symbol name for the test function
	File     string
	Line     int
}

// TestSuite represents a collection of tests found in parsed assembly code.
type TestSuite struct {
	Tests      []TestInfo
	SetupFn    string // Optional setup function name
	TeardownFn string // Optional teardown function name
}

// DiscoverTests scans the AST for test functions and returns a TestSuite.
func DiscoverTests(statements asm.Statement) (*TestSuite, error) {
	suite := &TestSuite{
		Tests: []TestInfo{},
	}

	// Walk through all statements looking for TestStatement nodes
	for stmt := statements; stmt != nil; stmt = stmt.Next() {
		switch s := stmt.(type) {
		case *asm.TestStatement:
			test := TestInfo{
				Name:     s.Name(),
				Function: s.Name(),
				File:     s.File(),
				Line:     s.Line(),
			}
			suite.Tests = append(suite.Tests, test)

		case *asm.LabelStatement:
			// Check for special test setup/teardown functions
			name := s.Name()
			if name == "test_setup" {
				suite.SetupFn = name
			} else if name == "test_teardown" {
				suite.TeardownFn = name
			}
		}
	}

	return suite, nil
}
