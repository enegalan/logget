package flags

import (
	"logget/src/io"
	"strings"
)

type HeaderArray struct {
	SimpleFlag[[]string]
}

func (h *HeaderArray) Type() string { return "<header|file>" }

func (h *HeaderArray) Set(value string) error {
	lines, _, err := io.TryReadAsFile(value, ":", func(val string) bool { return strings.Contains(val, ":") || val != "" },
		func(line string) bool { return strings.Contains(line, ":") }, func(line string) string { return line })
	if err != nil {
		return err
	}
	if h.Value == nil {
		h.Value = []string{}
	}
	h.Value = append(h.Value, lines...)
	return nil
}

func (h *HeaderArray) Get() []string {
	if h.Value == nil {
		return []string{}
	}
	return h.Value
}
