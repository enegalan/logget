package command

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	chrome "logget/src/chrome"
	"logget/src/colors"
	"logget/src/core"
	"logget/src/flags"
	helpers "logget/src/helpers"
	"logget/src/io"

	"gopkg.in/yaml.v3"
)

func RunLogget(cfg Config, url string) {
	logger, formatter, cfg := initCommand(cfg, url)
	chromeCtx, chromeCancel, url := setupChromeContext(cfg, url, logger)
	defer chromeCancel()
	initialProtocol, initialStatusCode := getInitialResponse(cfg, url, logger)
	state := initializeState(initialProtocol, initialStatusCode)
	syncMaps := helpers.NewSyncMaps()
	handlers := createEventHandlers(cfg, url, state, syncMaps)
	setupEventListeners(chromeCtx, cfg, url, state, syncMaps, handlers)
	setupHeadersAndCookies(chromeCtx, cfg, url, logger)
	if cfg.FollowMode {
		runFollowMode(cfg, url, logger, formatter)
		return
	}
	runNormalMode(chromeCtx, cfg, url, state, logger, formatter)
}

func initCommand(cfg Config, url string) (*core.Logger, *OutputFormatter, Config) {
	logger := core.NewLogger(cfg.Verbose, !cfg.NoColor)
	logger.SetQuiet(cfg.Quiet)
	formatter := NewOutputFormatter(!cfg.NoColor)
	if cfg.VersionFlag {
		logger.PrintHeader(cfg.Version)
		os.Exit(0)
	}
	if url == "" {
		logger.PrintError(fmt.Errorf("no URL specified"))
		logger.PrintUsage()
		os.Exit(1)
	}
	ValidateOutputFormats(cfg)
	cfg = compileConfigRegexp(cfg)
	hasAnyOutput := cfg.ShowLogs || cfg.ShowNetwork || cfg.Verbose || cfg.JSONOutput || cfg.YAMLOutput || cfg.HAROutput || cfg.FollowMode
	if !hasAnyOutput {
		logger.PrintUsage()
		os.Exit(0)
	}
	return logger, formatter, cfg
}

func setupChromeContext(cfg Config, url string, logger *core.Logger) (*chrome.ChromeContext, context.CancelFunc, string) {
	url = helpers.NormalizeURL(url)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.Timeout)*time.Millisecond)
	chromeCtx, chromeCancel, err := chrome.CreateChromeContext(ctx, cfg.SkipSSLVerify, cfg.Quiet)
	if err != nil {
		cancel()
		logger.Fatal("Failed to create Chrome context: %v", err)
	}
	return chromeCtx, func() {
		chromeCancel()
		cancel()
	}, url
}

func getInitialResponse(cfg Config, url string, logger *core.Logger) (string, int) {
	initialProtocol, initialStatusCode, err := chrome.GetInitialResponse(cfg.SkipSSLVerify, cfg.UserAgent, cfg.Headers, url)
	if err != nil {
		logger.Warn("Failed to get initial response: %v", err)
		initialProtocol = "HTTP/1.1"
		initialStatusCode = 200
	}
	return initialProtocol, initialStatusCode
}

func setupHeadersAndCookies(chromeCtx *chrome.ChromeContext, cfg Config, url string, logger *core.Logger) {
	if len(cfg.Headers) > 0 || cfg.UserAgent != "" {
		if err := core.SetHeaders(chromeCtx, cfg.UserAgent, cfg.Headers); err != nil {
			logger.Fatal("Failed to set headers: %v", err)
		}
	}
	if len(cfg.Cookies) > 0 {
		if err := core.SetCookies(chromeCtx, url, cfg.Cookies); err != nil {
			logger.Fatal("Failed to set cookies: %v", err)
		}
	}
}

func runFollowMode(cfg Config, url string, logger *core.Logger, formatter *OutputFormatter) {
	PrepareOutputFile(cfg, url, logger)
	if cfg.OutputFile != "" {
		outputWriter, err := io.NewOutputWriter(cfg.OutputFile, cfg.AppendMode)
		if err != nil {
			logger.Fatal("Failed to create output writer: %v", err)
		}
		cfg.OutputWriter = outputWriter
		defer outputWriter.Close()
	}
	filterRegex, excludeRegex := helpers.CompileRegexPatterns(cfg.FilterPattern, cfg.ExcludePattern)
	headerWrittenLogs := false
	headerWrittenNet := false
	outputErrorTracker := NewOutputErrorTracker()
	onLog := func(le chrome.LogEntry) {
		if !helpers.FilterLog(le, filterRegex, excludeRegex) {
			return
		}
		OutputLogEntry(le, cfg, &headerWrittenLogs, outputErrorTracker, cfg.JSONOutput, cfg.YAMLOutput, cfg.CSVOutput, formatter, logger)
	}
	onNet := func(ne chrome.NetworkEntry) {
		if !helpers.FilterNetworkEntry(ne, filterRegex, excludeRegex, cfg.MinSize, cfg.MaxSize) {
			return
		}
		OutputNetworkEntry(ne, cfg, &headerWrittenNet, outputErrorTracker, cfg.JSONOutput, cfg.YAMLOutput, cfg.CSVOutput, formatter, logger)
	}
	onJavaScriptResult := func(result interface{}, err error) {
		if err != nil {
			return
		}
		outputJSResult(result, cfg, logger)
	}
	sigCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	streamCfg := chrome.StreamConfig{
		SkipSSLVerify:       cfg.SkipSSLVerify,
		ShowLogs:            cfg.ShowLogs,
		ShowNetwork:         cfg.ShowNetwork,
		Headers:             cfg.Headers,
		Cookies:             cfg.Cookies,
		UserAgent:           cfg.UserAgent,
		RotateFingerprints:  cfg.RotateFingerprints,
		FingerprintInterval: cfg.FingerprintInterval,
		XHROnly:             cfg.XHROnly,
		DocumentOnly:        cfg.DocumentOnly,
		CssOnly:             cfg.CssOnly,
		ScriptOnly:          cfg.ScriptOnly,
		FontOnly:            cfg.FontOnly,
		ImgOnly:             cfg.ImgOnly,
		MediaOnly:           cfg.MediaOnly,
		ManifestOnly:        cfg.ManifestOnly,
		WebSocketOnly:       cfg.WebSocketOnly,
		MimeRegex:           cfg.MimeRegex,
		StatusRegex:         cfg.StatusRegex,
		DomainRegex:         cfg.DomainRegex,
		MinSize:             cfg.MinSize,
		MaxSize:             cfg.MaxSize,
		ExecuteJS:           cfg.ExecuteJS,
	}
	if cfg.RotateFingerprints {
		if cfg.FingerprintInterval <= 0 {
			logger.Warn("Fingerprint rotation interval is invalid, disabling rotation")
			streamCfg.RotateFingerprints = false
		}
	}
	if err := chrome.StreamLogsRealTime(streamCfg, sigCtx, url, onLog, onNet, core.SetHeaders, core.SetCookies, core.StartFingerprintRotation, core.ExecuteJavaScript, onJavaScriptResult); err != nil {
		logger.Fatal("Error streaming logs: %v", err)
	}
}

