package command

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	chrome "logget/src/chrome"
	"logget/src/core"
	helpers "logget/src/helpers"

	cdpnetwork "github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

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

func setupNetworkHandlerForResponse(evResp *cdpnetwork.EventResponseReceived, handlers *chrome.EventHandlers, showNetwork bool, network *[]chrome.NetworkEntry, networkEntriesMap *sync.Map) {
	if !showNetwork {
		return
	}
	handlers.OnNetwork = func(ne chrome.NetworkEntry) {
		*network = append(*network, ne)
		networkEntriesMap.Store(evResp.RequestID.String(), &(*network)[len(*network)-1])
	}
}

func setupEventListeners(chromeCtx context.Context, cfg Config, url string, state *executionState, syncMaps *helpers.SyncMaps, handlers *chrome.EventHandlers) {
	showNetworkOrVerbose := cfg.ShowNetwork || cfg.Verbose
	showLogs := cfg.ShowLogs
	chromedp.ListenTarget(chromeCtx, func(ev interface{}) {
		if showNetworkOrVerbose {
			if evReq, ok := ev.(*cdpnetwork.EventRequestWillBeSent); ok {
				chrome.ProcessNetworkEventRequestWillBeSent(evReq, &syncMaps.RequestMethods, &syncMaps.RequestURLs, &syncMaps.RequestStartTimes, state.startTime, handlers)
			}
			if evExtra, ok := ev.(*cdpnetwork.EventRequestWillBeSentExtraInfo); ok {
				if requestURL, ok := helpers.LoadStringFromSyncMap(&syncMaps.RequestURLs, evExtra.RequestID.String()); ok && requestURL == url {
					headers := chrome.ConvertEventHeaders(evExtra.Headers)
					captureVerboseHeaders(headers, requestURL, url, cfg.Verbose, &state.requestCaptured, &state.requestHeaders)
				}
			}
		}
		if showLogs {
			chrome.ProcessLogEvent(ev, handlers)
		}
		if showNetworkOrVerbose {
			processResponseEvents(ev, handlers, cfg, state, syncMaps)
		}
	})
}

func processResponseEvents(ev interface{}, handlers *chrome.EventHandlers, cfg Config, state *executionState, syncMaps *helpers.SyncMaps) {
	if evResp, ok := ev.(*cdpnetwork.EventResponseReceived); ok {
		setupNetworkHandlerForResponse(evResp, handlers, cfg.ShowNetwork, &state.network, &syncMaps.NetworkEntriesMap)
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
		ne := chrome.ProcessNetworkEventResponseReceived(evResp, networkCfg, &syncMaps.RequestMethods, &syncMaps.RequestStartTimes, state.startTime, &syncMaps.NetworkEntriesMap, handlers)
		if ne != nil && cfg.Verbose && !state.responseCaptured {
			state.responseHeaders = ne.Headers
			state.responseCaptured = true
		}
		handlers.OnNetwork = nil
	}
	if evFinished, ok := ev.(*cdpnetwork.EventLoadingFinished); ok {
		chrome.ProcessNetworkEventLoadingFinished(evFinished, &syncMaps.NetworkEntriesMap, state.startTime, handlers)
	}
	if evFailed, ok := ev.(*cdpnetwork.EventLoadingFailed); ok {
		ne := chrome.HandleLoadingFailedEvent(evFailed, &syncMaps.RequestMethods, &syncMaps.RequestURLs)
		if ne != nil && cfg.ShowNetwork {
			state.network = append(state.network, *ne)
		}
	}
}
