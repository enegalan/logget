package flags

import (
	"logget/helpers"
	"strings"
)

type CookieArray []string

func (c *CookieArray) String() string {
	return strings.Join(*c, ", ")
}

func (c *CookieArray) Set(value string) error {
	processCookieLine := func(line string) string {
		if strings.Contains(line, "\t") {
			fields := strings.Split(line, "\t")
			if len(fields) >= 7 {
				name := strings.TrimSpace(fields[5])
				value := strings.TrimSpace(fields[6])
				if name != "" {
					return name + "=" + value
				}
			}
			return ""
		}
		if strings.Contains(line, "=") {
			return line
		}
		return ""
	}
	lines, _, err := helpers.TryReadAsFile(
		value,
		"=",
		func(val string) bool { return strings.Contains(val, "=") || val != "" },
		func(line string) bool { return strings.Contains(line, "\t") || strings.Contains(line, "=") },
		processCookieLine,
	)
	if err != nil {
		return err
	}
	*c = append(*c, lines...)
	return nil
}

func (c *CookieArray) Type() string {
	return "<data|filename>"
}
