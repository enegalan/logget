package command

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	helpers "logget/src"
	chrome "logget/src/chrome"

	cdpnetwork "github.com/chromedp/cdproto/network"
)

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

func handleNavigationError(err error, responseProtocol, url string, startTime time.Time, verbose bool, logger *helpers.Logger) {
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

func shouldIncludeNetworkEntry(ne chrome.NetworkEntry, filterRegex *regexp.Regexp, excludeRegex *regexp.Regexp, minSize, maxSize int64) bool {
	if !helpers.ShouldShowLine(ne.URL, filterRegex, excludeRegex) {
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

func setupNetworkHandlerForResponse(evResp *cdpnetwork.EventResponseReceived, handlers *chrome.EventHandlers, showNetwork bool, network *[]chrome.NetworkEntry, networkEntriesMap *sync.Map) {
	if !showNetwork {
		return
	}
	handlers.OnNetwork = func(ne chrome.NetworkEntry) {
		*network = append(*network, ne)
		networkEntriesMap.Store(evResp.RequestID.String(), &(*network)[len(*network)-1])
	}
}
