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

package asm

import "fmt"

const (
	MessageError   = iota
	MessageWarning = iota
	MessageInfo    = iota
)

// Message describes a compiler error, warning, or informational message.
type Message struct {
	messageType int
	file        string
	line        int // 1-based line number
	column      int // 1-based column number
	message     string
}

func (m Message) String() string {
	return fmt.Sprintf("%s:%d:%d %s", m.file, m.line, m.column, m.message)
}

// Messages is a collection of compiler messages.
type Messages struct {
	errors   int
	warnings int
	messages []Message
}

func (m *Messages) AddMessage(messageType int, file string, line, column int, message string) {
	msg := Message{
		messageType: messageType,
		file:        file,
		line:        line,
		column:      column,
		message:     message,
	}
	m.messages = append(m.messages, msg)
	if messageType == MessageError {
		m.errors++
	}
	if messageType == MessageWarning {
		m.warnings++
	}
}

func (m *Messages) Warn(file string, line int, column int, message string) {
	m.AddMessage(MessageWarning, file, line, column, message)
}

func (m *Messages) Error(file string, line int, column int, message string) {
	m.AddMessage(MessageError, file, line, column, message)
}

func (m *Messages) Info(file string, line int, column int, message string) {
	m.AddMessage(MessageInfo, file, line, column, message)
}

func (m *Messages) Print() {
	for _, msg := range m.messages {
		fmt.Println(msg)
	}
	if m.warnings > 0 {
		fmt.Printf("%d warnings.\n", m.warnings)
	}
	if m.errors > 0 {
		fmt.Printf("%d errors.\n", m.errors)
	}
}
