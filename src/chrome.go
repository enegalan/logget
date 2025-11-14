package helpers

import (
	"context"
	"encoding/json"
	"fmt"
	neturl "net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	cdplog "github.com/chromedp/cdproto/log"
	cdpnetwork "github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
)

func GetChromeOptions(skipSSLVerify bool) []chromedp.ExecAllocatorOption {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("disable-plugins", true),
		chromedp.Flag("disable-web-security", true),
		chromedp.Flag("disable-features", "VizDisplayCompositor"),
		chromedp.Flag("ignore-certificate-errors", true),
		chromedp.Flag("ignore-ssl-errors", true),
		chromedp.Flag("allow-running-insecure-content", true),
		chromedp.Flag("disable-certificate-verification", true),
		chromedp.Flag("disable-background-networking", true),
		chromedp.Flag("disable-background-timer-throttling", true),
		chromedp.Flag("disable-breakpad", true),
		chromedp.Flag("disable-client-side-phishing-detection", true),
		chromedp.Flag("disable-default-apps", true),
		chromedp.Flag("disable-hang-monitor", true),
		chromedp.Flag("disable-popup-blocking", true),
		chromedp.Flag("disable-prompt-on-repost", true),
		chromedp.Flag("disable-sync", true),
		chromedp.Flag("disable-translate", true),
		chromedp.Flag("metrics-recording-only", true),
		chromedp.Flag("no-first-run", true),
		chromedp.Flag("safebrowsing-disable-auto-update", true),
		chromedp.Flag("enable-automation", false),
		chromedp.Flag("password-store", "basic"),
		chromedp.Flag("use-mock-keychain", true),
	)
	if skipSSLVerify {
		opts = append(opts,
			chromedp.Flag("ignore-certificate-errors-spki-list", true),
			chromedp.Flag("ignore-ssl-errors", true),
			chromedp.Flag("ignore-certificate-errors", true),
		)
	}
	return opts
}

func ConvertEventHeaders(headersMap map[string]interface{}) map[string]string {
	headers := make(map[string]string, len(headersMap))
	for name, value := range headersMap {
		if str, ok := value.(string); ok {
			headers[name] = str
		} else {
			headers[name] = fmt.Sprintf("%v", value)
		}
	}
	return headers
}

type LogEntry struct {
	Level   string    `json:"level"`
	Message string    `json:"message"`
	Time    time.Time `json:"time"`
	Source  string    `json:"source"`
}

type NetworkEntry struct {
	URL                          string            `json:"url"`
	Method                       string            `json:"method"`
	Status                       int               `json:"status"`
	Headers                      map[string]string `json:"headers"`
	Timestamp                    time.Time         `json:"timestamp"`
	Type                         string            `json:"type"`
	Size                         int64             `json:"size"`
	ResourceType                 string            `json:"resourceType"`
	Error                        string            `json:"error,omitempty"`
	ErrorType                    string            `json:"errorType,omitempty"`
	Duration                     float64           `json:"duration,omitempty"`
	DurationFormatted            string            `json:"durationFormatted,omitempty"` // Formatted representation (ms or s)
	TimeToFirstByte              float64           `json:"timeToFirstByte,omitempty"`
	TimeToFirstByteFormatted     string            `json:"timeToFirstByteFormatted,omitempty"`
	ConnectTime                  float64           `json:"connectTime,omitempty"`
	ConnectTimeFormatted         string            `json:"connectTimeFormatted,omitempty"`
	DNSLookupTime                float64           `json:"dnsLookupTime,omitempty"`
	DNSLookupTimeFormatted       string            `json:"dnsLookupTimeFormatted,omitempty"`
	SSLTime                      float64           `json:"sslTime,omitempty"`
	SSLTimeFormatted             string            `json:"sslTimeFormatted,omitempty"`
	ReceiveTime                  float64           `json:"receiveTime,omitempty"`
	ReceiveTimeFormatted         string            `json:"receiveTimeFormatted,omitempty"`
	ContentDownloadTime          float64           `json:"contentDownloadTime,omitempty"` // Full content download time (like Chrome DevTools)
	ContentDownloadTimeFormatted string            `json:"contentDownloadTimeFormatted,omitempty"`
	SendTime                     float64           `json:"sendTime,omitempty"`
	SendTimeFormatted            string            `json:"sendTimeFormatted,omitempty"`
	WaitTime                     float64           `json:"waitTime,omitempty"`
	WaitTimeFormatted            string            `json:"waitTimeFormatted,omitempty"`
	RequestStartTime             float64           `json:"requestStartTime,omitempty"`
	RequestStartTimeFormatted    string            `json:"requestStartTimeFormatted,omitempty"`
	ResponseStartTime            float64           `json:"responseStartTime,omitempty"`
	ResponseStartTimeFormatted   string            `json:"responseStartTimeFormatted,omitempty"`
	QueuedTime                   float64           `json:"queuedTime,omitempty"` // Time from navigation start to request start
	QueuedTimeFormatted          string            `json:"queuedTimeFormatted,omitempty"`
	Total                        float64           `json:"total,omitempty"`          // Sum of all timing phases in milliseconds
	TotalFormatted               string            `json:"totalFormatted,omitempty"` // Formatted representation (ms or s)
}

