package tests

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"kf3/src/assembler"
)

func TestEncodeSpecification(t *testing.T) {
	cases := []struct {
		name string
		want []byte
	}{
		{"load_const", []byte{0x80, 0x2E, 0x00, 0x00}},
		{"read_mem", []byte{0xC7, 0x5F, 0x00, 0x00}},
		{"write_mem", []byte{0x53, 0x71, 0x03, 0x00}},
		{"sub", []byte{0xA5, 0x00, 0x00, 0x00}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := assembler.Encode(readFixtureProgram(t, tc.name))
			if !bytes.Equal(got, tc.want) {
				t.Fatalf("got % X, want % X", got, tc.want)
			}
		})
	}
}

func TestEncodeAllCommands(t *testing.T) {
	want := []byte{
		0x80, 0x2E, 0x00, 0x00,
		0xC7, 0x5F, 0x00, 0x00,
		0x53, 0x71, 0x03, 0x00,
		0xA5, 0x00, 0x00, 0x00,
	}
	got := assembler.Encode(readFixtureProgram(t, "all_commands"))
	if !bytes.Equal(got, want) {
		t.Fatalf("got % X, want % X", got, want)
	}
}

func TestEncodeOperandBoundaries(t *testing.T) {
	cases := []struct {
		name    string
		opcode  uint32
		operand uint32
		want    []byte
	}{
		{"constant/zero", assembler.OpcodeLoadConstant, 0, []byte{0x80, 0x00, 0x00, 0x00}},
		{"constant/max", assembler.OpcodeLoadConstant, assembler.MaxConstant, []byte{0x80, 0xFF, 0xFF, 0xFF}},
		{"address/zero", assembler.OpcodeReadMemory, 0, []byte{0xC7, 0x00, 0x00, 0x00}},
		{"address/max", assembler.OpcodeReadMemory, assembler.MaxAddress, []byte{0xC7, 0xFF, 0xFF, 0x01}},
		{"offset/zero", assembler.OpcodeWriteMemory, 0, []byte{0x53, 0x00, 0x00, 0x00}},
		{"offset/max", assembler.OpcodeWriteMemory, assembler.MaxOffset, []byte{0x53, 0xFF, 0x3F, 0x00}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			program := assembler.Program{{"A": tc.opcode, "B": tc.operand}}
			if got := assembler.Encode(program); !bytes.Equal(got, tc.want) {
				t.Fatalf("got % X, want % X", got, tc.want)
			}
		})
	}
}

func TestEncodeEmptyProgram(t *testing.T) {
	if got := assembler.Encode(assembler.Program{}); len(got) != 0 {
		t.Fatalf("expected no bytes, got % X", got)
	}
}

func readFixtureProgram(t *testing.T, name string) assembler.Program {
	t.Helper()
	source, err := os.ReadFile(filepath.Join("fixtures", name+".json"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := assembler.Parse(bytes.NewReader(source))
	if err != nil {
		t.Fatal(err)
	}
	return program
}
