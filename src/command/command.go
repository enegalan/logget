package command

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	helpers "logget/src"
	chrome "logget/src/chrome"
	"logget/src/flags"

	cdpnetwork "github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

func RunLogget(cmdConfig CommandConfig, url string) {
	logger, formatter, cfg := initializeCommand(cmdConfig, url)
	chromeCtx, chromeCancel, url := setupChromeContext(cmdConfig, url, logger)
	defer chromeCancel()
	initialProtocol, initialStatusCode := getInitialResponse(cfg, url, logger)
	state := initializeState(initialProtocol, initialStatusCode)
	enableChromeDomains(chromeCtx, cmdConfig, logger)
	syncMaps := initializeSyncMaps()
	handlers := createEventHandlers(cmdConfig, url, state, syncMaps)
	setupEventListeners(chromeCtx, cmdConfig, url, cfg, state, syncMaps, handlers)
	configureHeadersAndCookies(chromeCtx, cmdConfig, url, logger)
	if cmdConfig.FollowMode {
		runFollowMode(cfg, cmdConfig, url, logger, formatter)
		return
	}
	runNormalMode(chromeCtx, cmdConfig, cfg, url, state, logger, formatter)
}

func initializeCommand(cmdConfig CommandConfig, url string) (*helpers.Logger, *helpers.OutputFormatter, helpers.Config) {
	logger := helpers.NewLogger(cmdConfig.Verbose, !cmdConfig.NoColor)
	logger.SetQuiet(cmdConfig.Quiet)
	formatter := helpers.NewOutputFormatter(!cmdConfig.NoColor)
	if cmdConfig.VersionFlag {
		logger.PrintHeader(cmdConfig.Version)
		os.Exit(0)
	}
	if url == "" {
		logger.PrintError(fmt.Errorf("no URL specified"))
		logger.PrintUsage()
		os.Exit(1)
	}
	validateOutputFormats(cmdConfig)
	cfg := buildConfig(cmdConfig)
	hasAnyOutput := cmdConfig.ShowLogs || cmdConfig.ShowNetwork || cmdConfig.Verbose || cmdConfig.JSONOutput || cmdConfig.YAMLOutput || cmdConfig.HAROutput || cmdConfig.FollowMode
	if !hasAnyOutput {
		logger.PrintUsage()
		os.Exit(0)
	}
	return logger, formatter, cfg
}

func setupChromeContext(cmdConfig CommandConfig, url string, logger *helpers.Logger) (context.Context, context.CancelFunc, string) {
	url = helpers.NormalizeURL(url)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cmdConfig.Timeout)*time.Millisecond)
	chromeCtx, chromeCancel, err := chrome.CreateChromeContext(ctx, cmdConfig.SkipSSLVerify)
	if err != nil {
		cancel()
		logger.Fatal("Failed to create Chrome context: %v", err)
	}
	return chromeCtx, func() {
		chromeCancel()
		cancel()
	}, url
}

func getInitialResponse(cfg helpers.Config, url string, logger *helpers.Logger) (string, int) {
	initialProtocol, initialStatusCode, err := helpers.GetInitialResponse(cfg, url)
	if err != nil {
		logger.Warn("Failed to get initial response: %v", err)
		initialProtocol = "HTTP/1.1"
		initialStatusCode = 200
	}
	return initialProtocol, initialStatusCode
}

type executionState struct {
	logs               []chrome.LogEntry
	network            []chrome.NetworkEntry
	responseProtocol   string
	responseStatusCode int
	responseHeaders    map[string]string
	responseCaptured   bool
	requestHeaders     map[string]string
	requestCaptured    bool
	startTime          time.Time
}

func initializeState(initialProtocol string, initialStatusCode int) *executionState {
	return &executionState{
		logs:               make([]chrome.LogEntry, 0),
		network:            make([]chrome.NetworkEntry, 0),
		responseProtocol:   initialProtocol,
		responseStatusCode: initialStatusCode,
		responseHeaders:    make(map[string]string),
		responseCaptured:   false,
		requestHeaders:     make(map[string]string),
		requestCaptured:    false,
		startTime:          time.Now(),
	}
}

func enableChromeDomains(chromeCtx context.Context, cmdConfig CommandConfig, logger *helpers.Logger) {
	showNetworkOrVerbose := cmdConfig.ShowNetwork || cmdConfig.Verbose
	if err := chrome.EnableChromeDomains(chromeCtx, cmdConfig.ShowLogs, showNetworkOrVerbose); err != nil {
		if cmdConfig.ShowLogs {
			logger.Fatal("Failed to enable log domains: %v", err)
		} else {
			logger.Error("Failed to enable network domain: %v", err)
		}
	}
}

