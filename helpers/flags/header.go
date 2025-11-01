package flags

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type HeaderArray []string

func (h *HeaderArray) String() string {
	return strings.Join(*h, ", ")
}

func (h *HeaderArray) Set(value string) error {
	// Check if the value contains ':' - if so, treat it as direct header data
	if strings.Contains(value, ":") {
		*h = append(*h, value)
		return nil
	}
	// Otherwise, try to read it as a filename
	fileInfo, err := os.Stat(value)
	if err != nil {
		// If file doesn't exist, treat as header data even without ':'
		*h = append(*h, value)
		return nil
	}
	// If it's not a regular file, treat as header data
	if fileInfo.IsDir() {
		*h = append(*h, value)
		return nil
	}
	// Read the file and parse headers
	file, err := os.Open(value)
	if err != nil {
		return fmt.Errorf("failed to open header file %s: %v", value, err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		// Skip empty lines and comments (lines starting with #)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Header format: Key: Value
		if strings.Contains(line, ":") {
			*h = append(*h, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading header file %s at line %d: %v", value, lineNum, err)
	}
	return nil
}

func (h *HeaderArray) Type() string {
	return "<header|file>"
}
