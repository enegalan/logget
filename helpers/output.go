package helpers

import (
	"fmt"
	"os"
	"path/filepath"
)

func WriteOutput(cfg Config, content string) error {
	if cfg.OutputFile != "" {
		var filePath string
		if cfg.OutputDir != "" {
			if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
				return fmt.Errorf("failed to create output directory: %v", err)
			}
			filePath = filepath.Join(cfg.OutputDir, cfg.OutputFile)
		} else {
			filePath = cfg.OutputFile
		}
		var file *os.File
		var err error
		if cfg.AppendMode {
			file, err = os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		} else {
			file, err = os.Create(filePath)
		}
		if err != nil {
			return fmt.Errorf("failed to open output file: %v", err)
		}
		defer file.Close()
		if _, err = file.WriteString(content); err != nil {
			return fmt.Errorf("failed to write to output file: %v", err)
		}
		return nil
	}
	fmt.Print(content)
	return nil
}