type syncMaps struct {
	requestMethods    sync.Map
	requestHeadersMap sync.Map
	requestURLs       sync.Map
	requestStartTimes sync.Map
	networkEntriesMap sync.Map
}

func initializeSyncMaps() *syncMaps { return &syncMaps{} }

func createEventHandlers(cmdConfig CommandConfig, url string, state *executionState, syncMaps *syncMaps) *chrome.EventHandlers {
	return &chrome.EventHandlers{
		OnLog: func(le chrome.LogEntry) {
			if cmdConfig.ShowLogs {
				state.logs = append(state.logs, le)
			}
		},
		OnNetwork: nil,
		OnRequestWillBeSent: func(requestID string, method, reqURL string, headers map[string]string, startTime float64) {
			syncMaps.requestHeadersMap.Store(requestID, headers)
			captureVerboseHeaders(headers, reqURL, url, cmdConfig.Verbose, &state.requestCaptured, &state.requestHeaders)
		},
	}
}

func setupEventListeners(chromeCtx context.Context, cmdConfig CommandConfig, url string, cfg helpers.Config, state *executionState, syncMaps *syncMaps, handlers *chrome.EventHandlers) {
	showNetworkOrVerbose := cmdConfig.ShowNetwork || cmdConfig.Verbose
	showLogs := cmdConfig.ShowLogs
	chromedp.ListenTarget(chromeCtx, func(ev interface{}) {
		if showNetworkOrVerbose {
			if evReq, ok := ev.(*cdpnetwork.EventRequestWillBeSent); ok {
				chrome.ProcessNetworkEventRequestWillBeSent(evReq, &syncMaps.requestMethods, &syncMaps.requestURLs, &syncMaps.requestStartTimes, state.startTime, handlers)
			}
			if evExtra, ok := ev.(*cdpnetwork.EventRequestWillBeSentExtraInfo); ok {
				if requestURL, ok := chrome.LoadStringFromSyncMap(&syncMaps.requestURLs, evExtra.RequestID.String()); ok && requestURL == url {
					headers := chrome.ConvertEventHeaders(evExtra.Headers)
					captureVerboseHeaders(headers, requestURL, url, cmdConfig.Verbose, &state.requestCaptured, &state.requestHeaders)
				}
			}
		}
		if showLogs {
			chrome.ProcessLogEvent(ev, handlers)
		}
		if showNetworkOrVerbose {
			processResponseEvents(ev, handlers, cmdConfig, cfg, state, syncMaps)
		}
	})
}

func processResponseEvents(ev interface{}, handlers *chrome.EventHandlers, cmdConfig CommandConfig, cfg helpers.Config, state *executionState, syncMaps *syncMaps) {
	if evResp, ok := ev.(*cdpnetwork.EventResponseReceived); ok {
		setupNetworkHandlerForResponse(evResp, handlers, cmdConfig.ShowNetwork, &state.network, &syncMaps.networkEntriesMap)
		networkCfg := chrome.StreamNetworkConfig{
			XHROnly:       cfg.XHROnly,
			DocumentOnly:  cfg.DocumentOnly,
			CssOnly:       cfg.CssOnly,
			ScriptOnly:    cfg.ScriptOnly,
			FontOnly:      cfg.FontOnly,
			ImgOnly:       cfg.ImgOnly,
			MediaOnly:     cfg.MediaOnly,
			ManifestOnly:  cfg.ManifestOnly,
			WebSocketOnly: cfg.WebSocketOnly,
			MimeRegex:     cfg.MimeRegex,
			StatusRegex:   cfg.StatusRegex,
			DomainRegex:   cfg.DomainRegex,
			MinSize:       cfg.MinSize,
			MaxSize:       cfg.MaxSize,
			ShowNetwork:   cfg.ShowNetwork,
			ShowLogs:      cfg.ShowLogs,
		}
		ne := chrome.ProcessNetworkEventResponseReceived(evResp, networkCfg, &syncMaps.requestMethods, &syncMaps.requestStartTimes, state.startTime, &syncMaps.networkEntriesMap, handlers)
		if ne != nil && cmdConfig.Verbose && !state.responseCaptured {
			state.responseHeaders = ne.Headers
			state.responseCaptured = true
		}
		handlers.OnNetwork = nil
	}
	if evFinished, ok := ev.(*cdpnetwork.EventLoadingFinished); ok {
		chrome.ProcessNetworkEventLoadingFinished(evFinished, &syncMaps.networkEntriesMap, state.startTime, handlers)
	}
	if evFailed, ok := ev.(*cdpnetwork.EventLoadingFailed); ok {
		ne := chrome.HandleLoadingFailedEvent(evFailed, &syncMaps.requestMethods, &syncMaps.requestURLs)
		if ne != nil && cmdConfig.ShowNetwork {
			state.network = append(state.network, *ne)
		}
	}
}

