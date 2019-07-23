package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"unicode"
)

// from ISO/IEC 7816-4:2005(E) 5.1.2
const values = `
0x04; DEACTIVATE FILE; Part 9
0x0C; ERASE RECORD (S); 7.3.8
0x0E, 0x0F; ERASE BINARY; 7.2.7
0x10; PERFORM SCQL OPERATION; Part 7
0x12; PERFORM TRANSACTION OPERATION; Part 7
0x14; PERFORM USER OPERATION; Part 7
0x20, 0x21; VERIFY; 7.5.6
0x22; MANAGE SECURITY ENVIRONMENT; 7.5.11
0x24; CHANGE REFERENCE DATA; 7.5.7
0x26; DISABLE VERIFICATION REQUIREMENT; 7.5.9
0x28; ENABLE VERIFICATION REQUIREMENT; 7.5.8
0x2A; PERFORM SECURITY OPERATION; Part 8
0x2C; RESET RETRY COUNTER; 7.5.10
0x44; ACTIVATE FILE; Part 9
0x46; GENERATE ASYMMETRIC KEY PAIR; Part 8
0x70; MANAGE CHANNEL; 7.1.2
0x82; EXTERNAL (/ MUTUAL) AUTHENTICATE; 7.5.4
0x84; GET CHALLENGE; 7.5.3
0x86, 0x87; GENERAL AUTHENTICATE; 7.5.5
0x88; INTERNAL AUTHENTICATE; 7.5.2
0xA0, 0xA1; SEARCH BINARY; 7.2.6
0xA2; SEARCH RECORD; 7.3.7
0xA4; SELECT; 7.1.1
0xB0, 0xB1; READ BINARY; 7.2.3
0xB2, 0xB3; READ RECORD; (S) 7.3.3
0xC0; GET RESPONSE; 7.6.1
0xC2, 0xC3; ENVELOPE; 7.6.2
0xCA, 0xCB; GET DATA; 7.4.2
0xD0, 0xD1; WRITE BINARY; 7.2.6
0xD2; WRITE RECORD; 7.3.4
0xD6, 0xD7; UPDATE BINARY; 7.2.5
0xDA, 0xDB; PUT DATA; 7.4.3
0xDC, 0xDD; UPDATE RECORD; 7.3.5
0xE0; CREATE FILE; Part 9
0xE2; APPEND RECORD; 7.3.6
0xE4; DELETE FILE; Part 9
0xE6; TERMINATE DF; Part 9
0xE8; TERMINATE EF; Part 9
0xFE; TERMINATE CARD USAGE; Part 9
`

func main() {
	if err := do(); err != nil {
		panic(err)
	}
}

func capitalize(s string) string {
	letters := []rune(s)
	letters[0] = unicode.ToTitle(letters[0])
	for i, v := range letters[1:] {
		letters[i+1] = unicode.ToLower(v)
	}

	return string(letters)
}

func toCamelCase(s string) string {
	words := strings.Fields(s)
	for i, v := range words {
		words[i] = capitalize(v)
	}

	return strings.Join(words, "")
}

func do() (err error) {
	args := os.Args[1:]

	var rawInstructions []string
	var instructions []string
	for _, v := range strings.Split(values, "\n") {
		line := strings.TrimSpace(v)
		if line == "" {
			continue
		}

		fs := strings.Split(line, ";")
		for i, v := range fs {
			fs[i] = strings.TrimSpace(v)
		}

		if len(fs) < 3 {
			panic("unexpected invalid line: " + line)
		}

		byteSequence, name, ref := fs[0], fs[1], fs[2]

		ident := toCamelCase(name)
		comment := fmt.Sprintf("%s represents the %+q instruction as defined in ISO/IEC 7816-4:2005(E) %s", ident, name, ref)

		rawInstructions = append(rawInstructions, fmt.Sprintf(
			`// RawInstr%s
RawInstr%s = [...]byte{%s}
`,
			comment, ident, byteSequence,
		))

		instructions = append(instructions, fmt.Sprintf(
			`// Instr%s
Instr%s = Instruction(RawInstr%s[:])
`, comment, ident, ident,
		))

	}

	out := fmt.Sprintf(
		`
package apdu

// Code generated DO NOT EDIT.

//go:generate go run github.com/zemnmez/cardauth/apdu/gen/instrs %s
//go:generate gofmt -w -s $GOFILE

var(
%s
)

type Instruction []byte
var (
%s
)`+"\n",
		strings.Join(args, " "),
		strings.Join(rawInstructions, "\n"),
		strings.Join(instructions, "\n"),
	)

	w := os.Stdout

	if len(args) > 1 {
		return errors.New("expecting a maximum of 1 argument")
	}

	if len(args) == 1 {
		var f *os.File
		f, err = os.OpenFile(args[0], os.O_WRONLY|os.O_TRUNC, 0700)
		if err != nil {
			return
		}
		w = f
		defer f.Close()
	}

	w.Write([]byte(out))

	return
}