package tests

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"kf3/src/cli"
)

func TestCLIPrintsAllCommandBytes(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "program.bin")
	args := []string{"-input", "fixtures/all_commands.json", "-output", outputPath, "-test"}
	var stdout, stderr bytes.Buffer
	if code := cli.Run(args, &stdout, &stderr); code != successExitCode {
		t.Fatalf("exit code %d: %s", code, &stderr)
	}
	want := "Bytes:\n" +
		"0x80, 0x2E, 0x00, 0x00\n" +
		"0xC7, 0x5F, 0x00, 0x00\n" +
		"0x53, 0x71, 0x03, 0x00\n" +
		"0xA5, 0x00, 0x00, 0x00\n" +
		"Size: 16 bytes\n"
	if !strings.HasSuffix(stdout.String(), want) || stderr.Len() != 0 {
		t.Fatalf("unexpected output: stdout=%q stderr=%q", &stdout, &stderr)
	}
}

func TestCLIEmptyProgram(t *testing.T) {
	directory := t.TempDir()
	inputPath := filepath.Join(directory, "empty.json")
	if err := os.WriteFile(inputPath, []byte("[]"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"-test=true", "-test=false"} {
		t.Run(mode, func(t *testing.T) {
			outputPath := filepath.Join(directory, "empty.bin")
			args := []string{"-input", inputPath, "-output", outputPath, mode}
			var stdout, stderr bytes.Buffer
			if code := cli.Run(args, &stdout, &stderr); code != successExitCode {
				t.Fatalf("exit code %d: %s", code, &stderr)
			}
			want := "Size: 0 bytes\n"
			if mode == "-test=true" {
				want = "[]\nBytes:\n" + want
			}
			if stdout.String() != want || stderr.Len() != 0 {
				t.Fatalf("unexpected output: stdout=%q stderr=%q", &stdout, &stderr)
			}
			assertFileBytes(t, outputPath, []byte{})
		})
	}
}

func TestCLIWriteErrors(t *testing.T) {
	directory := t.TempDir()
	missingParent := filepath.Join(directory, "missing")
	for _, path := range []string{directory, filepath.Join(missingParent, "program.bin")} {
		t.Run(filepath.Base(path), func(t *testing.T) {
			args := []string{"-input", "fixtures/sub.json", "-output", path, "-test"}
			var stdout, stderr bytes.Buffer
			if code := cli.Run(args, &stdout, &stderr); code != failureExitCode {
				t.Fatalf("expected exit code 1, got %d", code)
			}
			if stdout.Len() != 0 || !strings.Contains(stderr.String(), "write binary:") {
				t.Fatalf("invalid diagnostics: stdout=%q stderr=%q", &stdout, &stderr)
			}
		})
	}
	assertFileAbsent(t, missingParent)
}

func TestCLIPreservesOutputOnSourceError(t *testing.T) {
	directory := t.TempDir()
	invalidPath := filepath.Join(directory, "invalid.json")
	if err := os.WriteFile(invalidPath, []byte(`[{"command":"sub"},{"command":"unknown"}]`), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, inputPath := range []string{invalidPath, filepath.Join(directory, "missing.json")} {
		t.Run(filepath.Base(inputPath), func(t *testing.T) {
			outputPath := filepath.Join(directory, "existing.bin")
			original := []byte("previous binary")
			if err := os.WriteFile(outputPath, original, 0o600); err != nil {
				t.Fatal(err)
			}
			args := []string{"-input", inputPath, "-output", outputPath, "-test"}
			var stdout, stderr bytes.Buffer
			if code := cli.Run(args, &stdout, &stderr); code != failureExitCode {
				t.Fatalf("expected exit code 1, got %d", code)
			}
			if stdout.Len() != 0 || stderr.Len() == 0 {
				t.Fatalf("invalid diagnostics: stdout=%q stderr=%q", &stdout, &stderr)
			}
			assertFileBytes(t, outputPath, original)
		})
	}
}