func ShouldIncludeNetworkEvent(cfg Config, ev *cdpnetwork.EventResponseReceived) bool {
	typeChecks := []struct {
		enabled bool
		check   func() bool
	}{
		{cfg.XHROnly, func() bool {
			return ev.Type == cdpnetwork.ResourceTypeXHR || ev.Type == cdpnetwork.ResourceTypeFetch
		}},
		{cfg.DocumentOnly, func() bool { return ev.Type == cdpnetwork.ResourceTypeDocument }},
		{cfg.CssOnly, func() bool { return ev.Type == cdpnetwork.ResourceTypeStylesheet }},
		{cfg.ScriptOnly, func() bool { return ev.Type == cdpnetwork.ResourceTypeScript }},
		{cfg.FontOnly, func() bool { return ev.Type == cdpnetwork.ResourceTypeFont }},
		{cfg.ImgOnly, func() bool { return ev.Type == cdpnetwork.ResourceTypeImage }},
		{cfg.MediaOnly, func() bool { return ev.Type == cdpnetwork.ResourceTypeMedia }},
		{cfg.ManifestOnly, func() bool { return ev.Type == cdpnetwork.ResourceTypeManifest }},
		{cfg.WebSocketOnly, func() bool { return ev.Type == cdpnetwork.ResourceTypeWebSocket }},
	}
	for _, tc := range typeChecks {
		if tc.enabled && !tc.check() {
			return false
		}
	}
	patternChecks := []struct {
		regex *regexp.Regexp
		check func() bool
	}{
		{cfg.MimeRegex, func() bool {
			if cfg.MimeRegex == nil {
				return true
			}
			if ev.Response == nil {
				return false
			}
			mimeCandidate := strings.ToLower(string(ev.Response.MimeType))
			if ctRaw, ok := ev.Response.Headers["Content-Type"]; ok {
				if s, ok := ctRaw.(string); ok && s != "" {
					mimeCandidate = strings.ToLower(s)
				}
			}
			return cfg.MimeRegex.MatchString(mimeCandidate)
		}},
		{cfg.StatusRegex, func() bool {
			if cfg.StatusRegex == nil {
				return true
			}
			if ev.Response == nil {
				return false
			}
			return cfg.StatusRegex.MatchString(strconv.Itoa(int(ev.Response.Status)))
		}},
		{cfg.DomainRegex, func() bool {
			if cfg.DomainRegex == nil {
				return true
			}
			if ev.Response == nil {
				return false
			}
			if u, err := neturl.Parse(ev.Response.URL); err == nil {
				return cfg.DomainRegex.MatchString(u.Hostname())
			}
			return false
		}},
	}
	for _, pc := range patternChecks {
		if pc.regex != nil && !pc.check() {
			return false
		}
	}
	if cfg.MinSize > 0 || cfg.MaxSize > 0 {
		if ev.Response == nil {
			return false
		}
		size := int64(ev.Response.EncodedDataLength)
		if cfg.MinSize > 0 && size < cfg.MinSize {
			return false
		}
		if cfg.MaxSize > 0 && size > cfg.MaxSize {
			return false
		}
	}
	return true
}

