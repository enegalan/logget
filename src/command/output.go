package command

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	chrome "logget/src/chrome"
	"logget/src/core"
	"logget/src/io"

	"gopkg.in/yaml.v3"
)

const maxOutputErrorsToShow = 5

func ValidateOutputFormats(cfg Config) {
	formatCount := 0
	for _, f := range []bool{cfg.JSONOutput, cfg.YAMLOutput, cfg.CSVOutput, cfg.HAROutput} {
		if f {
			formatCount++
		}
	}
	if formatCount > 1 {
		fmt.Println("logget: Only one output format can be specified at a time")
		os.Exit(1)
	}
}

type OutputData struct {
	URL      string                `json:"url"`
	Logs     []chrome.LogEntry     `json:"logs,omitempty"`
	Network  []chrome.NetworkEntry `json:"network,omitempty"`
	Duration time.Duration         `json:"duration"`
}

type outputErrorTracker struct {
	count      int
	maxErrors  int
	suppressed bool
}

func NewOutputErrorTracker() *outputErrorTracker {
	return &outputErrorTracker{maxErrors: maxOutputErrorsToShow}
}

func (t *outputErrorTracker) handleError(err error, logger *core.Logger, message string) {
	if err == nil {
		return
	}
	t.count++
	if t.count <= t.maxErrors {
		logger.Warn(message, err)
	}
	if t.count == t.maxErrors && !t.suppressed {
		logger.Warn("Too many output errors, suppressing further warnings...")
		t.suppressed = true
	}
}

func getRequestHeadersList(cfg Config, url string, requestCaptured bool, requestHeaders map[string]string) []string {
	if requestCaptured && len(requestHeaders) > 0 {
		var result []string
		for name, value := range requestHeaders {
			result = append(result, fmt.Sprintf("%s: %s", name, value))
		}
		return result
	}
	return core.GenerateHeaders(cfg.UserAgent, cfg.Headers, url)
}

func writeOutputAndLog(cfg Config, content string, outputType string, logger *core.Logger) {
	if err := io.WriteOutput(io.WriteConfig{
		OutputWriter: cfg.OutputWriter,
		OutputFile:   cfg.OutputFile,
		AppendMode:   cfg.AppendMode,
		FollowMode:   cfg.FollowMode,
	}, content); err != nil {
		logger.Fatal("Failed to write %s: %v", outputType, err)
	}
	if cfg.OutputFile != "" {
		io.LogOutputFileSuccess(cfg.OutputFile, cfg.AppendMode, outputType, logger)
	}
}

func marshalAndWrite(cfg Config, data []byte, outputType string, logger *core.Logger) {
	writeOutputAndLog(cfg, string(data)+"\n", outputType, logger)
}

func WriteFinalOutput(cfg Config, output OutputData, network []chrome.NetworkEntry, url string, startTime time.Time, responseProtocol string, statusCode int, duration time.Duration, logger *core.Logger, formatter *OutputFormatter, jsonOutput bool, yamlOutput bool, verbose bool, showLogs bool, showNetwork bool, requestCaptured bool, requestHeaders map[string]string, responseHeaders map[string]string) {
	if cfg.HAROutput {
		harData, err := core.ConvertNetworkEntriesToHAR(network, url, startTime)
		if err != nil {
			logger.Fatal("Failed to generate HAR: %v", err)
		}
		marshalAndWrite(cfg, harData, "HAR output", logger)
		return
	}
	if jsonOutput {
		jsonData, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			logger.Fatal("Failed to marshal JSON: %v", err)
		}
		marshalAndWrite(cfg, jsonData, "JSON output", logger)
		return
	}
	if yamlOutput {
		yamlData, err := yaml.Marshal(output)
		if err != nil {
			logger.Fatal("Failed to marshal YAML: %v", err)
		}
		marshalAndWrite(cfg, yamlData, "YAML output", logger)
		return
	}
	var outputContent strings.Builder
	if verbose && !jsonOutput && !yamlOutput {
		outputContent.WriteString(formatter.FormatHTTPResponse(responseProtocol, statusCode, duration))
		requestHeadersList := getRequestHeadersList(cfg, url, requestCaptured, requestHeaders)
		outputContent.WriteString(formatter.FormatRequestHeaders(requestHeadersList))
		outputContent.WriteString(formatter.FormatResponseHeaders(responseHeaders))
	}
	if showLogs && len(output.Logs) > 0 {
		outputContent.WriteString(formatter.FormatConsoleLogs(output.Logs))
	}
	if showNetwork && len(output.Network) > 0 {
		outputContent.WriteString(formatter.FormatNetworkRequests(output.Network))
	}
	writeOutputAndLog(cfg, outputContent.String(), "Output", logger)
}

