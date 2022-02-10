package asm

import (
	"bytes"
	"testing"
)

func TestSource_GetTokenImage(t *testing.T) {
	s := NewSource("input.txt", bytes.NewReader([]byte("foo")))
	if s.NextCh() == 'f' && s.NextCh() == 'o' && s.NextCh() == 'o' && s.NextCh() == -1 {
		// all good
	} else {
		t.Fatal("unexpected data")
	}
}