func FormatTiming(ms float64) string {
	if ms <= 0 {
		return ""
	}
	if ms >= 1000 {
		return strconv.FormatFloat(ms/1000.0, 'f', 2, 64) + "s"
	}
	return strconv.FormatFloat(ms, 'f', 2, 64) + "ms"
}

var navigationStartTime float64 = -1

var responseStartTimesMap = sync.Map{}

var consoleAPIMessages = sync.Map{}

var exceptionMessages = sync.Map{}

func StoreResponseStartTime(requestID string, responseStartTime float64) {
	responseStartTimesMap.Store(requestID, responseStartTime)
}

func GetResponseStartTime(requestID string) (float64, bool) {
	if val, ok := responseStartTimesMap.Load(requestID); ok {
		if timeVal, ok := val.(float64); ok {
			return timeVal, true
		}
	}
	return 0, false
}

func calcTimeDiff(start, end float64) float64 {
	if start >= 0 && end >= 0 {
		return end - start
	}
	return 0
}

func BuildNetworkEntryFromEvent(ev *cdpnetwork.EventResponseReceived, requestMethod string, requestTiming *cdpnetwork.ResourceTiming, responseTime float64, requestTime float64) NetworkEntry {
	response := ev.Response
	headers := ConvertEventHeaders(response.Headers)
	method := getDefaultMethod(requestMethod)
	var duration, ttfb, connectTime, dnsTime, sslTime, receiveTime, sendTime, waitTime float64
	var requestStartTime, responseStartTime float64
	if requestTiming != nil {
		if requestTiming.RequestTime >= 0 {
			requestStartTime = requestTiming.RequestTime
		}
		dnsTime = calcTimeDiff(requestTiming.DNSStart, requestTiming.DNSEnd)
		connectTime = calcTimeDiff(requestTiming.ConnectStart, requestTiming.ConnectEnd)
		sslTime = calcTimeDiff(requestTiming.SslStart, requestTiming.SslEnd)
		sendTime = calcTimeDiff(requestTiming.SendStart, requestTiming.SendEnd)
		waitTime = calcTimeDiff(requestTiming.SendEnd, requestTiming.ReceiveHeadersStart)
		receiveTime = calcTimeDiff(requestTiming.ReceiveHeadersStart, requestTiming.ReceiveHeadersEnd)
		if requestTiming.SendStart >= 0 && requestTiming.ReceiveHeadersStart >= requestTiming.SendStart {
			ttfb = requestTiming.ReceiveHeadersStart - requestTiming.SendStart
		}
		if requestTiming.ReceiveHeadersEnd >= 0 && requestTiming.RequestTime >= 0 && requestTiming.ReceiveHeadersEnd >= requestTiming.RequestTime {
			duration = requestTiming.ReceiveHeadersEnd - requestTiming.RequestTime
		} else if requestTime > 0 && responseTime > 0 {
			duration = responseTime - requestTime
		}
		if requestTiming.ReceiveHeadersStart >= 0 {
			responseStartTime = requestTiming.ReceiveHeadersStart
		}
	} else {
		if requestTime > 0 && responseTime > 0 {
			duration = responseTime - requestTime
			requestStartTime = requestTime
			responseStartTime = responseTime
		}
	}
	var queuedTime float64
	if requestTiming != nil {
		if requestTiming.RequestTime >= 0 {
			if ev.Type == cdpnetwork.ResourceTypeDocument && navigationStartTime < 0 {
				navigationStartTime = requestTiming.RequestTime
				queuedTime = 0
			} else if navigationStartTime >= 0 {
				queuedTime = requestTiming.RequestTime - navigationStartTime
			} else {
				queuedTime = requestTiming.RequestTime
				navigationStartTime = requestTiming.RequestTime
			}
		}
		if requestTiming.SendStart >= 0 {
			requestStartTime = requestTiming.SendStart
		} else if requestTiming.RequestTime >= 0 {
			requestStartTime = requestTiming.RequestTime
		}
	}
	total := queuedTime + dnsTime + connectTime + sslTime + sendTime + waitTime + receiveTime

	ne := NetworkEntry{
		URL:               response.URL,
		Method:            method,
		Status:            int(response.Status),
		Headers:           headers,
		Timestamp:         time.Now(),
		Type:              string(response.MimeType),
		Size:              int64(response.EncodedDataLength),
		ResourceType:      ev.Type.String(),
		Duration:          duration,
		TimeToFirstByte:   ttfb,
		ConnectTime:       connectTime,
		DNSLookupTime:     dnsTime,
		SSLTime:           sslTime,
		ReceiveTime:       receiveTime,
		SendTime:          sendTime,
		WaitTime:          waitTime,
		RequestStartTime:  requestStartTime,
		ResponseStartTime: responseStartTime,
		QueuedTime:        queuedTime,
		Total:             total,
	}
	return ne
}

