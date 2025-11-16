package flags

import (
	helpers "logget/src"
	"strings"
)

type CookieArray struct {
	SimpleFlag[[]string]
}

func (c *CookieArray) Type() string { return "<data|file>" }

func (c *CookieArray) Set(value string) error {
	processCookieLine := func(line string) string {
		if strings.Contains(line, "\t") {
			fields := strings.Split(line, "\t")
			if len(fields) >= 7 {
				name := strings.TrimSpace(fields[5])
				val := strings.TrimSpace(fields[6])
				if name != "" {
					return name + "=" + val
				}
			}
			return ""
		}
		if strings.Contains(line, "=") {
			return line
		}
		return ""
	}
	lines, _, err := helpers.TryReadAsFile(value, "=",
		func(val string) bool { return strings.Contains(val, "=") || val != "" },
		func(line string) bool { return strings.Contains(line, "\t") || strings.Contains(line, "=") },
		processCookieLine)
	if err != nil {
		return err
	}
	if c.Value == nil {
		c.Value = []string{}
	}
	c.Value = append(c.Value, lines...)
	return nil
}

func (c *CookieArray) Get() []string {
	if c.Value == nil {
		return []string{}
	}
	return c.Value
}
