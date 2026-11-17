package tests

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"kf3/src/assembler"
)

func TestSpecification(t *testing.T) {
	cases := []struct {
		name string
		want assembler.Program
	}{
		{"load_const", assembler.Program{{"A": 128, "B": 46}}},
		{"read_mem", assembler.Program{{"A": 199, "B": 95}}},
		{"write_mem", assembler.Program{{"A": 83, "B": 881}}},
		{"sub", assembler.Program{{"A": 165}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			source, err := os.ReadFile(filepath.Join("fixtures", tc.name+".json"))
			if err != nil {
				t.Fatal(err)
			}
			got, err := assembler.Parse(strings.NewReader(string(source)))
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestInstructionOrder(t *testing.T) {
	source := `[
		{"command":"read_mem","address":95},
		{"command":"load_const","value":46},
		{"command":"sub"},
		{"command":"write_mem","offset":881},
		{"command":"load_const","value":0}
	]`
	want := assembler.Program{
		{"A": assembler.OpcodeReadMemory, "B": 95},
		{"A": assembler.OpcodeLoadConstant, "B": 46},
		{"A": assembler.OpcodeSubtract},
		{"A": assembler.OpcodeWriteMemory, "B": 881},
		{"A": assembler.OpcodeLoadConstant, "B": 0},
	}
	got, err := assembler.Parse(strings.NewReader(source))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestOperandBoundaries(t *testing.T) {
	for _, spec := range operandCases() {
		for _, value := range []uint32{0, spec.max} {
			name := fmt.Sprintf("%s/%d", spec.command, value)
			t.Run(name, func(t *testing.T) {
				source := operandSource(spec, fmt.Sprint(value))
				got, err := assembler.Parse(strings.NewReader(source))
				if err != nil {
					t.Fatal(err)
				}
				if got[0]["B"] != value {
					t.Fatalf("got operand %d, want %d", got[0]["B"], value)
				}
			})
		}
	}
}

func TestRejectInvalidOperands(t *testing.T) {
	for _, spec := range operandCases() {
		values := []string{"-1", "1.5", "1e2", `"46"`, "true", "null", "4294967296"}
		values = append(values, fmt.Sprint(spec.max+1))
		for _, value := range values {
			t.Run(spec.command+"/"+value, func(t *testing.T) {
				assertParseError(t, operandSource(spec, value))
			})
		}
	}
}

func TestRejectInvalidPrograms(t *testing.T) {
	for name, source := range invalidSourceCases() {
		t.Run(name, func(t *testing.T) {
			assertParseError(t, source)
		})
	}
}

func TestEmptyProgram(t *testing.T) {
	got, err := assembler.Parse(strings.NewReader("[] \n\t"))
	if err != nil {
		t.Fatal(err)
	}
	if got == nil || len(got) != 0 {
		t.Fatalf("expected an empty non-nil program, got %v", got)
	}
}

func TestSourceErrorLocation(t *testing.T) {
	source := `[{"command":"sub"},{"command":"load_const","value":-1}]`
	got, err := assembler.Parse(strings.NewReader(source))
	if err == nil || !strings.Contains(err.Error(), "instruction 2:") {
		t.Fatalf("expected the second instruction in the error, got %v", err)
	}
	if got != nil {
		t.Fatalf("invalid input returned a partial program: %v", got)
	}
}

type operandCase struct {
	command string
	field   string
	max     uint32
}

func operandCases() []operandCase {
	return []operandCase{
		{"load_const", "value", 16777215},
		{"read_mem", "address", 131071},
		{"write_mem", "offset", 16383},
	}
}

func operandSource(spec operandCase, value string) string {
	return fmt.Sprintf(`[{"command":%q,%q:%s}]`, spec.command, spec.field, value)
}

func assertParseError(t *testing.T, source string) {
	t.Helper()
	if got, err := assembler.Parse(strings.NewReader(source)); err == nil {
		t.Fatalf("expected an error for %q, got %v", source, got)
	}
}

func invalidSourceCases() map[string]string {
	return map[string]string{
		"empty source":       "",
		"null source":        "null",
		"object source":      "{}",
		"scalar source":      "123",
		"malformed JSON":     `[{"command":"sub"}`,
		"trailing comma":     `[{"command":"sub"},]`,
		"trailing document":  `[] []`,
		"trailing garbage":   `[] invalid`,
		"string instruction": `["sub"]`,
		"null instruction":   `[null]`,
		"missing command":    `[{}]`,
		"numeric command":    `[{"command":128}]`,
		"null command":       `[{"command":null}]`,
		"unknown command":    `[{"command":"multiply"}]`,
		"wrong command case": `[{"command":"SUB"}]`,
		"missing value":      `[{"command":"load_const"}]`,
		"missing address":    `[{"command":"read_mem"}]`,
		"missing offset":     `[{"command":"write_mem"}]`,
		"unknown field":      `[{"command":"load_const","value":46,"other":0}]`,
		"foreign operand":    `[{"command":"read_mem","address":95,"value":46}]`,
		"null extra field":   `[{"command":"write_mem","offset":881,"address":null}]`,
		"sub operand":        `[{"command":"sub","value":0}]`,
		"sub empty key":      `[{"command":"sub","":0}]`,
	}
}