func CategorizeError(errorText string) string {
	errorTextLower := strings.ToLower(errorText)
	if strings.Contains(errorTextLower, "blocked") && strings.Contains(errorTextLower, "origin") {
		return "cors"
	}
	errorPatterns := []struct {
		category string
		patterns []string
		useLower bool
	}{
		{"timeout", []string{"timeout", "ERR_TIMED_OUT", "ERR_NET_TIMED_OUT"}, false},
		{"dns", []string{"dns", "ERR_NAME_NOT_RESOLVED", "ERR_NAME_RESOLUTION_FAILED", "ERR_DNS_"}, false},
		{"cors", []string{"cors", "cross-origin", "ERR_BLOCKED_BY_CLIENT"}, true},
		{"connection_refused", []string{"ERR_CONNECTION_REFUSED"}, false},
		{"connection_reset", []string{"ERR_CONNECTION_RESET"}, false},
		{"connection_closed", []string{"ERR_CONNECTION_CLOSED"}, false},
		{"network_changed", []string{"ERR_NETWORK_CHANGED"}, false},
		{"ssl", []string{"ERR_SSL_"}, false},
		{"certificate", []string{"ERR_CERT_"}, false},
	}
	for _, ep := range errorPatterns {
		text := errorText
		if ep.useLower {
			text = errorTextLower
		}
		for _, pattern := range ep.patterns {
			if strings.Contains(text, pattern) {
				return ep.category
			}
		}
	}
	return "unknown"
}

func getDefaultMethod(method string) string {
	if method == "" {
		return "GET"
	}
	return method
}

func BuildNetworkEntryFromErrorEvent(ev *cdpnetwork.EventLoadingFailed, requestMethod string, requestURL string) NetworkEntry {
	return NetworkEntry{
		URL:          requestURL,
		Method:       getDefaultMethod(requestMethod),
		Status:       0,
		Headers:      make(map[string]string),
		Timestamp:    time.Now(),
		Type:         "",
		Size:         0,
		ResourceType: ev.Type.String(),
		Error:        ev.ErrorText,
		ErrorType:    CategorizeError(ev.ErrorText),
	}
}

func HandleLoadingFailedEvent(ev *cdpnetwork.EventLoadingFailed, requestMethods *sync.Map, requestURLs *sync.Map) *NetworkEntry {
	requestID := ev.RequestID.String()
	method, _ := LoadStringFromSyncMap(requestMethods, requestID)
	requestURL, ok := LoadStringFromSyncMap(requestURLs, requestID)
	if !ok || requestURL == "" {
		return nil
	}
	ne := BuildNetworkEntryFromErrorEvent(ev, method, requestURL)
	return &ne
}

type EventHandlers struct {
	OnLog               func(LogEntry)
	OnNetwork           func(NetworkEntry)
	OnRequestWillBeSent func(requestID string, method, url string, headers map[string]string, startTime float64)
}

func isJavaScriptException(message string) bool {
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?:^|\s)(https?://[^\s:]+):(\d+):(\d+)`), // URL:line:col
		regexp.MustCompile(`at\s+(https?://[^\s:]+):(\d+):(\d+)`),    // at URL:line:col
	}
	for _, re := range patterns {
		if matches := re.FindStringSubmatch(message); len(matches) >= 3 {
			locationKey := matches[1] + ":" + matches[2]
			if _, exists := exceptionMessages.Load(locationKey); exists {
				return true
			}
		}
	}
	return false
}

func isChromeInternalMessage(message string) bool {
	_, exists := consoleAPIMessages.Load(message)
	return !exists
}

