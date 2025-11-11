package helpers

import (
	"fmt"
	"os"
)

var fileWriteState = make(map[string]bool)

func getFileOpenFlags(cfg Config, filePath string) (int, error) {
	if cfg.FollowMode {
		firstWrite := !fileWriteState[filePath]
		if firstWrite && !cfg.AppendMode {
			fileWriteState[filePath] = true
			return os.O_CREATE | os.O_WRONLY | os.O_TRUNC, nil
		}
		return os.O_CREATE | os.O_WRONLY | os.O_APPEND, nil
	}
	if cfg.AppendMode {
		return os.O_CREATE | os.O_WRONLY | os.O_APPEND, nil
	}
	if _, err := os.Stat(filePath); err == nil {
		return os.O_WRONLY | os.O_TRUNC, nil
	}
	return 0, os.ErrNotExist
}

func WriteOutput(cfg Config, content string) error {
	if cfg.OutputFile == "" {
		fmt.Print(content)
		return nil
	}
	filePath := cfg.OutputFile
	flags, flagErr := getFileOpenFlags(cfg, filePath)
	if flagErr == os.ErrNotExist {
		file, err := os.Create(filePath)
		if err != nil {
			return fmt.Errorf("failed to create output file: %v", err)
		}
		defer file.Close()
		if _, err := file.WriteString(content); err != nil {
			return fmt.Errorf("failed to write to output file: %v", err)
		}
		return nil
	}
	if flagErr != nil {
		return fmt.Errorf("failed to determine file flags: %v", flagErr)
	}
	file, err := os.OpenFile(filePath, flags, 0644)
	if err != nil {
		return fmt.Errorf("failed to open output file: %v", err)
	}
	defer file.Close()
	if _, err := file.WriteString(content); err != nil {
		return fmt.Errorf("failed to write to output file: %v", err)
	}
	return nil
}

func LogOutputFileSuccess(cfg Config, outputType string, logger *Logger) {
	filePath := cfg.OutputFile
	if cfg.AppendMode {
		logger.Success("%s appended to: %s", outputType, filePath)
	} else {
		logger.Success("%s written to: %s", outputType, filePath)
	}
}

func WriteHAROutput(filePath string, harData []byte) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create HAR file: %v", err)
	}
	defer file.Close()
	if _, err = file.Write(harData); err != nil {
		return fmt.Errorf("failed to write HAR file: %v", err)
	}
	return nil
}

func LogHARFileSuccess(filePath string, logger *Logger) {
	logger.Success("HAR file written to: %s", filePath)
}
