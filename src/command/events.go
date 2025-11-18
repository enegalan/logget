package command

import (
	"fmt"
	"strings"
	"sync"
	"time"

	chrome "logget/src/chrome"
	"logget/src/core"
	helpers "logget/src/helpers"

	"github.com/go-rod/rod/lib/proto"
)

func convertProtoHeadersForEvents(headers proto.NetworkHeaders) map[string]string {
	result := make(map[string]string, len(headers))
	for k, v := range headers {
		result[k] = fmt.Sprintf("%v", v)
	}
	return result
}

func captureVerboseHeaders(headers map[string]string, reqURL, targetURL string, verbose bool, requestCaptured *bool, requestHeaders *map[string]string) {
	if verbose && !*requestCaptured && reqURL == targetURL {
		*requestHeaders = headers
		*requestCaptured = true
	}
}

func handleNavigationError(err error, responseProtocol, url string, startTime time.Time, verbose bool, logger *core.Logger) {
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

func createEventHandlers(cfg Config, url string, state *executionState, syncMaps *helpers.SyncMaps) *chrome.EventHandlers {
	return &chrome.EventHandlers{
		OnLog: func(le chrome.LogEntry) {
			if cfg.ShowLogs {
				state.logs = append(state.logs, le)
			}
		},
		OnNetwork: nil,
		OnRequestWillBeSent: func(requestID string, method, reqURL string, headers map[string]string, startTime float64) {
			syncMaps.RequestHeadersMap.Store(requestID, headers)
			captureVerboseHeaders(headers, reqURL, url, cfg.Verbose, &state.requestCaptured, &state.requestHeaders)
		},
	}
}

func setupNetworkHandlerForResponse(evResp *proto.NetworkResponseReceived, handlers *chrome.EventHandlers, showNetwork bool, network *[]chrome.NetworkEntry, networkEntriesMap *sync.Map) {
	if !showNetwork {
		return
	}
	handlers.OnNetwork = func(ne chrome.NetworkEntry) {
		*network = append(*network, ne)
		networkEntriesMap.Store(string(evResp.RequestID), &(*network)[len(*network)-1])
	}
}

func setupEventListeners(chromeCtx *chrome.ChromeContext, cfg Config, url string, state *executionState, syncMaps *helpers.SyncMaps, handlers *chrome.EventHandlers) {
	showNetworkOrVerbose := cfg.ShowNetwork || cfg.Verbose
	showLogs := cfg.ShowLogs
	go chromeCtx.Page.EachEvent(func(ev *proto.NetworkRequestWillBeSent) {
		if showNetworkOrVerbose {
			chrome.ProcessNetworkEventRequestWillBeSent(ev, &syncMaps.RequestMethods, &syncMaps.RequestURLs, &syncMaps.RequestStartTimes, state.startTime, handlers)
		}
	}, func(ev *proto.NetworkRequestWillBeSentExtraInfo) {
		if showNetworkOrVerbose {
			if requestURL, ok := helpers.LoadStringFromSyncMap(&syncMaps.RequestURLs, string(ev.RequestID)); ok && requestURL == url {
				headers := convertProtoHeadersForEvents(ev.Headers)
				captureVerboseHeaders(headers, requestURL, url, cfg.Verbose, &state.requestCaptured, &state.requestHeaders)
			}
		}
	}, func(ev *proto.NetworkResponseReceived) {
		if showNetworkOrVerbose {
			processResponseEvents(ev, handlers, cfg, state, syncMaps)
		}
	}, func(ev *proto.NetworkLoadingFinished) {
		if showNetworkOrVerbose {
			processResponseEvents(ev, handlers, cfg, state, syncMaps)
		}
	}, func(ev *proto.NetworkLoadingFailed) {
		if showNetworkOrVerbose {
			processResponseEvents(ev, handlers, cfg, state, syncMaps)
		}
	}, func(ev *proto.LogEntryAdded) {
		if showLogs {
			chrome.ProcessLogEvent(ev, handlers)
		}
	}, func(ev *proto.RuntimeConsoleAPICalled) {
		if showLogs {
			chrome.ProcessLogEvent(ev, handlers)
		}
	}, func(ev *proto.RuntimeExceptionThrown) {
		if showLogs {
			chrome.ProcessLogEvent(ev, handlers)
		}
	})()
}

func processResponseEvents(ev interface{}, handlers *chrome.EventHandlers, cfg Config, state *executionState, syncMaps *helpers.SyncMaps) {
	switch e := ev.(type) {
	case *proto.NetworkResponseReceived:
		setupNetworkHandlerForResponse(e, handlers, cfg.ShowNetwork, &state.network, &syncMaps.NetworkEntriesMap)
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
		ne := chrome.ProcessNetworkEventResponseReceived(e, networkCfg, &syncMaps.RequestMethods, &syncMaps.RequestStartTimes, state.startTime, &syncMaps.NetworkEntriesMap, handlers)
		if ne != nil && cfg.Verbose && !state.responseCaptured {
			state.responseHeaders = ne.Headers
			state.responseCaptured = true
		}
		handlers.OnNetwork = nil
	case *proto.NetworkLoadingFinished:
		chrome.ProcessNetworkEventLoadingFinished(e, &syncMaps.NetworkEntriesMap, state.startTime, handlers)
	case *proto.NetworkLoadingFailed:
		ne := chrome.HandleLoadingFailedEvent(e, &syncMaps.RequestMethods, &syncMaps.RequestURLs)
		if ne != nil && cfg.ShowNetwork {
			state.network = append(state.network, *ne)
		}
	}
}
