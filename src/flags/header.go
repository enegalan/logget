package flags

import (
	helpers "logget/src"
	"strings"
)

type HeaderArray []string

func (h *HeaderArray) String() string {
	return strings.Join(*h, ", ")
}

func (h *HeaderArray) Set(value string) error {
	lines, _, err := helpers.TryReadAsFile(
		value,
		":",
		func(val string) bool { return strings.Contains(val, ":") || val != "" },
		func(line string) bool { return strings.Contains(line, ":") },
		func(line string) string { return line },
	)
	if err != nil {
		return err
	}
	*h = append(*h, lines...)
	return nil
}

func (h *HeaderArray) Type() string {
	return "<header|file>"
}
