package tests

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"kf3/src/cli"
)

func TestInterpreterCLICopiesArray(t *testing.T) {
	directory := t.TempDir()
	binaryPath := filepath.Join(directory, "copy array.bin")
	dumpPath := filepath.Join(directory, "memory dump.csv")
	args := []string{"-input", "fixtures/copy_array.json", "-output", binaryPath}
	var stderr bytes.Buffer
	if code := cli.Run(args, io.Discard, &stderr); code != successExitCode {
		t.Fatalf("assembler exit code %d: %s", code, &stderr)
	}
	if err := os.WriteFile(dumpPath, bytes.Repeat([]byte("old content"), 100), 0o600); err != nil {
		t.Fatal(err)
	}
	args = []string{"-input", binaryPath, "-output", dumpPath, "-start", "0", "-end", "12"}
	if code := cli.RunInterpreter(args, &stderr); code != successExitCode || stderr.Len() != 0 {
		t.Fatalf("interpreter exit code %d: %s", code, &stderr)
	}
	want, err := os.ReadFile("fixtures/copy_array.csv")
	if err != nil {
		t.Fatal(err)
	}
	assertFileBytes(t, dumpPath, want)
}

func TestInterpreterCLIUsageErrors(t *testing.T) {
	valid := []string{"-input", "program.bin", "-output", "memory.csv", "-start", "0", "-end", "1"}
	cases := [][]string{
		{}, {"-huh"}, {"-input"},
		{"-output", "memory.csv", "-start", "0", "-end", "1"},
		{"-input", "program.bin", "-start", "0", "-end", "1"},
		{"-input", "program.bin", "-output", "memory.csv"},
		append(append([]string{}, valid...), "extra"),
	}
	for _, bounds := range [][2]string{{"-1", "0"}, {"2", "1"}, {"0", "131072"}, {"x", "1"}} {
		args := append([]string{}, valid[:4]...)
		cases = append(cases, append(args, "-start", bounds[0], "-end", bounds[1]))
	}
	for _, flag := range []string{"-start", "-end"} {
		cases = append(cases, append(append([]string{}, valid[:4]...), flag, "0"))
	}
	for _, args := range cases {
		var stderr bytes.Buffer
		if code := cli.RunInterpreter(args, &stderr); code != usageExitCode || stderr.Len() == 0 {
			t.Fatalf("args %v: exit code %d, stderr=%q", args, code, &stderr)
		}
	}
}

func TestInterpreterCLIPreservesDumpOnExecutionError(t *testing.T) {
	directory := t.TempDir()
	cases := map[string][]byte{
		"truncated": {0x80},
		"unknown":   {0x00, 0x00, 0x00, 0x00},
		"underflow": {0x53, 0x00, 0x00, 0x00},
		"sub":       {0xA5, 0x00, 0x00, 0x00},
	}
	for name, code := range cases {
		t.Run(name, func(t *testing.T) {
			binaryPath := filepath.Join(directory, name+".bin")
			if err := os.WriteFile(binaryPath, code, 0o600); err != nil {
				t.Fatal(err)
			}
			assertInterpreterInputFailure(t, binaryPath)
		})
	}
	assertInterpreterInputFailure(t, filepath.Join(directory, "missing.bin"))
	assertInterpreterInputFailure(t, directory)
}

func TestInterpreterCLIDumpErrors(t *testing.T) {
	directory := t.TempDir()
	inputPath := filepath.Join(directory, "empty.bin")
	if err := os.WriteFile(inputPath, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	missingParent := filepath.Join(directory, "missing")
	for _, path := range []string{directory, filepath.Join(missingParent, "memory.csv")} {
		args := []string{"-input", inputPath, "-output", path, "-start", "0", "-end", "0"}
		var stderr bytes.Buffer
		code := cli.RunInterpreter(args, &stderr)
		if code != failureExitCode || !strings.Contains(stderr.String(), "create memory dump:") {
			t.Fatalf("exit code %d, stderr=%q", code, &stderr)
		}
	}
	assertFileAbsent(t, missingParent)
}

func TestInterpreterCLIEmptyProgram(t *testing.T) {
	directory := t.TempDir()
	inputPath := filepath.Join(directory, "empty.bin")
	dumpPath := filepath.Join(directory, "empty.csv")
	if err := os.WriteFile(inputPath, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	args := []string{"-input", inputPath, "-output", dumpPath, "-start", "131071", "-end", "131071"}
	var stderr bytes.Buffer
	if code := cli.RunInterpreter(args, &stderr); code != successExitCode || stderr.Len() != 0 {
		t.Fatalf("exit code %d, stderr=%q", code, &stderr)
	}
	assertFileBytes(t, dumpPath, []byte("address,value\n131071,0\n"))
}

func TestInterpreterCLIHelp(t *testing.T) {
	var stderr bytes.Buffer
	if code := cli.RunInterpreter([]string{"-h"}, &stderr); code != successExitCode {
		t.Fatalf("expected help exit code 0, got %d", code)
	}
	for _, flagName := range []string{"-input", "-output", "-start", "-end"} {
		if !strings.Contains(stderr.String(), flagName) {
			t.Errorf("help does not describe %s", flagName)
		}
	}
}

func assertInterpreterInputFailure(t *testing.T, inputPath string) {
	t.Helper()
	for _, existing := range []bool{false, true} {
		dumpPath := filepath.Join(t.TempDir(), "memory.csv")
		original := []byte("previous dump")
		if existing {
			if err := os.WriteFile(dumpPath, original, 0o600); err != nil {
				t.Fatal(err)
			}
		}
		args := []string{"-input", inputPath, "-output", dumpPath, "-start", "0", "-end", "1"}
		var stderr bytes.Buffer
		if code := cli.RunInterpreter(args, &stderr); code != failureExitCode || stderr.Len() == 0 {
			t.Fatalf("exit code %d, stderr=%q", code, &stderr)
		}
		if existing {
			assertFileBytes(t, dumpPath, original)
		} else {
			assertFileAbsent(t, dumpPath)
		}
	}
}
