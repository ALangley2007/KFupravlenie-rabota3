// Package cli provides the stage-one assembler command-line interface.
package cli

import (
	"encoding/json"
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

// Run parses arguments, validates the source, and optionally prints its intermediate representation.
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
	if cfg.testMode {
		if err := printProgram(stdout, program); err != nil {
			return reportError(stderr, err, exitFailure)
		}
	}
	return exitSuccess
}

func parseFlags(args []string, stderr io.Writer) (config, error) {
	var cfg config
	flags := flag.NewFlagSet("assembler", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.StringVar(&cfg.inputPath, "input", "", "path to the JSON source (required)")
	flags.StringVar(&cfg.outputPath, "output", "", "binary path reserved for stage 2 (required; no file is written)")
	flags.BoolVar(&cfg.testMode, "test", false, "print the intermediate representation as JSON")
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

func printProgram(writer io.Writer, program assembler.Program) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(program); err != nil {
		return fmt.Errorf("print intermediate representation: %w", err)
	}
	return nil
}

func reportError(stderr io.Writer, err error, code int) int {
	fmt.Fprintln(stderr, "error:", err)
	return code
}
