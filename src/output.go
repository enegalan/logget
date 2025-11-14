package helpers

import (
	"fmt"
	"os"
	"sync"
)

var fileWriteState = make(map[string]bool)

type OutputWriter struct {
	file   *os.File
	mu     sync.Mutex
	closed bool
}

func NewOutputWriter(filePath string, appendMode bool) (*OutputWriter, error) {
	flags := os.O_CREATE | os.O_WRONLY
	if appendMode {
		flags |= os.O_APPEND
	} else {
		flags |= os.O_TRUNC
	}
	file, err := os.OpenFile(filePath, flags, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open output file: %v", err)
	}
	return &OutputWriter{file: file}, nil
}

func (w *OutputWriter) Write(content string) error {
	if w == nil || w.closed {
		return fmt.Errorf("output writer is closed")
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.closed {
		return fmt.Errorf("output writer is closed")
	}
	_, err := w.file.WriteString(content)
	return err
}

func (w *OutputWriter) Close() error {
	if w == nil || w.closed {
		return nil
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.closed {
		return nil
	}
	w.closed = true
	return w.file.Close()
}

func getFileOpenFlags(cfg Config, filePath string) (int, error) {
	if cfg.FollowMode {
		if !fileWriteState[filePath] && !cfg.AppendMode {
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
	if cfg.OutputWriter != nil {
		return cfg.OutputWriter.Write(content)
	}
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
