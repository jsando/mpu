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

import (
	"fmt"
	"io"
)

// Source represents a source file and the current state of lexing it.
type Source struct {
	messages           *Messages
	filename           string
	reader             io.Reader
	buffer             []byte
	count              int
	next               int
	lastCh             int
	lastLineStart      int
	currentLineStart   int
	currentLineNumber  int
	previousLineStart  int // offset into buffer at which line containing last token started
	previousLineNumber int
	currentTokenStart  int
	currentTokenEnd    int
	previousTokenStart int
	previousTokenEnd   int
	endOfInput         bool
	gobbledCR          bool // whether it automatically skipped a CR to get to a LF
}

const readSize = 32768

func NewSource(filename string, reader io.Reader) *Source {
	source := &Source{
		messages: &Messages{},
		filename: filename,
		reader:   reader,
		buffer:   make([]byte, readSize+512),
	}
	return source
}

func (s *Source) fill() error {
	if s.endOfInput {
		return io.EOF
	}
	// Move existing data to start of buffer to make room at the end
	if s.count > 0 && s.previousTokenStart > 0 {
		delta := s.previousTokenStart
		copy(s.buffer[:], s.buffer[delta:s.count])
		s.previousLineStart -= delta
		s.previousTokenStart -= delta
		s.previousTokenEnd -= delta
		s.lastLineStart -= delta
		s.currentLineStart -= delta
		s.currentTokenStart -= delta
		s.currentTokenEnd -= delta
		s.count -= delta
		s.next -= delta
	}
	// Then read more data if we can.
	read, err := s.reader.Read(s.buffer[s.count:])
	s.count += read
	if err == io.EOF {
		s.endOfInput = true
	}
	return err
}

func (s *Source) NextCh() int {
	s.gobbledCR = false
	//if (lineListener != null) {
	//	if (lastCh == 0 || lastCh == '\n') {
	//		lineListener.lineStarted (buffer, next, count);
	//	}
	//}
	if s.next >= s.count {
		if s.endOfInput {
			s.lastCh = -1
			return s.lastCh
		}
		err := s.fill()
		if err != nil {
			s.error(fmt.Sprintf("I/O error reading from file (%s)", err.Error()))
			s.lastCh = -1
			return s.lastCh
		}
	}
	ch := int(s.buffer[s.next])
	s.next++
	if ch == '\n' {
		s.currentLineNumber++
		s.lastLineStart = s.currentLineStart
		s.currentLineStart = s.next
	} else if ch == '\r' {
		// Look ahead to swallow the LF if it exists
		if s.next >= s.count && !s.endOfInput {
			err := s.fill()
			if err != nil {
				s.error(fmt.Sprintf("I/O error reading from file: %s", err.Error()))
				s.lastCh = -1
				return s.lastCh
			}
		}
		if s.next < s.count {
			nch := s.buffer[s.next]
			s.next++
			if nch == '\n' {
				ch = '\n'
				s.gobbledCR = true
			} else {
				s.next--
			}
		}
		s.currentLineNumber++
		s.lastLineStart = s.currentLineStart
		s.currentLineStart = s.next
	}
	s.lastCh = ch
	return s.lastCh
}

// SkipToEOL advances to the EOL marker (either CR or CRLF), such that the next call
// will return the character following the EOL.
func (s *Source) SkipToEOL() {
	for ch := s.NextCh(); ch != -1 && ch != '\n'; ch = s.NextCh() {
	}
}

// Backup one character.
func (s *Source) Backup() {
	if s.next > 0 {
		s.next--
		if s.gobbledCR {
			s.next--
			s.gobbledCR = false
		}
		if s.lastCh == '\n' {
			s.currentLineNumber--
			s.currentLineStart = s.lastLineStart
			s.lastCh = -1
		}
	}
}

// MarkStart sets the last read character as the start of a token.
func (s *Source) MarkStart() {
	s.previousLineNumber = s.currentLineNumber
	s.previousLineStart = s.currentLineStart
	s.previousTokenStart = s.currentTokenStart
	s.previousTokenEnd = s.currentTokenEnd
	mark := s.next
	if mark > 0 {
		mark--
		if s.gobbledCR {
			mark--
		}
	}
	s.currentTokenEnd = mark
	s.currentTokenStart = mark
}

// MarkEnd sets the last read character as the end of the current token.
func (s *Source) MarkEnd() {
	mark := s.next
	if mark > 0 {
		mark--
		if s.gobbledCR {
			mark--
		}
	}
	s.currentTokenEnd = mark
}

// GetTokenImage gets the marked token as a string.
func (s *Source) GetTokenImage() string {
	return string(s.buffer[s.currentTokenStart:s.currentTokenEnd])
}

// Error reporting helper methods
func (s *Source) info(message string) {
	column := s.currentTokenStart - s.currentLineStart + 1
	s.messages.Info(s.filename, s.currentLineNumber, column, message)
}

func (s *Source) warn(message string) {
	column := s.currentTokenStart - s.currentLineStart + 1
	s.messages.Warn(s.filename, s.currentLineNumber, column, message)
}

func (s *Source) error(message string) {
	column := s.currentTokenStart - s.currentLineStart + 1
	s.messages.Error(s.filename, s.currentLineNumber, column, message)
}
