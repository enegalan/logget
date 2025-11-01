package flags

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type CookieArray []string

func (c *CookieArray) String() string {
	return strings.Join(*c, ", ")
}

func (c *CookieArray) Set(value string) error {
	// Check if the value contains '=' - if so, treat it as direct cookie data
	if strings.Contains(value, "=") {
		*c = append(*c, value)
		return nil
	}
	// Otherwise, try to read it as a filename
	fileInfo, err := os.Stat(value)
	if err != nil {
		// If file doesn't exist, treat as cookie data even without '='
		*c = append(*c, value)
		return nil
	}
	// If it's not a regular file, treat as cookie data
	if fileInfo.IsDir() {
		*c = append(*c, value)
		return nil
	}
	// Read the file and parse cookies
	file, err := os.Open(value)
	if err != nil {
		return fmt.Errorf("failed to open cookie file %s: %v", value, err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Parse Netscape cookie format or simple format
		// Netscape format: domain	flag	path	secure	expiration	name	value
		// Simple format: name=value or name=value; domain=example.com
		if strings.Contains(line, "\t") {
			// Netscape cookie format
			fields := strings.Split(line, "\t")
			if len(fields) >= 7 {
				// Extract name and value (fields are: domain, flag, path, secure, expiration, name, value)
				name := strings.TrimSpace(fields[5])
				value := strings.TrimSpace(fields[6])
				if name != "" {
					// Build cookie string: name=value
					*c = append(*c, name+"="+value)
				}
			}
		} else if strings.Contains(line, "=") {
			// Simple format: name=value or name=value; domain=example.com
			*c = append(*c, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading cookie file %s at line %d: %v", value, lineNum, err)
	}
	return nil
}

func (c *CookieArray) Type() string {
	return "<data|filename>"
}