func PrepareOutputFile(cfg Config, url string, logger *core.Logger) {
	if cfg.OutputFile == "" {
		logger.Progress("Streaming logs from %s (Press Ctrl+C to stop)", url)
		return
	}
	filePath := cfg.OutputFile
	var file *os.File
	var err error
	if cfg.AppendMode {
		file, err = os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	} else {
		file, err = os.Create(filePath)
	}
	if err != nil {
		logger.Fatal("Failed to create/open output file '%s': %v", filePath, err)
	}
	file.Close()
	if !cfg.AppendMode {
		if err := os.Truncate(filePath, 0); err != nil {
			logger.Warn("Failed to truncate output file: %v", err)
		}
	}
	logger.Progress("Streaming logs from %s (Press Ctrl+C to stop) -> %s", url, filePath)
}

func outputJSONEntry(entry interface{}, cfg Config) error {
	b, _ := json.Marshal(entry)
	return io.WriteOutput(io.WriteConfig{
		OutputWriter: cfg.OutputWriter,
		OutputFile:   cfg.OutputFile,
		AppendMode:   cfg.AppendMode,
		FollowMode:   cfg.FollowMode,
	}, string(b)+"\n")
}

func outputYAMLEntry(entry interface{}, cfg Config) error {
	b, err := yaml.Marshal(entry)
	if err != nil {
		return err
	}
	return io.WriteOutput(io.WriteConfig{
		OutputWriter: cfg.OutputWriter,
		OutputFile:   cfg.OutputFile,
		AppendMode:   cfg.AppendMode,
		FollowMode:   cfg.FollowMode,
	}, string(b)+"\n")
}

func outputCSVEntry(entry interface{}, cfg Config, headerWritten *bool, tracker *outputErrorTracker, csvFunc func(interface{}, Config, bool) error, errorMsg string, logger *core.Logger) {
	err := csvFunc(entry, cfg, !*headerWritten)
	tracker.handleError(err, logger, errorMsg)
	if err == nil {
		*headerWritten = true
	}
}

func outputEntry(entry interface{}, cfg Config, headerWritten *bool, tracker *outputErrorTracker, jsonOutput, yamlOutput, csvOutput bool, logger *core.Logger, csvFunc func(interface{}, Config, bool) error, csvErrorMsg string, formatFunc func(Config) error) {
	if jsonOutput {
		if err := outputJSONEntry(entry, cfg); err != nil {
			tracker.handleError(err, logger, "Failed to write to output file: %v")
		}
		return
	}
	if yamlOutput {
		if err := outputYAMLEntry(entry, cfg); err != nil {
			tracker.handleError(err, logger, "Failed to write to output file: %v")
		}
		return
	}
	if csvOutput {
		outputCSVEntry(entry, cfg, headerWritten, tracker, csvFunc, csvErrorMsg, logger)
		return
	}
	tracker.handleError(formatFunc(cfg), logger, "Failed to write entry: %v")
}

func OutputLogEntry(le chrome.LogEntry, cfg Config, headerWritten *bool, tracker *outputErrorTracker, jsonOutput, yamlOutput, csvOutput bool, formatter *OutputFormatter, logger *core.Logger) {
	outputEntry(le, cfg, headerWritten, tracker, jsonOutput, yamlOutput, csvOutput, logger,
		func(e interface{}, c Config, h bool) error {
			return formatter.FormatAndOutputLogCSVRow(e.(chrome.LogEntry), c, h)
		},
		"Failed to write CSV log row: %v",
		func(c Config) error { return formatter.FormatAndOutputLog(le, c) })
}

func OutputNetworkEntry(ne chrome.NetworkEntry, cfg Config, headerWritten *bool, tracker *outputErrorTracker, jsonOutput, yamlOutput, csvOutput bool, formatter *OutputFormatter, logger *core.Logger) {
	outputEntry(ne, cfg, headerWritten, tracker, jsonOutput, yamlOutput, csvOutput, logger,
		func(e interface{}, c Config, h bool) error {
			return formatter.FormatAndOutputNetworkCSVRow(e.(chrome.NetworkEntry), c, h)
		},
		"Failed to write CSV network row: %v",
		func(c Config) error { return formatter.FormatAndOutputNetwork(ne, c) })
}
