package tests

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"kf3/src/assembler"
	"kf3/src/interpreter"
)

func TestMachineSpecificationMemoryOperations(t *testing.T) {
	code := []byte{
		0x80, 0x2E, 0x00, 0x00,
		0x80, 0x00, 0x00, 0x00,
		0x53, 0x71, 0x03, 0x00,
		0xC7, 0x71, 0x03, 0x00,
		0x80, 0x5F, 0x00, 0x00,
		0x53, 0x00, 0x00, 0x00,
	}
	machine, err := interpreter.New(code)
	if err != nil {
		t.Fatal(err)
	}
	if err := machine.Run(); err != nil {
		t.Fatal(err)
	}
	if got := memoryDump(t, machine, 95, 95); got != "address,value\n95,46\n" {
		t.Fatalf("unexpected copied value: %q", got)
	}
	if got := memoryDump(t, machine, 881, 881); got != "address,value\n881,46\n" {
		t.Fatalf("source value changed: %q", got)
	}
}

func TestMachineOperandBoundaries(t *testing.T) {
	program := assembler.Program{
		{"A": assembler.OpcodeLoadConstant, "B": assembler.MaxConstant},
		{"A": assembler.OpcodeLoadConstant, "B": assembler.MaxAddress - assembler.MaxOffset},
		{"A": assembler.OpcodeWriteMemory, "B": assembler.MaxOffset},
		{"A": assembler.OpcodeReadMemory, "B": assembler.MaxAddress},
		{"A": assembler.OpcodeLoadConstant, "B": 0},
		{"A": assembler.OpcodeWriteMemory, "B": 0},
	}
	machine := newMachine(t, program)
	if err := machine.Run(); err != nil {
		t.Fatal(err)
	}
	for _, address := range []int{0, interpreter.MemorySize - 1} {
		want := fmt.Sprintf("address,value\n%d,%d\n", address, assembler.MaxConstant)
		if got := memoryDump(t, machine, address, address); got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	}
}

func TestMachineInstructionMemoryIsSeparate(t *testing.T) {
	program := assembler.Program{
		{"A": assembler.OpcodeReadMemory, "B": 0},
		{"A": assembler.OpcodeLoadConstant, "B": 0},
		{"A": assembler.OpcodeWriteMemory, "B": 0},
		{"A": assembler.OpcodeLoadConstant, "B": 46},
		{"A": assembler.OpcodeLoadConstant, "B": 1},
		{"A": assembler.OpcodeWriteMemory, "B": 0},
	}
	code := assembler.Encode(program)
	machine, err := interpreter.New(code)
	if err != nil {
		t.Fatal(err)
	}
	clear(code)
	if err := machine.Run(); err != nil {
		t.Fatal(err)
	}
	want := "address,value\n0,0\n1,46\n2,0\n"
	if got := memoryDump(t, machine, 0, 2); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
	if err := machine.Run(); err != nil || memoryDump(t, machine, 0, 2) != want {
		t.Fatalf("completed program was executed again: %v", err)
	}
}

func TestMachineCopyArray(t *testing.T) {
	machine := newMachine(t, readFixtureProgram(t, "copy_array"))
	if err := machine.Run(); err != nil {
		t.Fatal(err)
	}
	want := "address,value\n10,46\n11,0\n12,16777215\n"
	if got := memoryDump(t, machine, 10, 12); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
	want = "address,value\n0,46\n1,0\n2,16777215\n"
	if got := memoryDump(t, machine, 0, 2); got != want {
		t.Fatalf("source array changed: got %q, want %q", got, want)
	}
}

func newMachine(t *testing.T, program assembler.Program) *interpreter.Machine {
	t.Helper()
	machine, err := interpreter.New(assembler.Encode(program))
	if err != nil {
		t.Fatal(err)
	}
	return machine
}

func memoryDump(t *testing.T, machine *interpreter.Machine, start, end int) string {
	t.Helper()
	var output bytes.Buffer
	if err := machine.Dump(&output, start, end); err != nil {
		t.Fatal(err)
	}
	return output.String()
}

func assertExecutionError(t *testing.T, machine *interpreter.Machine, want string) {
	t.Helper()
	if err := machine.Run(); err == nil || !strings.Contains(err.Error(), want) {
		t.Fatalf("expected error containing %q, got %v", want, err)
	}
}
