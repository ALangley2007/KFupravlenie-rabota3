package tests

import (
	"bytes"
	"testing"

	"kf3/src/assembler"
	"kf3/src/interpreter"
)

const executionAttempts = 2

func TestMachineRejectsIncompleteBinary(t *testing.T) {
	for _, size := range []int{1, 2, 3, 5, 6, 7} {
		if machine, err := interpreter.New(make([]byte, size)); err == nil || machine != nil {
			t.Fatalf("expected error for %d bytes, got %v, %v", size, machine, err)
		}
	}
}

func TestMachineRejectsInvalidInstructions(t *testing.T) {
	cases := []struct {
		name    string
		command map[string]uint32
		message string
	}{
		{"unknown", map[string]uint32{"A": 0}, "unknown opcode 0"},
		{"read bits", map[string]uint32{"A": assembler.OpcodeReadMemory, "B": assembler.MaxAddress + 1},
			"nonzero unused instruction bits"},
		{"write bits", map[string]uint32{"A": assembler.OpcodeWriteMemory, "B": assembler.MaxOffset + 1},
			"nonzero unused instruction bits"},
		{"sub", map[string]uint32{"A": assembler.OpcodeSubtract}, "sub is not supported until stage 4"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			machine := newMachine(t, assembler.Program{{"A": assembler.OpcodeLoadConstant}, tc.command})
			assertExecutionError(t, machine, "instruction 2 (byte 4): "+tc.message)
			if got := memoryDump(t, machine, 0, 0); got != "address,value\n0,0\n" {
				t.Fatalf("invalid instruction changed memory: %q", got)
			}
		})
	}
}

func TestMachineWriteRequiresTwoStackElements(t *testing.T) {
	cases := []assembler.Program{
		{{"A": assembler.OpcodeWriteMemory}},
		{{"A": assembler.OpcodeLoadConstant}, {"A": assembler.OpcodeWriteMemory}},
		{
			{"A": assembler.OpcodeLoadConstant, "B": 46},
			{"A": assembler.OpcodeLoadConstant},
			{"A": assembler.OpcodeWriteMemory},
			{"A": assembler.OpcodeWriteMemory},
		},
	}
	for _, program := range cases {
		machine := newMachine(t, program)
		assertExecutionError(t, machine, "write_mem requires a value and a base address on the stack")
	}
}

func TestMachineInvalidWritePreservesState(t *testing.T) {
	cases := [][2]uint32{
		{assembler.MaxAddress, 1},
		{uint32(interpreter.MemorySize), 0},
		{assembler.MaxConstant, assembler.MaxOffset},
	}
	for _, operands := range cases {
		program := assembler.Program{
			{"A": assembler.OpcodeLoadConstant, "B": 46},
			{"A": assembler.OpcodeLoadConstant},
			{"A": assembler.OpcodeWriteMemory},
			{"A": assembler.OpcodeLoadConstant, "B": 95},
			{"A": assembler.OpcodeLoadConstant, "B": operands[0]},
			{"A": assembler.OpcodeWriteMemory, "B": operands[1]},
		}
		machine := newMachine(t, program)
		for attempt := 0; attempt < executionAttempts; attempt++ {
			assertExecutionError(t, machine, "instruction 6 (byte 20): write_mem address")
			if got := memoryDump(t, machine, 0, 1); got != "address,value\n0,46\n1,0\n" {
				t.Fatalf("invalid write changed memory: %q", got)
			}
		}
	}
}

func TestMachineEmptyProgramAndDumpRange(t *testing.T) {
	machine := newMachine(t, assembler.Program{})
	if err := machine.Run(); err != nil {
		t.Fatal(err)
	}
	if got := memoryDump(t, machine, 0, 1); got != "address,value\n0,0\n1,0\n" {
		t.Fatalf("data memory is not initialized to zero: %q", got)
	}
	for _, bounds := range [][2]int{{-1, 0}, {1, 0}, {0, interpreter.MemorySize}} {
		var output bytes.Buffer
		if err := machine.Dump(&output, bounds[0], bounds[1]); err == nil || output.Len() != 0 {
			t.Fatalf("invalid range %v: error=%v output=%q", bounds, err, &output)
		}
	}
	if err := machine.Dump(failingWriter{}, 0, 0); err == nil {
		t.Fatal("CSV output failure was ignored")
	}
	if err := machine.Dump(failingWriter{}, 0, interpreter.MemorySize-1); err == nil {
		t.Fatal("CSV row failure was ignored")
	}
}
