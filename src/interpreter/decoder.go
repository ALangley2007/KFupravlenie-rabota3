package interpreter

import (
	"encoding/binary"
	"errors"
	"fmt"

	"kf3/src/assembler"
)

const (
	opcodeBits = 8
	opcodeMask = (1 << opcodeBits) - 1
)

type instruction struct {
	opcode  uint32
	operand uint32
}

func decode(code []byte) (instruction, error) {
	word := binary.LittleEndian.Uint32(code)
	command := instruction{opcode: word & opcodeMask, operand: word >> opcodeBits}
	switch command.opcode {
	case assembler.OpcodeLoadConstant:
		return command, nil
	case assembler.OpcodeReadMemory:
		return command, validateOperand(command.operand, assembler.MaxAddress)
	case assembler.OpcodeWriteMemory:
		return command, validateOperand(command.operand, assembler.MaxOffset)
	case assembler.OpcodeSubtract:
		return command, errors.New("sub is not supported until stage 4")
	default:
		return command, fmt.Errorf("unknown opcode %d", command.opcode)
	}
}

func validateOperand(operand, maximum uint32) error {
	if operand > maximum {
		return errors.New("nonzero unused instruction bits")
	}
	return nil
}
