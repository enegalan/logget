package command

import (
	chrome "logget/src/chrome"
	"time"
)

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
