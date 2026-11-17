// Package assembler parses JSON assembly and encodes it as machine code.
package assembler

// OpcodeLoadConstant pushes an unsigned constant onto the stack.
const OpcodeLoadConstant uint32 = 128

// OpcodeReadMemory pushes the value at a data memory address onto the stack.
const OpcodeReadMemory uint32 = 199

// OpcodeWriteMemory identifies a memory write with an address offset.
const OpcodeWriteMemory uint32 = 83

// OpcodeSubtract identifies the subtraction instruction.
const OpcodeSubtract uint32 = 165

// MaxConstant is the largest unsigned value of bits 8 through 31.
const MaxConstant uint32 = (1 << 24) - 1

// MaxAddress is the largest unsigned value of bits 8 through 24.
const MaxAddress uint32 = (1 << 17) - 1

// MaxOffset is the largest unsigned value of bits 8 through 21.
const MaxOffset uint32 = (1 << 14) - 1

// Program is an ordered list of instruction dictionaries containing A and optional B.
type Program []map[string]uint32

type commandSpec struct {
	opcode     uint32
	operand    string
	maxOperand uint32
}

const commandField = "command"

var commands = map[string]commandSpec{
	"load_const": {OpcodeLoadConstant, "value", MaxConstant},
	"read_mem":   {OpcodeReadMemory, "address", MaxAddress},
	"write_mem":  {OpcodeWriteMemory, "offset", MaxOffset},
	"sub":        {opcode: OpcodeSubtract},
}
