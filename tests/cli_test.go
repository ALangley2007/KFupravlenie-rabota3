package tests

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"kf3/src/assembler"
	"kf3/src/cli"
)

const (
	successExitCode = 0
	failureExitCode = 1
	usageExitCode   = 2
)

func TestCLIPrintsRepresentation(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "program.bin")
	args := []string{"-input", "fixtures/load_const.json", "-output", outputPath, "-test"}
	var stdout, stderr bytes.Buffer
	if code := cli.Run(args, &stdout, &stderr); code != successExitCode {
		t.Fatalf("exit code %d: %s", code, &stderr)
	}
	var got assembler.Program
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	want := assembler.Program{{"A": 128, "B": 46}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	if stderr.Len() != 0 {
		t.Fatalf("unexpected stderr: %s", &stderr)
	}
	assertFileAbsent(t, outputPath)
}

func TestCLIValidatesWithoutTestMode(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "missing", "program.bin")
	args := []string{"-input", "fixtures/sub.json", "-output", outputPath}
	var stdout, stderr bytes.Buffer
	if code := cli.Run(args, &stdout, &stderr); code != successExitCode {
		t.Fatalf("exit code %d: %s", code, &stderr)
	}
	if stdout.Len() != 0 || stderr.Len() != 0 {
		t.Fatalf("unexpected output: stdout=%q stderr=%q", &stdout, &stderr)
	}
	assertFileAbsent(t, filepath.Dir(outputPath))
}

func TestCLIPreservesExistingOutput(t *testing.T) {
	for _, mode := range []string{"-test=true", "-test=false"} {
		t.Run(mode, func(t *testing.T) {
			outputPath := filepath.Join(t.TempDir(), "program.bin")
			original := []byte("existing output must survive stage one")
			if err := os.WriteFile(outputPath, original, 0o600); err != nil {
				t.Fatal(err)
			}
			args := []string{"-input", "fixtures/sub.json", "-output", outputPath, mode}
			var stderr bytes.Buffer
			if code := cli.Run(args, io.Discard, &stderr); code != successExitCode {
				t.Fatalf("exit code %d: %s", code, &stderr)
			}
			got, err := os.ReadFile(outputPath)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, original) {
				t.Fatalf("output was modified: %q", got)
			}
		})
	}
}

func TestCLIUsageErrors(t *testing.T) {
	cases := map[string][]string{
		"no flags":       {},
		"missing input":  {"-output", "program.bin"},
		"missing output": {"-input", "fixtures/sub.json"},
		"unknown flag":   {"-unknown"},
		"missing value":  {"-input"},
		"invalid bool":   {"-test=maybe"},
		"positional":     {"-input", "fixtures/sub.json", "-output", "program.bin", "extra"},
	}
	for name, args := range cases {
		t.Run(name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			if code := cli.Run(args, &stdout, &stderr); code != usageExitCode {
				t.Fatalf("expected exit code 2, got %d", code)
			}
			if stdout.Len() != 0 || stderr.Len() == 0 {
				t.Fatalf("invalid diagnostics: stdout=%q stderr=%q", &stdout, &stderr)
			}
		})
	}
}

func TestCLISourceErrors(t *testing.T) {
	directory := t.TempDir()
	invalidPath := filepath.Join(directory, "invalid.json")
	if err := os.WriteFile(invalidPath, []byte(`[{"command":"unknown"}]`), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, inputPath := range []string{filepath.Join(directory, "missing.json"), invalidPath} {
		t.Run(filepath.Base(inputPath), func(t *testing.T) {
			outputPath := filepath.Join(directory, "program.bin")
			args := []string{"-input", inputPath, "-output", outputPath, "-test"}
			var stdout, stderr bytes.Buffer
			if code := cli.Run(args, &stdout, &stderr); code != failureExitCode {
				t.Fatalf("expected exit code 1, got %d", code)
			}
			if stdout.Len() != 0 || !strings.Contains(stderr.String(), "error:") {
				t.Fatalf("invalid diagnostics: stdout=%q stderr=%q", &stdout, &stderr)
			}
			assertFileAbsent(t, outputPath)
		})
	}
}

func TestCLIHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := cli.Run([]string{"-h"}, &stdout, &stderr); code != successExitCode {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	for _, flagName := range []string{"-input", "-output", "-test"} {
		if !strings.Contains(stderr.String(), flagName) {
			t.Errorf("help does not describe %s", flagName)
		}
	}
}

func TestCLIOutputError(t *testing.T) {
	args := []string{"-input", "fixtures/sub.json", "-output", "unused.bin", "-test"}
	var stderr bytes.Buffer
	if code := cli.Run(args, failingWriter{}, &stderr); code != failureExitCode {
		t.Fatalf("expected exit code 1, got %d", code)
	}
	if !strings.Contains(stderr.String(), "print intermediate representation") {
		t.Fatalf("missing output error: %s", &stderr)
	}
}

type failingWriter struct{}

func (failingWriter) Write(data []byte) (int, error) {
	return 0, errors.New("writer is unavailable")
}

func assertFileAbsent(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("stage one created an output artifact at %s: %v", path, err)
	}
}