func runNormalMode(chromeCtx *chrome.ChromeContext, cfg Config, url string, state *executionState, logger *core.Logger, formatter *OutputFormatter) {
	logger.Progress("Navigating to %s...", url)
	if err := chromeCtx.Page.Navigate(url); err != nil {
		handleNavigationError(err, state.responseProtocol, url, state.startTime, cfg.Verbose, logger)
	}
	chromeCtx.Page.MustWaitLoad()
	time.Sleep(time.Duration(cfg.Wait) * time.Millisecond)
	if cfg.RotateFingerprints {
		if cfg.FingerprintInterval <= 0 {
			logger.Warn("Fingerprint rotation interval is invalid, disabling rotation")
		} else {
			if err := core.StartFingerprintRotation(chromeCtx, cfg.FingerprintInterval); err != nil {
				logger.Warn("Failed to start fingerprint rotation: %v", err)
			}
		}
	}
	if cfg.ExecuteJS != "" {
		result, err := core.ExecuteJavaScript(chromeCtx, cfg.ExecuteJS)
		if err != nil {
			logger.Error("%v", err)
		} else {
			outputJSResult(result, cfg, logger)
		}
	}
	logger.Success("Successfully loaded page: %s", url)
	statusCode := state.responseStatusCode
	duration := time.Since(state.startTime)
	filterRegex, excludeRegex := helpers.CompileRegexPatterns(cfg.FilterPattern, cfg.ExcludePattern)
	filteredLogs := helpers.FilterLogs(state.logs, filterRegex, excludeRegex)
	filteredNetwork := helpers.FilterNetworkEntries(state.network, filterRegex, excludeRegex, cfg.MinSize, cfg.MaxSize)
	outputData := OutputData{
		URL:      url,
		Logs:     filteredLogs,
		Network:  filteredNetwork,
		Duration: duration,
	}
	WriteFinalOutput(cfg, outputData, filteredNetwork, url, state.startTime, state.responseProtocol, statusCode, duration, logger, formatter, cfg.JSONOutput, cfg.YAMLOutput, cfg.Verbose, cfg.ShowLogs, cfg.ShowNetwork, state.requestCaptured, state.requestHeaders, state.responseHeaders)
}

func outputJSResult(result interface{}, cfg Config, logger *core.Logger) {
	if result == nil {
		return
	}
	var output strings.Builder
	if cfg.JSONOutput {
		jsonData, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			logger.Warn("Failed to marshal JavaScript result: %v", err)
			return
		}
		output.WriteString(string(jsonData))
		output.WriteString("\n")
	} else if cfg.YAMLOutput {
		yamlData, err := yaml.Marshal(result)
		if err != nil {
			logger.Warn("Failed to marshal JavaScript result: %v", err)
			return
		}
		output.WriteString(string(yamlData))
		output.WriteString("\n")
	} else {
		formatter := NewOutputFormatter(!cfg.NoColor)
		output.WriteString("\n=== ")
		headerText := formatter.theme.Bold("JAVASCRIPT RESULT")
		header := formatter.Colorize(colors.Cyan, headerText)
		output.WriteString(header)
		output.WriteString(" ===\n")
		jsonData, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			output.WriteString(fmt.Sprintf("%v\n", result))
		} else {
			output.WriteString(string(jsonData))
			output.WriteString("\n")
		}
	}
	writeOutputAndLog(cfg, output.String(), "JavaScript result", logger)
}

func FormatCobraError(err error) string {
	errStr := err.Error()
	if strings.Contains(errStr, "unknown flag: --") {
		return flags.FormatUnknownFlag(flags.ExtractFlagName(strings.TrimPrefix(errStr, "unknown flag: --")), false)
	}
	if strings.Contains(errStr, "unknown shorthand flag: ") {
		if startIdx := strings.Index(errStr, "'"); startIdx != -1 {
			if endIdx := strings.Index(errStr[startIdx+1:], "'"); endIdx != -1 {
				return flags.FormatUnknownFlag(flags.ExtractFlagName(errStr[startIdx+1:startIdx+1+endIdx]), true)
			}
		}
		return flags.FormatUnknownFlag(flags.ExtractFlagName(strings.TrimPrefix(errStr, "unknown shorthand flag: ")), true)
	}
	return errStr
}