func configureHeadersAndCookies(chromeCtx context.Context, cmdConfig CommandConfig, url string, logger *helpers.Logger) {
	if len(cmdConfig.Headers) > 0 || cmdConfig.UserAgent != "" {
		if err := helpers.SetHeaders(chromeCtx, cmdConfig.UserAgent, cmdConfig.Headers); err != nil {
			logger.Fatal("Failed to set headers: %v", err)
		}
	}
	if len(cmdConfig.Cookies) > 0 {
		if err := helpers.SetCookies(chromeCtx, url, cmdConfig.Cookies); err != nil {
			logger.Fatal("Failed to set cookies: %v", err)
		}
	}
}

func runFollowMode(cfg helpers.Config, cmdConfig CommandConfig, url string, logger *helpers.Logger, formatter *helpers.OutputFormatter) {
	prepareOutputFile(cfg, url, logger)
	if cfg.OutputFile != "" {
		outputWriter, err := helpers.NewOutputWriter(cfg.OutputFile, cfg.AppendMode)
		if err != nil {
			logger.Fatal("Failed to create output writer: %v", err)
		}
		cfg.OutputWriter = outputWriter
		defer outputWriter.Close()
	}
	filterRegex, excludeRegex := compileRegexPatterns(cmdConfig.FilterPattern, cmdConfig.ExcludePattern)
	headerWrittenLogs := false
	headerWrittenNet := false
	outputErrorTracker := newOutputErrorTracker()
	onLog := func(le chrome.LogEntry) {
		if !helpers.ShouldShowLine(le.Message, filterRegex, excludeRegex) {
			return
		}
		outputLogEntry(le, cfg, &headerWrittenLogs, outputErrorTracker, cmdConfig.JSONOutput, cmdConfig.YAMLOutput, cmdConfig.CSVOutput, formatter, logger)
	}
	onNet := func(ne chrome.NetworkEntry) {
		if !shouldIncludeNetworkEntry(ne, filterRegex, excludeRegex, cfg.MinSize, cfg.MaxSize) {
			return
		}
		outputNetworkEntry(ne, cfg, &headerWrittenNet, outputErrorTracker, cmdConfig.JSONOutput, cmdConfig.YAMLOutput, cmdConfig.CSVOutput, formatter, logger)
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
	}
	if err := chrome.StreamLogsRealTime(streamCfg, sigCtx, url, onLog, onNet, helpers.SetHeaders, helpers.SetCookies, helpers.StartFingerprintRotation); err != nil {
		logger.Fatal("Error streaming logs: %v", err)
	}
}

func runNormalMode(chromeCtx context.Context, cmdConfig CommandConfig, cfg helpers.Config, url string, state *executionState, logger *helpers.Logger, formatter *helpers.OutputFormatter) {
	logger.Progress("Navigating to %s...", url)
	tasks := []chromedp.Action{
		chromedp.Navigate(url),
		chromedp.Sleep(time.Duration(cmdConfig.Wait) * time.Millisecond),
	}
	if err := chromedp.Run(chromeCtx, tasks...); err != nil {
		handleNavigationError(err, state.responseProtocol, url, state.startTime, cmdConfig.Verbose, logger)
	}
	if cfg.RotateFingerprints {
		if err := helpers.StartFingerprintRotation(chromeCtx, cfg.FingerprintInterval); err != nil {
			logger.Warn("Failed to start fingerprint rotation: %v", err)
		}
	}
	logger.Success("Successfully loaded page: %s", url)
	statusCode := state.responseStatusCode
	duration := time.Since(state.startTime)
	output := OutputData{
		URL:      url,
		Logs:     state.logs,
		Network:  state.network,
		Duration: duration,
	}
	writeFinalOutput(cfg, output, state.network, url, state.startTime, state.responseProtocol, statusCode, duration, logger, formatter, cmdConfig.JSONOutput, cmdConfig.YAMLOutput, cmdConfig.Verbose, cmdConfig.ShowLogs, cmdConfig.ShowNetwork, state.requestCaptured, state.requestHeaders, state.responseHeaders)
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
