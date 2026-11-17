package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"kf3/src/assembler"
)

func writeBinary(path string, code []byte) error {
	if err := os.WriteFile(path, code, 0o644); err != nil {
		return fmt.Errorf("write binary: %w", err)
	}
	return nil
}

func printResult(writer io.Writer, program assembler.Program, code []byte, testMode bool) error {
	if testMode {
		if err := printProgram(writer, program); err != nil {
			return err
		}
		if err := printBytes(writer, code); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(writer, "Size: %d bytes\n", len(code)); err != nil {
		return fmt.Errorf("print file size: %w", err)
	}
	return nil
}

func printProgram(writer io.Writer, program assembler.Program) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(program); err != nil {
		return fmt.Errorf("print intermediate representation: %w", err)
	}
	return nil
}

func printBytes(writer io.Writer, code []byte) error {
	if _, err := fmt.Fprintln(writer, "Bytes:"); err != nil {
		return fmt.Errorf("print machine-code bytes: %w", err)
	}
	for offset := 0; offset < len(code); offset += assembler.InstructionSize {
		_, err := fmt.Fprintf(writer, "0x%02X, 0x%02X, 0x%02X, 0x%02X\n",
			code[offset], code[offset+1], code[offset+2], code[offset+3])
		if err != nil {
			return fmt.Errorf("print machine-code bytes: %w", err)
		}
	}
	return nil
}
