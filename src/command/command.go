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
	url = helpers.NormalizeURL(url)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cmdConfig.Timeout)*time.Millisecond)
	defer cancel()
	chromeCtx, chromeCancel, err := chrome.CreateChromeContext(ctx, cmdConfig.SkipSSLVerify)
	if err != nil {
		logger.Fatal("Failed to create Chrome context: %v", err)
	}
	defer chromeCancel()
	initialProtocol, initialStatusCode, err := helpers.GetInitialResponse(cfg, url)
	if err != nil {
		logger.Warn("Failed to get initial response: %v", err)
		initialProtocol = "HTTP/1.1"
		initialStatusCode = 200
	}
	logs := make([]chrome.LogEntry, 0, 100)
	network := make([]chrome.NetworkEntry, 0, 500)
	responseProtocol, responseStatusCode := initialProtocol, initialStatusCode
	responseHeaders := make(map[string]string)
	responseCaptured := false
	requestHeaders := make(map[string]string)
	requestCaptured := false
	startTime := time.Now()
	showNetworkOrVerbose := cmdConfig.ShowNetwork || cmdConfig.Verbose
	if err := chrome.EnableChromeDomains(chromeCtx, cmdConfig.ShowLogs, showNetworkOrVerbose); err != nil {
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
	handlers := &chrome.EventHandlers{
		OnLog: func(le chrome.LogEntry) {
			if cmdConfig.ShowLogs {
				logs = append(logs, le)
			}
		},
		OnNetwork: nil,
		OnRequestWillBeSent: func(requestID string, method, reqURL string, headers map[string]string, startTime float64) {
			requestHeadersMap.Store(requestID, headers)
			captureVerboseHeaders(headers, reqURL, url, cmdConfig.Verbose, &requestCaptured, &requestHeaders)
		},
	}
	showLogs := cmdConfig.ShowLogs
	chromedp.ListenTarget(chromeCtx, func(ev interface{}) {
		if showNetworkOrVerbose {
			if evReq, ok := ev.(*cdpnetwork.EventRequestWillBeSent); ok {
				chrome.ProcessNetworkEventRequestWillBeSent(evReq, &requestMethods, &requestURLs, &requestStartTimes, startTime, handlers)
			}
			if evExtra, ok := ev.(*cdpnetwork.EventRequestWillBeSentExtraInfo); ok {
				if requestURL, ok := chrome.LoadStringFromSyncMap(&requestURLs, evExtra.RequestID.String()); ok && requestURL == url {
					headers := chrome.ConvertEventHeaders(evExtra.Headers)
					captureVerboseHeaders(headers, requestURL, url, cmdConfig.Verbose, &requestCaptured, &requestHeaders)
				}
			}
		}
		if showLogs {
			chrome.ProcessLogEvent(ev, handlers)
		}
		if showNetworkOrVerbose {
			if evResp, ok := ev.(*cdpnetwork.EventResponseReceived); ok {
				setupNetworkHandlerForResponse(evResp, handlers, cmdConfig.ShowNetwork, &network, &networkEntriesMap)
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
				ne := chrome.ProcessNetworkEventResponseReceived(evResp, networkCfg, &requestMethods, &requestStartTimes, startTime, &networkEntriesMap, handlers)
				if ne != nil && cmdConfig.Verbose && !responseCaptured {
					responseHeaders = ne.Headers
					responseCaptured = true
				}
				handlers.OnNetwork = nil
			}
			if evFinished, ok := ev.(*cdpnetwork.EventLoadingFinished); ok {
				chrome.ProcessNetworkEventLoadingFinished(evFinished, &networkEntriesMap, startTime, handlers)
			}
			if evFailed, ok := ev.(*cdpnetwork.EventLoadingFailed); ok {
				ne := chrome.HandleLoadingFailedEvent(evFailed, &requestMethods, &requestURLs)
				if ne != nil && cmdConfig.ShowNetwork {
					network = append(network, *ne)
				}
			}
		}
	})
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
	if cmdConfig.FollowMode {
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
		if err := helpers.StartFingerprintRotation(chromeCtx, cfg.FingerprintInterval); err != nil {
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
