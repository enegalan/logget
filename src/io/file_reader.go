package io

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func ReadLinesFromFile(filePath string, validateLine func(string) bool) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %v", filePath, err)
	}
	defer file.Close()
	var lines []string
	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if validateLine(line) {
			lines = append(lines, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file %s at line %d: %v", filePath, lineNum, err)
	}
	return lines, nil
}

func TryReadAsFile(value string, delimiter string, validateDirectValue func(string) bool, validateLine func(string) bool, processLine func(string) string) ([]string, bool, error) {
	if strings.Contains(value, delimiter) && validateDirectValue(value) {
		return []string{value}, false, nil
	}
	fileInfo, err := os.Stat(value)
	if err != nil {
		if validateDirectValue(value) {
			return []string{value}, false, nil
		}
		return nil, false, fmt.Errorf("file does not exist and value is not valid: %s", value)
	}
	if fileInfo.IsDir() {
		if validateDirectValue(value) {
			return []string{value}, false, nil
		}
		return nil, false, fmt.Errorf("path is a directory: %s", value)
	}
	lines, err := ReadLinesFromFile(value, validateLine)
	if err != nil {
		return nil, false, err
	}
	var result []string
	for _, line := range lines {
		processed := processLine(line)
		if processed != "" {
			result = append(result, processed)
		}
	}
	return result, true, nil
}
