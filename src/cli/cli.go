// Package cli provides the assembler and interpreter command-line interfaces.
package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"kf3/src/assembler"
)

const (
	exitSuccess = 0
	exitFailure = 1
	exitUsage   = 2
)

type config struct {
	inputPath  string
	outputPath string
	testMode   bool
}

// Run assembles a source file, writes machine code, and reports its size and optional test output.
func Run(args []string, stdout, stderr io.Writer) int {
	cfg, err := parseFlags(args, stderr)
	if errors.Is(err, flag.ErrHelp) {
		return exitSuccess
	}
	if err != nil {
		return reportError(stderr, err, exitUsage)
	}
	program, err := readProgram(cfg.inputPath)
	if err != nil {
		return reportError(stderr, err, exitFailure)
	}
	code := assembler.Encode(program)
	if err := writeBinary(cfg.outputPath, code); err != nil {
		return reportError(stderr, err, exitFailure)
	}
	if err := printResult(stdout, program, code, cfg.testMode); err != nil {
		return reportError(stderr, err, exitFailure)
	}
	return exitSuccess
}

func parseFlags(args []string, stderr io.Writer) (config, error) {
	var cfg config
	flags := flag.NewFlagSet("assembler", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.StringVar(&cfg.inputPath, "input", "", "path to the JSON source (required)")
	flags.StringVar(&cfg.outputPath, "output", "", "path to the output binary (required)")
	flags.BoolVar(&cfg.testMode, "test", false, "print the intermediate representation and machine-code bytes")
	if err := flags.Parse(args); err != nil {
		return cfg, err
	}
	if flags.NArg() != 0 {
		return cfg, errors.New("positional arguments are not supported; use -input, -output and -test")
	}
	if cfg.inputPath == "" || cfg.outputPath == "" {
		return cfg, errors.New("both -input and -output are required")
	}
	return cfg, nil
}

func readProgram(path string) (assembler.Program, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open source: %w", err)
	}
	defer file.Close()
	return assembler.Parse(file)
}

func reportError(stderr io.Writer, err error, code int) int {
	fmt.Fprintln(stderr, "error:", err)
	return code
}
