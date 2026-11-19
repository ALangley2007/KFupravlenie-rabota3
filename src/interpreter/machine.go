// Package interpreter executes machine code using separate instruction and data memory.
package interpreter

import (
	"errors"
	"fmt"

	"kf3/src/assembler"
)

// MemorySize is the number of data cells addressable by read_mem.
const MemorySize = int(assembler.MaxAddress) + 1

const writeOperands = 2

// Machine holds private instruction memory, zero-initialized uint32 data cells, and a stack.
// It must be constructed with New.
type Machine struct {
	code  []byte
	data  []uint32
	stack []uint32
	pc    int
}

// New copies the binary program and allocates data memory independently of its bytes.
// A program must contain a whole number of four-byte instructions; an empty program is valid.
func New(code []byte) (*Machine, error) {
	if len(code)%assembler.InstructionSize != 0 {
		return nil, errors.New("binary size must be a multiple of 4 bytes")
	}
	return &Machine{
		code: append([]byte(nil), code...),
		data: make([]uint32, MemorySize),
	}, nil
}

// Run fetches, decodes, and executes instructions until the program ends or an error occurs.
// Errors identify the instruction and byte offset. A failed instruction does not change state.
// Further calls resume at the current instruction; completed programs do nothing.
func (machine *Machine) Run() error {
	for machine.pc < len(machine.code) {
		end := machine.pc + assembler.InstructionSize
		command, err := decode(machine.code[machine.pc:end])
		if err == nil {
			err = machine.execute(command)
		}
		if err != nil {
			index := machine.pc/assembler.InstructionSize + 1
			return fmt.Errorf("instruction %d (byte %d): %w", index, machine.pc, err)
		}
		machine.pc = end
	}
	return nil
}

func (machine *Machine) execute(command instruction) error {
	switch command.opcode {
	case assembler.OpcodeLoadConstant:
		machine.stack = append(machine.stack, command.operand)
	case assembler.OpcodeReadMemory:
		machine.stack = append(machine.stack, machine.data[command.operand])
	case assembler.OpcodeWriteMemory:
		return machine.writeMemory(command.operand)
	}
	return nil
}

func (machine *Machine) writeMemory(offset uint32) error {
	if len(machine.stack) < writeOperands {
		return errors.New("write_mem requires a value and a base address on the stack")
	}
	top := len(machine.stack) - 1
	address := uint64(machine.stack[top]) + uint64(offset)
	if address >= uint64(MemorySize) {
		return fmt.Errorf("write_mem address %d is outside data memory 0..%d", address, MemorySize-1)
	}
	machine.data[address] = machine.stack[top-1]
	machine.stack = machine.stack[:len(machine.stack)-writeOperands]
	return nil
}