func formatFrame(frame *runtime.CallFrame) string {
	if frame.URL == "" && frame.FunctionName == "" {
		return ""
	}
	if frame.FunctionName != "" {
		if frame.URL != "" {
			return fmt.Sprintf("\n    at %s (%s:%d:%d)", frame.FunctionName, frame.URL, frame.LineNumber+1, frame.ColumnNumber+1)
		}
		return fmt.Sprintf("\n    at %s", frame.FunctionName)
	}
	return fmt.Sprintf("\n    at %s:%d:%d", frame.URL, frame.LineNumber+1, frame.ColumnNumber+1)
}

func storeExceptionLocation(ed *runtime.ExceptionDetails) {
	if ed == nil {
		return
	}
	var locationKey string
	if ed.URL != "" {
		locationKey = fmt.Sprintf("%s:%d", ed.URL, ed.LineNumber+1)
	} else if ed.StackTrace != nil && len(ed.StackTrace.CallFrames) > 0 {
		frame := ed.StackTrace.CallFrames[0]
		if frame.URL != "" {
			locationKey = fmt.Sprintf("%s:%d", frame.URL, frame.LineNumber+1)
		}
	}
	if locationKey != "" {
		exceptionMessages.Store(locationKey, true)
	}
}

func normalizeLogLevel(level string, message string) string {
	if strings.ToUpper(level) == "WARNING" && isJavaScriptException(message) {
		return "error"
	}
	return level
}

func createLogEntry(level, message, source string) LogEntry {
	return LogEntry{Level: level, Message: message, Time: time.Now(), Source: source}
}

func ProcessLogEvent(ev interface{}, handlers *EventHandlers) {
	if handlers == nil || handlers.OnLog == nil {
		return
	}
	if ev, ok := ev.(*runtime.EventExceptionThrown); ok {
		var message string
		if ev.ExceptionDetails != nil {
			if ev.ExceptionDetails.Exception != nil {
				if ev.ExceptionDetails.Exception.Description != "" {
					message = ev.ExceptionDetails.Exception.Description
				} else if ev.ExceptionDetails.Exception.ClassName != "" {
					message = ev.ExceptionDetails.Exception.ClassName
					if ev.ExceptionDetails.Text != "" {
						message += ": " + ev.ExceptionDetails.Text
					}
				}
			}
			if message == "" {
				message = ev.ExceptionDetails.Text
			}
			if ev.ExceptionDetails.StackTrace != nil && len(ev.ExceptionDetails.StackTrace.CallFrames) > 0 {
				for i, frame := range ev.ExceptionDetails.StackTrace.CallFrames {
					message += formatFrame(frame)
					if i == len(ev.ExceptionDetails.StackTrace.CallFrames)-1 && ev.ExceptionDetails.StackTrace.Parent != nil {
						parent := ev.ExceptionDetails.StackTrace.Parent
						for _, parentFrame := range parent.CallFrames {
							message += formatFrame(parentFrame)
						}
					}
				}
			} else if ev.ExceptionDetails.URL != "" {
				message += fmt.Sprintf(" at %s:%d:%d", ev.ExceptionDetails.URL, ev.ExceptionDetails.LineNumber+1, ev.ExceptionDetails.ColumnNumber+1)
			}
		}
		if message == "" {
			message = "Unhandled JavaScript exception"
		}
		storeExceptionLocation(ev.ExceptionDetails)
		handlers.OnLog(createLogEntry("error", message, "browser"))
		return
	}
	if ev, ok := ev.(*cdplog.EventEntryAdded); ok {
		if isChromeInternalMessage(ev.Entry.Text) {
			return
		}
		level := normalizeLogLevel(ev.Entry.Level.String(), ev.Entry.Text)
		handlers.OnLog(createLogEntry(level, ev.Entry.Text, "browser"))
		return
	}
	if ev, ok := ev.(*runtime.EventConsoleAPICalled); ok {
		var message string
		for _, arg := range ev.Args {
			if arg.Value == nil {
				continue
			}
			var str string
			if err := json.Unmarshal(arg.Value, &str); err == nil {
				message += str + " "
			} else {
				message += fmt.Sprintf("%v ", arg.Value)
			}
		}
		message = strings.TrimSpace(message)
		if message != "" {
			consoleAPIMessages.Store(message, true)
		}
		level := normalizeLogLevel(ev.Type.String(), message)
		handlers.OnLog(createLogEntry(level, message, "console"))
	}
}

