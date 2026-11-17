package assembler

import "encoding/binary"

// InstructionSize is the size of a machine instruction in bytes.
const InstructionSize = 4

const operandShift = 8

// Encode packs a program returned by Parse into four-byte little-endian instructions.
func Encode(program Program) []byte {
	code := make([]byte, len(program)*InstructionSize)
	for index, instruction := range program {
		word := instruction["A"] | instruction["B"]<<operandShift
		binary.LittleEndian.PutUint32(code[index*InstructionSize:], word)
	}
	return code
}
