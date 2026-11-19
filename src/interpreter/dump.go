package interpreter

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
)

// Dump writes a CSV header and every data cell in the inclusive address range [start, end].
// Addresses and unsigned values are written as decimal numbers in ascending address order.
func (machine *Machine) Dump(writer io.Writer, start, end int) error {
	if start < 0 || end < start || end >= MemorySize {
		return fmt.Errorf("dump range must satisfy 0 <= start <= end < %d", MemorySize)
	}
	output := csv.NewWriter(writer)
	if err := output.Write([]string{"address", "value"}); err != nil {
		return fmt.Errorf("write CSV header: %w", err)
	}
	for address := start; address <= end; address++ {
		row := []string{strconv.Itoa(address), strconv.FormatUint(uint64(machine.data[address]), 10)}
		if err := output.Write(row); err != nil {
			return fmt.Errorf("write CSV row: %w", err)
		}
	}
	output.Flush()
	if err := output.Error(); err != nil {
		return fmt.Errorf("flush CSV: %w", err)
	}
	return nil
}
