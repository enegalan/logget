package helpers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"

	cdpnetwork "github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"gopkg.in/yaml.v3"
)

type OutputData struct {
	URL      string         `json:"url"`
	Logs     []LogEntry     `json:"logs,omitempty"`
	Network  []NetworkEntry `json:"network,omitempty"`
	Duration time.Duration  `json:"duration"`
}

type CommandConfig struct {
	ShowLogs             bool
	ShowNetwork          bool
	JSONOutput           bool
	YAMLOutput           bool
	CSVOutput            bool
	Timeout              int
	Wait                 int
	UserAgent            string
	Headers              []string
	Cookies              []string
	VersionFlag          bool
	Verbose              bool
	OutputFile           string
	AppendMode           bool
	Version              string
	FollowMode           bool
	FilterPattern        string
	ExcludePattern       string
	StatusPattern        string
	DomainPattern        string
	MimePattern          string
	RefreshInterval      int
	SkipSSLVerify        bool
	NoRotateFingerprints bool
	FingerprintInterval  int
	HAROutput            bool
	XHROnly              bool
	DocumentOnly         bool
	CssOnly              bool
	ScriptOnly           bool
	FontOnly             bool
	ImgOnly              bool
	MediaOnly            bool
	ManifestOnly         bool
	WebSocketOnly        bool
	NoColor              bool
	Quiet                bool
	MinSize              int64
	MaxSize              int64
}

type outputErrorTracker struct {
	count      int
	maxErrors  int
	suppressed bool
}

func newOutputErrorTracker() *outputErrorTracker {
	return &outputErrorTracker{
		maxErrors: 5,
	}
}