func ProcessNetworkEventRequestWillBeSent(ev *cdpnetwork.EventRequestWillBeSent, requestMethods *sync.Map, requestURLs *sync.Map, requestStartTimes *sync.Map, startTime time.Time, handlers *EventHandlers) {
	if ev.Request == nil {
		return
	}
	requestID := ev.RequestID.String()
	requestMethods.Store(requestID, ev.Request.Method)
	requestURLs.Store(requestID, ev.Request.URL)
	headers := ConvertEventHeaders(ev.Request.Headers)
	requestStartTime := float64(time.Since(startTime).Nanoseconds()) / 1e6
	requestStartTimes.Store(requestID, requestStartTime)
	if handlers != nil && handlers.OnRequestWillBeSent != nil {
		handlers.OnRequestWillBeSent(requestID, ev.Request.Method, ev.Request.URL, headers, requestStartTime)
	}
}

func ProcessNetworkEventResponseReceived(ev *cdpnetwork.EventResponseReceived, cfg Config, requestMethods *sync.Map, requestStartTimes *sync.Map, startTime time.Time, networkEntriesMap *sync.Map, handlers *EventHandlers) *NetworkEntry {
	if !ShouldIncludeNetworkEvent(cfg, ev) {
		return nil
	}
	requestID := ev.RequestID.String()
	method, _ := LoadStringFromSyncMap(requestMethods, requestID)
	var requestTiming *cdpnetwork.ResourceTiming
	var requestStartTime, responseTime float64
	if ev.Response != nil && ev.Response.Timing != nil {
		requestTiming = ev.Response.Timing
	}
	requestStartTime, _ = LoadFloat64FromSyncMap(requestStartTimes, requestID)
	responseTime = float64(time.Since(startTime).Nanoseconds()) / 1e6
	ne := BuildNetworkEntryFromEvent(ev, method, requestTiming, responseTime, requestStartTime)
	if ne.ResponseStartTime > 0 {
		responseStartTimesMap.Store(requestID, responseTime)
	}
	if handlers != nil && handlers.OnNetwork != nil {
		handlers.OnNetwork(ne)
	}
	if networkEntriesMap != nil {
		networkEntriesMap.Store(requestID, ne)
	}
	return &ne
}

func updateEntryContentDownloadTime(entry *NetworkEntry, contentDownloadTime float64) {
	entry.ContentDownloadTime = contentDownloadTime
	if entry.Duration > 0 && entry.ContentDownloadTime > 0 {
		entry.Duration += entry.ContentDownloadTime
	}
	entry.Total += entry.ContentDownloadTime
}

func ProcessNetworkEventLoadingFinished(ev *cdpnetwork.EventLoadingFinished, networkEntriesMap *sync.Map, startTime time.Time, handlers *EventHandlers) {
	requestID := ev.RequestID.String()
	entryVal, ok := networkEntriesMap.Load(requestID)
	if !ok {
		return
	}
	loadingFinishedTime := float64(time.Since(startTime).Nanoseconds()) / 1e6
	responseStartTimeVal, ok := responseStartTimesMap.Load(requestID)
	if !ok {
		return
	}
	responseStartTime, ok := responseStartTimeVal.(float64)
	if !ok {
		return
	}
	contentDownloadTime := loadingFinishedTime - responseStartTime
	var entry *NetworkEntry
	switch v := entryVal.(type) {
	case NetworkEntry:
		entry = &v
		updateEntryContentDownloadTime(entry, contentDownloadTime)
		networkEntriesMap.Store(requestID, *entry)
	case *NetworkEntry:
		entry = v
		updateEntryContentDownloadTime(entry, contentDownloadTime)
	default:
		return
	}
	if handlers != nil && handlers.OnNetwork != nil {
		handlers.OnNetwork(*entry)
	}
}

