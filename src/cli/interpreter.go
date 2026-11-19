package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"kf3/src/interpreter"
)

type interpreterConfig struct {
	inputPath  string
	outputPath string
	start      int
	end        int
}

// RunInterpreter executes a binary file and writes the specified range of data memory to CSV.
// Help and errors are written to stderr; the exit codes match Run.
func RunInterpreter(args []string, stderr io.Writer) int {
	cfg, err := parseInterpreterFlags(args, stderr)
	if errors.Is(err, flag.ErrHelp) {
		return exitSuccess
	}
	if err != nil {
		return reportError(stderr, err, exitUsage)
	}
	machine, err := executeBinary(cfg.inputPath)
	if err != nil {
		return reportError(stderr, err, exitFailure)
	}
	if err := writeDump(cfg, machine); err != nil {
		return reportError(stderr, err, exitFailure)
	}
	return exitSuccess
}

func parseInterpreterFlags(args []string, stderr io.Writer) (interpreterConfig, error) {
	var cfg interpreterConfig
	flags := flag.NewFlagSet("interpreter", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.StringVar(&cfg.inputPath, "input", "", "path to the input binary (required)")
	flags.StringVar(&cfg.outputPath, "output", "", "path to the CSV memory dump (required)")
	flags.IntVar(&cfg.start, "start", -1, "first data address, inclusive (required)")
	flags.IntVar(&cfg.end, "end", -1, "last data address, inclusive (required)")
	if err := flags.Parse(args); err != nil {
		return cfg, err
	}
	if flags.NArg() != 0 {
		return cfg, errors.New("positional arguments are not supported; use -input, -output, -start and -end")
	}
	if cfg.inputPath == "" || cfg.outputPath == "" {
		return cfg, errors.New("both -input and -output are required")
	}
	if cfg.start < 0 || cfg.end < cfg.start || cfg.end >= interpreter.MemorySize {
		return cfg, fmt.Errorf("required range: 0 <= -start <= -end < %d", interpreter.MemorySize)
	}
	return cfg, nil
}

func executeBinary(path string) (*interpreter.Machine, error) {
	code, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read binary: %w", err)
	}
	machine, err := interpreter.New(code)
	if err != nil {
		return nil, err
	}
	if err := machine.Run(); err != nil {
		return nil, err
	}
	return machine, nil
}

func writeDump(cfg interpreterConfig, machine *interpreter.Machine) error {
	file, err := os.Create(cfg.outputPath)
	if err != nil {
		return fmt.Errorf("create memory dump: %w", err)
	}
	dumpErr := machine.Dump(file, cfg.start, cfg.end)
	closeErr := file.Close()
	if dumpErr != nil {
		return dumpErr
	}
	if closeErr != nil {
		return fmt.Errorf("close memory dump: %w", closeErr)
	}
	return nil
}
