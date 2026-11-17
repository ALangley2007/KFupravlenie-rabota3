package assembler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

// Parse validates a JSON source and returns its instructions in their original order.
func Parse(reader io.Reader) (Program, error) {
	source, err := decodeSource(reader)
	if err != nil {
		return nil, err
	}
	program := make(Program, 0, len(source))
	for index, instruction := range source {
		parsed, err := parseInstruction(instruction)
		if err != nil {
			return nil, fmt.Errorf("instruction %d: %w", index+1, err)
		}
		program = append(program, parsed)
	}
	return program, nil
}

func decodeSource(reader io.Reader) ([]map[string]json.RawMessage, error) {
	decoder := json.NewDecoder(reader)
	var source []map[string]json.RawMessage
	if err := decoder.Decode(&source); err != nil {
		return nil, fmt.Errorf("invalid JSON source: %w", err)
	}
	if source == nil {
		return nil, errors.New("source must be a JSON array")
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return nil, errors.New("source must contain a single JSON document")
	}
	return source, nil
}

func parseInstruction(source map[string]json.RawMessage) (map[string]uint32, error) {
	var command string
	if err := json.Unmarshal(source[commandField], &command); err != nil {
		return nil, errors.New(`"command" must be a string`)
	}
	spec, ok := commands[command]
	if !ok {
		return nil, fmt.Errorf("unknown command %q", command)
	}
	if err := validateFields(source, spec); err != nil {
		return nil, err
	}
	instruction := map[string]uint32{"A": spec.opcode}
	if spec.operand == "" {
		return instruction, nil
	}
	operand, err := parseOperand(source[spec.operand], spec)
	if err != nil {
		return nil, err
	}
	instruction["B"] = operand
	return instruction, nil
}

func validateFields(source map[string]json.RawMessage, spec commandSpec) error {
	for name := range source {
		if name != commandField && (spec.operand == "" || name != spec.operand) {
			return fmt.Errorf("unexpected field %q", name)
		}
	}
	return nil
}

func parseOperand(raw json.RawMessage, spec commandSpec) (uint32, error) {
	var value *uint32
	if err := json.Unmarshal(raw, &value); err != nil || value == nil {
		return 0, fmt.Errorf("%q must be an unsigned integer", spec.operand)
	}
	if *value > spec.maxOperand {
		return 0, fmt.Errorf("%q must be in range 0..%d", spec.operand, spec.maxOperand)
	}
	return *value, nil
}
