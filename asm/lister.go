package asm

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"io"
)

// WriteListing prints the generated bytes next to the original source, ie
// an assembly listing.
// Format:
// line address bytes original source line
func WriteListing(in io.Reader, out io.Writer, linker *Linker) error {
	scanner := bufio.NewScanner(in)
	line := 1
	frag := linker.fragment
	pc := 0
	for scanner.Scan() {
		if frag != nil {
			pc = frag.pcStart
		}
		text := scanner.Text()
		if frag == nil || line < frag.line {
			_, err := fmt.Fprintf(out, "%4d: 0x%04x  %-12s %s\n", line, pc, "", text)
			if err != nil {
				return err
			}
			line++
			continue
		}
		if frag != nil {
			bytesFor := linker.BytesFor(frag)
			if len(bytesFor) > 5 {
				bytesFor = bytesFor[:5]
			}
			bytes := hex.EncodeToString(bytesFor)
			_, err := fmt.Fprintf(out, "%4d: 0x%04x  %-12s %s\n", line, pc, bytes, text)
			if err != nil {
				return err
			}
			line++
			frag = frag.next
		}
	}
	return nil
}
