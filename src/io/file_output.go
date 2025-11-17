package io

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

type WriteConfig struct {
	OutputWriter OutputWriterInterface
	OutputFile   string
	AppendMode   bool
	FollowMode   bool
}

type OutputWriterInterface interface {
	Write(content string) error
	Close() error
}

func getFileOpenFlags(followMode, appendMode bool, filePath string) (int, error) {
	if followMode {
		if !fileWriteState[filePath] && !appendMode {
			fileWriteState[filePath] = true
			return os.O_CREATE | os.O_WRONLY | os.O_TRUNC, nil
		}
		return os.O_CREATE | os.O_WRONLY | os.O_APPEND, nil
	}
	if appendMode {
		return os.O_CREATE | os.O_WRONLY | os.O_APPEND, nil
	}
	if _, err := os.Stat(filePath); err == nil {
		return os.O_WRONLY | os.O_TRUNC, nil
	}
	return 0, os.ErrNotExist
}

func WriteOutput(cfg WriteConfig, content string) error {
	if cfg.OutputWriter != nil {
		return cfg.OutputWriter.Write(content)
	}
	if cfg.OutputFile == "" {
		fmt.Print(content)
		return nil
	}
	filePath := cfg.OutputFile
	flags, flagErr := getFileOpenFlags(cfg.FollowMode, cfg.AppendMode, filePath)
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

type LoggerInterface interface {
	Success(format string, args ...interface{})
}

func LogOutputFileSuccess(outputFile string, appendMode bool, outputType string, logger LoggerInterface) {
	if logger == nil {
		return
	}
	if appendMode {
		logger.Success("%s appended to: %s", outputType, outputFile)
	} else {
		logger.Success("%s written to: %s", outputType, outputFile)
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

func LogHARFileSuccess(filePath string, logger LoggerInterface) {
	if logger != nil {
		logger.Success("HAR file written to: %s", filePath)
	}
}