func ProcessNetworkEventLoadingFailed(ev *cdpnetwork.EventLoadingFailed, requestMethods *sync.Map, requestURLs *sync.Map, handlers *EventHandlers) {
	if ne := HandleLoadingFailedEvent(ev, requestMethods, requestURLs); ne != nil && handlers != nil && handlers.OnNetwork != nil {
		handlers.OnNetwork(*ne)
	}
}

func CreateChromeContext(ctx context.Context, skipSSLVerify bool) (context.Context, context.CancelFunc, error) {
	allocCtx, allocCancel := chromedp.NewExecAllocator(ctx, GetChromeOptions(skipSSLVerify)...)
	chromeCtx, chromeCancel := chromedp.NewContext(allocCtx)
	return chromeCtx, func() { chromeCancel(); allocCancel() }, nil
}

func EnableChromeDomains(ctx context.Context, showLogs, showNetwork bool) error {
	if showLogs {
		if err := chromedp.Run(ctx, cdplog.Enable()); err != nil {
			return fmt.Errorf("failed to enable log domain: %v", err)
		}
		if err := chromedp.Run(ctx, runtime.Enable()); err != nil {
			return fmt.Errorf("failed to enable runtime domain: %v", err)
		}
	}
	if showNetwork {
		if err := chromedp.Run(ctx, cdpnetwork.Enable()); err != nil {
			return fmt.Errorf("failed to enable network domain: %v", err)
		}
	}
	return nil
}

func StreamLogsRealTime(cfg Config, ctx context.Context, url string, onLog func(LogEntry), onNet func(NetworkEntry)) error {
	chromeCtx, cancel, err := CreateChromeContext(ctx, cfg.SkipSSLVerify)
	if err != nil {
		return err
	}
	defer cancel()
	if err := EnableChromeDomains(chromeCtx, cfg.ShowLogs, cfg.ShowNetwork); err != nil {
		return err
	}
	maps := GetNetworkMaps()
	pageStartTime := time.Now()
	handlers := &EventHandlers{
		OnLog:     onLog,
		OnNetwork: onNet,
	}
	chromedp.ListenTarget(chromeCtx, func(ev interface{}) {
		showNetwork := cfg.ShowNetwork
		showLogs := cfg.ShowLogs
		if showNetwork {
			if evReq, ok := ev.(*cdpnetwork.EventRequestWillBeSent); ok {
				ProcessNetworkEventRequestWillBeSent(evReq, &maps.Methods, &maps.URLs, &maps.StartTimes, pageStartTime, handlers)
			}
		}
		if showLogs {
			ProcessLogEvent(ev, handlers)
		}
		if showNetwork {
			if evResp, ok := ev.(*cdpnetwork.EventResponseReceived); ok {
				ProcessNetworkEventResponseReceived(evResp, cfg, &maps.Methods, &maps.StartTimes, pageStartTime, &maps.NetworkEntries, handlers)
			}
			if evFinished, ok := ev.(*cdpnetwork.EventLoadingFinished); ok {
				ProcessNetworkEventLoadingFinished(evFinished, &maps.NetworkEntries, pageStartTime, handlers)
			}
			if evFailed, ok := ev.(*cdpnetwork.EventLoadingFailed); ok {
				ProcessNetworkEventLoadingFailed(evFailed, &maps.Methods, &maps.URLs, handlers)
			}
		}
	})
	if len(cfg.Headers) > 0 || cfg.UserAgent != "" {
		if err := SetHeaders(chromeCtx, cfg.UserAgent, cfg.Headers); err != nil {
			return fmt.Errorf("failed to set headers: %v", err)
		}
	}
	if len(cfg.Cookies) > 0 {
		if err := SetCookies(chromeCtx, url, cfg.Cookies); err != nil {
			return fmt.Errorf("failed to set cookies: %v", err)
		}
	}
	if err := chromedp.Run(chromeCtx, chromedp.Navigate(url)); err != nil {
		return fmt.Errorf("failed to navigate to %s: %v", url, err)
	}
	if cfg.RotateFingerprints {
		if err := StartFingerprintRotation(chromeCtx, cfg.FingerprintInterval); err != nil {
			return fmt.Errorf("failed to start fingerprint rotation: %v", err)
		}
	}
	<-chromeCtx.Done()
	return nil
}