func (t *outputErrorTracker) handleError(err error, logger *Logger, message string) {
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

func compileRegexPattern(pattern string) *regexp.Regexp {
	if pattern == "" {
		return nil
	}
	if r, err := regexp.Compile(pattern); err == nil {
		return r
	}
	return nil
}

func compileRegexPatterns(filterPattern, excludePattern string) (*regexp.Regexp, *regexp.Regexp) {
	return compileRegexPattern(filterPattern), compileRegexPattern(excludePattern)
}

func captureVerboseHeaders(headers map[string]string, reqURL, targetURL string, verbose bool, requestCaptured *bool, requestHeaders *map[string]string) {
	if verbose && !*requestCaptured && reqURL == targetURL {
		*requestHeaders = headers
		*requestCaptured = true
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
	return GenerateHeaders(cfg, url)
}

func writeOutputAndLog(cfg Config, content string, outputType string, logger *Logger) {
	if err := WriteOutput(cfg, content); err != nil {
		logger.Fatal("Failed to write %s: %v", outputType, err)
	}
	if cfg.OutputFile != "" {
		LogOutputFileSuccess(cfg, outputType, logger)
	}
}

func writeFinalOutput(cfg Config, output OutputData, network []NetworkEntry, url string, startTime time.Time, responseProtocol string, statusCode int, duration time.Duration, logger *Logger, formatter *OutputFormatter, jsonOutput bool, yamlOutput bool, verbose bool, showLogs bool, showNetwork bool, requestCaptured bool, requestHeaders map[string]string, responseHeaders map[string]string) {
	if cfg.HAROutput {
		harData, err := ConvertNetworkEntriesToHAR(network, url, startTime)
		if err != nil {
			logger.Fatal("Failed to generate HAR: %v", err)
		}
		writeOutputAndLog(cfg, string(harData)+"\n", "HAR output", logger)
		return
	}
	if jsonOutput {
		jsonData, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			logger.Fatal("Failed to marshal JSON: %v", err)
		}
		writeOutputAndLog(cfg, string(jsonData)+"\n", "JSON output", logger)
		return
	}
	if yamlOutput {
		yamlData, err := yaml.Marshal(output)
		if err != nil {
			logger.Fatal("Failed to marshal YAML: %v", err)
		}
		writeOutputAndLog(cfg, string(yamlData)+"\n", "YAML output", logger)
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

func prepareOutputFile(cfg Config, url string, logger *Logger) {
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
	return WriteOutput(cfg, string(b)+"\n")
}

func outputYAMLEntry(entry interface{}, cfg Config) error {
	b, err := yaml.Marshal(entry)
	if err != nil {
		return err
	}
	return WriteOutput(cfg, string(b)+"\n")
}

func outputCSVEntry(entry interface{}, cfg Config, headerWritten *bool, tracker *outputErrorTracker, csvFunc func(interface{}, Config, bool) error, errorMsg string, logger *Logger) {
	err := csvFunc(entry, cfg, !*headerWritten)
	tracker.handleError(err, logger, errorMsg)
	if err == nil {
		*headerWritten = true
	}
}

func outputEntry(entry interface{}, cfg Config, headerWritten *bool, tracker *outputErrorTracker, jsonOutput, yamlOutput, csvOutput bool, formatter *OutputFormatter, logger *Logger, csvFunc func(interface{}, Config, bool) error, csvErrorMsg string, formatFunc func(Config) error) {
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

func outputLogEntry(le LogEntry, cfg Config, headerWritten *bool, tracker *outputErrorTracker, jsonOutput, yamlOutput, csvOutput bool, formatter *OutputFormatter, logger *Logger) {
	outputEntry(le, cfg, headerWritten, tracker, jsonOutput, yamlOutput, csvOutput, formatter, logger,
		func(e interface{}, c Config, h bool) error {
			return formatter.FormatAndOutputLogCSVRow(e.(LogEntry), c, h)
		},
		"Failed to write CSV log row: %v",
		func(c Config) error { return formatter.FormatAndOutputLog(le, c) })
}

func outputNetworkEntry(ne NetworkEntry, cfg Config, headerWritten *bool, tracker *outputErrorTracker, jsonOutput, yamlOutput, csvOutput bool, formatter *OutputFormatter, logger *Logger) {
	outputEntry(ne, cfg, headerWritten, tracker, jsonOutput, yamlOutput, csvOutput, formatter, logger,
		func(e interface{}, c Config, h bool) error {
			return formatter.FormatAndOutputNetworkCSVRow(e.(NetworkEntry), c, h)
		},
		"Failed to write CSV network row: %v",
		func(c Config) error { return formatter.FormatAndOutputNetwork(ne, c) })
}

func handleNavigationError(err error, responseProtocol, url string, startTime time.Time, verbose bool, logger *Logger) {
	if strings.Contains(err.Error(), "ERR_HTTP_RESPONSE_CODE_FAILURE") {
		if verbose {
			fmt.Printf("logget: %s Error (navigation failed)\n", responseProtocol)
			fmt.Printf("logget: Duration: %v\n", time.Since(startTime))
			fmt.Println()
		}
		logger.Fatal("Navigation failed: %v", err)
	}
	logger.Fatal("Failed to navigate to %s: %v", url, err)
}

func shouldIncludeNetworkEntry(ne NetworkEntry, filterRegex *regexp.Regexp, excludeRegex *regexp.Regexp, minSize, maxSize int64) bool {
	if !ShouldShowLine(ne.URL, filterRegex, excludeRegex) {
		return false
	}
	if minSize > 0 && ne.Size < minSize {
		return false
	}
	if maxSize > 0 && ne.Size > maxSize {
		return false
	}
	return true
}

func setupNetworkHandlerForResponse(evResp *cdpnetwork.EventResponseReceived, handlers *EventHandlers, showNetwork bool, network *[]NetworkEntry, networkEntriesMap *sync.Map) {
	if !showNetwork {
		return
	}
	handlers.OnNetwork = func(ne NetworkEntry) {
		*network = append(*network, NetworkEntry(ne))
		networkEntriesMap.Store(evResp.RequestID.String(), &(*network)[len(*network)-1])
	}
}

func buildConfig(cmdConfig CommandConfig) Config {
	return Config{
		UserAgent:           cmdConfig.UserAgent,
		Headers:             cmdConfig.Headers,
		Cookies:             cmdConfig.Cookies,
		OutputFile:          cmdConfig.OutputFile,
		AppendMode:          cmdConfig.AppendMode,
		FollowMode:          cmdConfig.FollowMode,
		SkipSSLVerify:       cmdConfig.SkipSSLVerify,
		ShowNetwork:         cmdConfig.ShowNetwork,
		ShowLogs:            cmdConfig.ShowLogs,
		JSONOutput:          cmdConfig.JSONOutput,
		YAMLOutput:          cmdConfig.YAMLOutput,
		FilterPattern:       cmdConfig.FilterPattern,
		ExcludePattern:      cmdConfig.ExcludePattern,
		StatusPattern:       cmdConfig.StatusPattern,
		DomainPattern:       cmdConfig.DomainPattern,
		MimePattern:         cmdConfig.MimePattern,
		XHROnly:             cmdConfig.XHROnly,
		DocumentOnly:        cmdConfig.DocumentOnly,
		CssOnly:             cmdConfig.CssOnly,
		ScriptOnly:          cmdConfig.ScriptOnly,
		FontOnly:            cmdConfig.FontOnly,
		ImgOnly:             cmdConfig.ImgOnly,
		MediaOnly:           cmdConfig.MediaOnly,
		ManifestOnly:        cmdConfig.ManifestOnly,
		WebSocketOnly:       cmdConfig.WebSocketOnly,
		MinSize:             cmdConfig.MinSize,
		MaxSize:             cmdConfig.MaxSize,
		RotateFingerprints:  !cmdConfig.NoRotateFingerprints,
		FingerprintInterval: cmdConfig.FingerprintInterval,
		HAROutput:           cmdConfig.HAROutput,
	}
}

func validateOutputFormats(cmdConfig CommandConfig, logger *Logger) {
	formats := []bool{cmdConfig.JSONOutput, cmdConfig.YAMLOutput, cmdConfig.CSVOutput, cmdConfig.HAROutput}
	formatCount := 0
	for _, f := range formats {
		if f {
			formatCount++
		}
	}
	if formatCount > 1 {
		fmt.Println("logget: Only one output format can be specified at a time")
		os.Exit(1)
	}
}
func RunLogget(cmdConfig CommandConfig, url string) {
	logger := NewLogger(cmdConfig.Verbose, !cmdConfig.NoColor)
	logger.SetQuiet(cmdConfig.Quiet)
	formatter := NewOutputFormatter(!cmdConfig.NoColor)
	if cmdConfig.VersionFlag {
		logger.PrintHeader(cmdConfig.Version)
		os.Exit(0)
	}
	if url == "" {
		logger.PrintError(fmt.Errorf("no URL specified"))
		logger.PrintUsage()
		os.Exit(1)
	}
	validateOutputFormats(cmdConfig, logger)
	cfg := buildConfig(cmdConfig)
	if !cmdConfig.ShowLogs && !cmdConfig.ShowNetwork && !cmdConfig.Verbose && !cmdConfig.JSONOutput && !cmdConfig.YAMLOutput && !cmdConfig.HAROutput && !cmdConfig.FollowMode {
		logger.PrintUsage()
		os.Exit(0)
	}
	url = NormalizeURL(url)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cmdConfig.Timeout)*time.Millisecond)
	defer cancel()
	chromeCtx, chromeCancel, err := CreateChromeContext(ctx, cmdConfig.SkipSSLVerify)
	if err != nil {
		logger.Fatal("Failed to create Chrome context: %v", err)
	}
	defer chromeCancel()
	initialProtocol, initialStatusCode, err := GetInitialResponse(cfg, url)
	if err != nil {
		logger.Warn("Failed to get initial response: %v", err)
		initialProtocol = "HTTP/1.1"
		initialStatusCode = 200
	}
	var logs []LogEntry
	var network []NetworkEntry
	var responseProtocol string = initialProtocol
	var responseStatusCode int = initialStatusCode
	responseHeaders := make(map[string]string)
	responseCaptured := false
	requestHeaders := make(map[string]string)
	requestCaptured := false
	startTime := time.Now()
	if err := EnableChromeDomains(chromeCtx, cmdConfig.ShowLogs, cmdConfig.ShowNetwork || cmdConfig.Verbose); err != nil {
		if cmdConfig.ShowLogs {
			logger.Fatal("Failed to enable log domains: %v", err)
		} else {
			logger.Error("Failed to enable network domain: %v", err)
		}
	}
	requestMethods := sync.Map{}
	requestHeadersMap := sync.Map{}
	requestURLs := sync.Map{}
	requestStartTimes := sync.Map{}
	networkEntriesMap := sync.Map{}
	handlers := &EventHandlers{
		OnLog: func(le LogEntry) {
			if cmdConfig.ShowLogs {
				logs = append(logs, LogEntry(le))
			}
		},
		OnNetwork: nil,
		OnRequestWillBeSent: func(requestID string, method, reqURL string, headers map[string]string, startTime float64) {
			requestHeadersMap.Store(requestID, headers)
			captureVerboseHeaders(headers, reqURL, url, cmdConfig.Verbose, &requestCaptured, &requestHeaders)
		},
	}
	chromedp.ListenTarget(chromeCtx, func(ev interface{}) {
		if cmdConfig.ShowNetwork || cmdConfig.Verbose {
			if evReq, ok := ev.(*cdpnetwork.EventRequestWillBeSent); ok {
				ProcessNetworkEventRequestWillBeSent(evReq, &requestMethods, &requestURLs, &requestStartTimes, startTime, handlers)
			}
			if evExtra, ok := ev.(*cdpnetwork.EventRequestWillBeSentExtraInfo); ok {
				if requestURL, ok := LoadStringFromSyncMap(&requestURLs, evExtra.RequestID.String()); ok && requestURL == url {
					headers := ConvertEventHeaders(evExtra.Headers)
					captureVerboseHeaders(headers, requestURL, url, cmdConfig.Verbose, &requestCaptured, &requestHeaders)
				}
			}
		}
		if cmdConfig.ShowLogs {
			ProcessLogEvent(ev, handlers)
		}
		if cmdConfig.ShowNetwork || cmdConfig.Verbose {
			if evResp, ok := ev.(*cdpnetwork.EventResponseReceived); ok {
				setupNetworkHandlerForResponse(evResp, handlers, cmdConfig.ShowNetwork, &network, &networkEntriesMap)
				ne := ProcessNetworkEventResponseReceived(evResp, cfg, &requestMethods, &requestStartTimes, startTime, &networkEntriesMap, handlers)
				if ne != nil && cmdConfig.Verbose && !responseCaptured {
					responseHeaders = ne.Headers
					responseCaptured = true
				}
				handlers.OnNetwork = nil
			}
			if evFinished, ok := ev.(*cdpnetwork.EventLoadingFinished); ok {
				ProcessNetworkEventLoadingFinished(evFinished, &networkEntriesMap, startTime, handlers)
			}
			if evFailed, ok := ev.(*cdpnetwork.EventLoadingFailed); ok {
				ne := HandleLoadingFailedEvent(evFailed, &requestMethods, &requestURLs)
				if ne != nil && cmdConfig.ShowNetwork {
					network = append(network, NetworkEntry(*ne))
				}
			}
		}
	})
	if len(cmdConfig.Headers) > 0 || cmdConfig.UserAgent != "" {
		if err := SetHeaders(chromeCtx, cmdConfig.UserAgent, cmdConfig.Headers); err != nil {
			logger.Fatal("Failed to set headers: %v", err)
		}
	}
	if len(cmdConfig.Cookies) > 0 {
		if err := SetCookies(chromeCtx, url, cmdConfig.Cookies); err != nil {
			logger.Fatal("Failed to set cookies: %v", err)
		}
	}
	if cmdConfig.FollowMode {
		prepareOutputFile(cfg, url, logger)
		filterRegex, excludeRegex := compileRegexPatterns(cmdConfig.FilterPattern, cmdConfig.ExcludePattern)
		headerWrittenLogs := false
		headerWrittenNet := false
		outputErrorTracker := newOutputErrorTracker()
		onLog := func(le LogEntry) {
			if !ShouldShowLine(le.Message, filterRegex, excludeRegex) {
				return
			}
			outputLogEntry(le, cfg, &headerWrittenLogs, outputErrorTracker, cmdConfig.JSONOutput, cmdConfig.YAMLOutput, cmdConfig.CSVOutput, formatter, logger)
		}
		onNet := func(ne NetworkEntry) {
			if !shouldIncludeNetworkEntry(ne, filterRegex, excludeRegex, cfg.MinSize, cfg.MaxSize) {
				return
			}
			outputNetworkEntry(ne, cfg, &headerWrittenNet, outputErrorTracker, cmdConfig.JSONOutput, cmdConfig.YAMLOutput, cmdConfig.CSVOutput, formatter, logger)
		}
		sigCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()
		if err := StreamLogsRealTime(cfg, sigCtx, url, onLog, onNet); err != nil {
			logger.Fatal("Error streaming logs: %v", err)
		}
		return
	}
	logger.Progress("Navigating to %s...", url)
	tasks := []chromedp.Action{
		chromedp.Navigate(url),
		chromedp.Sleep(time.Duration(cmdConfig.Wait) * time.Millisecond),
	}
	if err = chromedp.Run(chromeCtx, tasks...); err != nil {
		handleNavigationError(err, responseProtocol, url, startTime, cmdConfig.Verbose, logger)
	}
	if cfg.RotateFingerprints {
		if err := StartFingerprintRotation(chromeCtx, cfg.FingerprintInterval); err != nil {
			logger.Warn("Failed to start fingerprint rotation: %v", err)
		}
	}
	logger.Success("Successfully loaded page: %s", url)
	statusCode := responseStatusCode
	duration := time.Since(startTime)
	output := OutputData{
		URL:      url,
		Logs:     logs,
		Network:  network,
		Duration: duration,
	}
	writeFinalOutput(cfg, output, network, url, startTime, responseProtocol, statusCode, duration, logger, formatter, cmdConfig.JSONOutput, cmdConfig.YAMLOutput, cmdConfig.Verbose, cmdConfig.ShowLogs, cmdConfig.ShowNetwork, requestCaptured, requestHeaders, responseHeaders)
}

func FormatUnknownFlag(flag string, isShort bool) string {
	if isShort {
		return "option -" + flag + ": is unknown"
	}
	return "option --" + flag + ": is unknown"
}

func extractFlagName(flagStr string) string {
	flagStr = strings.Trim(flagStr, "'\"")
	if idx := strings.Index(flagStr, " "); idx != -1 {
		flagStr = flagStr[:idx]
	}
	if idx := strings.Index(flagStr, " in"); idx != -1 {
		flagStr = flagStr[:idx]
	}
	return flagStr
}

func FormatCobraError(err error) string {
	errStr := err.Error()
	if strings.Contains(errStr, "unknown flag: --") {
		flag := extractFlagName(strings.TrimPrefix(errStr, "unknown flag: --"))
		return FormatUnknownFlag(flag, false)
	}
	if strings.Contains(errStr, "unknown shorthand flag: ") {
		if startIdx := strings.Index(errStr, "'"); startIdx != -1 {
			if endIdx := strings.Index(errStr[startIdx+1:], "'"); endIdx != -1 {
				flag := extractFlagName(errStr[startIdx+1 : startIdx+1+endIdx])
				return FormatUnknownFlag(flag, true)
			}
		}
		flag := extractFlagName(strings.TrimPrefix(errStr, "unknown shorthand flag: "))
		return FormatUnknownFlag(flag, true)
	}
	return errStr
}
