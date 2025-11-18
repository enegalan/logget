package chrome

import (
	"fmt"
	neturl "net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-rod/rod/lib/proto"
)

func convertProtoHeaders(headers proto.NetworkHeaders) map[string]string {
	result := make(map[string]string, len(headers))
	for k, v := range headers {
		result[k] = fmt.Sprintf("%v", v)
	}
	return result
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
	DurationFormatted            string            `json:"durationFormatted,omitempty"`
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
	ContentDownloadTime          float64           `json:"contentDownloadTime,omitempty"`
	ContentDownloadTimeFormatted string            `json:"contentDownloadTimeFormatted,omitempty"`
	SendTime                     float64           `json:"sendTime,omitempty"`
	SendTimeFormatted            string            `json:"sendTimeFormatted,omitempty"`
	WaitTime                     float64           `json:"waitTime,omitempty"`
	WaitTimeFormatted            string            `json:"waitTimeFormatted,omitempty"`
	RequestStartTime             float64           `json:"requestStartTime,omitempty"`
	RequestStartTimeFormatted    string            `json:"requestStartTimeFormatted,omitempty"`
	ResponseStartTime            float64           `json:"responseStartTime,omitempty"`
	ResponseStartTimeFormatted   string            `json:"responseStartTimeFormatted,omitempty"`
	QueuedTime                   float64           `json:"queuedTime,omitempty"`
	QueuedTimeFormatted          string            `json:"queuedTimeFormatted,omitempty"`
	Total                        float64           `json:"total,omitempty"`
	TotalFormatted               string            `json:"totalFormatted,omitempty"`
}

var navigationStartTime float64 = -1

var responseStartTimesMap = sync.Map{}

type StreamNetworkConfig struct {
	XHROnly       bool
	DocumentOnly  bool
	CssOnly       bool
	ScriptOnly    bool
	FontOnly      bool
	ImgOnly       bool
	MediaOnly     bool
	ManifestOnly  bool
	WebSocketOnly bool
	MimeRegex     *regexp.Regexp
	StatusRegex   *regexp.Regexp
	DomainRegex   *regexp.Regexp
	MinSize       int64
	MaxSize       int64
	ShowNetwork   bool
	ShowLogs      bool
}

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

func FormatTiming(ms float64) string {
	if ms <= 0 {
		return ""
	}
	if ms >= 1000 {
		return strconv.FormatFloat(ms/1000.0, 'f', 2, 64) + "s"
	}
	return strconv.FormatFloat(ms, 'f', 2, 64) + "ms"
}

func ShouldIncludeNetworkEvent(cfg StreamNetworkConfig, ev *proto.NetworkResponseReceived) bool {
	typeChecks := []struct {
		enabled bool
		check   func() bool
	}{
		{cfg.XHROnly, func() bool {
			return ev.Type == proto.NetworkResourceTypeXHR || ev.Type == proto.NetworkResourceTypeFetch
		}},
		{cfg.DocumentOnly, func() bool { return ev.Type == proto.NetworkResourceTypeDocument }},
		{cfg.CssOnly, func() bool { return ev.Type == proto.NetworkResourceTypeStylesheet }},
		{cfg.ScriptOnly, func() bool { return ev.Type == proto.NetworkResourceTypeScript }},
		{cfg.FontOnly, func() bool { return ev.Type == proto.NetworkResourceTypeFont }},
		{cfg.ImgOnly, func() bool { return ev.Type == proto.NetworkResourceTypeImage }},
		{cfg.MediaOnly, func() bool { return ev.Type == proto.NetworkResourceTypeMedia }},
		{cfg.ManifestOnly, func() bool { return ev.Type == proto.NetworkResourceTypeManifest }},
		{cfg.WebSocketOnly, func() bool { return ev.Type == proto.NetworkResourceTypeWebSocket }},
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
			mimeCandidate := strings.ToLower(ev.Response.MIMEType)
			if ctRaw, ok := ev.Response.Headers["Content-Type"]; ok {
				ctStr := fmt.Sprintf("%v", ctRaw)
				if ctStr != "" {
					mimeCandidate = strings.ToLower(ctStr)
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

func calcTimeDiff(start, end float64) float64 {
	if start >= 0 && end >= 0 {
		return end - start
	}
	return 0
}

func getDefaultMethod(method string) string {
	if method == "" {
		return "GET"
	}
	return method
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

func BuildNetworkEntryFromEvent(ev *proto.NetworkResponseReceived, requestMethod string, requestTiming *proto.NetworkResourceTiming, responseTime float64, requestTime float64) NetworkEntry {
	response := ev.Response
	headers := convertProtoHeaders(response.Headers)
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
			if ev.Type == proto.NetworkResourceTypeDocument && navigationStartTime < 0 {
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
		Type:              response.MIMEType,
		Size:              int64(response.EncodedDataLength),
		ResourceType:      string(ev.Type),
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

func BuildNetworkEntryFromErrorEvent(ev *proto.NetworkLoadingFailed, requestMethod string, requestURL string) NetworkEntry {
	return NetworkEntry{
		URL:          requestURL,
		Method:       getDefaultMethod(requestMethod),
		Status:       0,
		Headers:      make(map[string]string),
		Timestamp:    time.Now(),
		Type:         "",
		Size:         0,
		ResourceType: string(ev.Type),
		Error:        ev.ErrorText,
		ErrorType:    CategorizeError(ev.ErrorText),
	}
}

func LoadStringFromSyncMap(m *sync.Map, key string) (string, bool) {
	if val, ok := m.Load(key); ok {
		if str, ok := val.(string); ok {
			return str, true
		}
	}
	return "", false
}

func LoadFloat64FromSyncMap(m *sync.Map, key string) (float64, bool) {
	if val, ok := m.Load(key); ok {
		if f, ok := val.(float64); ok {
			return f, true
		}
	}
	return 0, false
}

func HandleLoadingFailedEvent(ev *proto.NetworkLoadingFailed, requestMethods *sync.Map, requestURLs *sync.Map) *NetworkEntry {
	requestID := string(ev.RequestID)
	method, _ := LoadStringFromSyncMap(requestMethods, requestID)
	requestURL, ok := LoadStringFromSyncMap(requestURLs, requestID)
	if !ok || requestURL == "" {
		return nil
	}
	ne := BuildNetworkEntryFromErrorEvent(ev, method, requestURL)
	return &ne
}

func ProcessNetworkEventRequestWillBeSent(ev *proto.NetworkRequestWillBeSent, requestMethods *sync.Map, requestURLs *sync.Map, requestStartTimes *sync.Map, startTime time.Time, handlers *EventHandlers) {
	if ev.Request == nil {
		return
	}
	requestID := string(ev.RequestID)
	requestMethods.Store(requestID, ev.Request.Method)
	requestURLs.Store(requestID, ev.Request.URL)
	headers := convertProtoHeaders(ev.Request.Headers)
	requestStartTime := float64(time.Since(startTime).Nanoseconds()) / 1e6
	requestStartTimes.Store(requestID, requestStartTime)
	if handlers != nil && handlers.OnRequestWillBeSent != nil {
		handlers.OnRequestWillBeSent(requestID, ev.Request.Method, ev.Request.URL, headers, requestStartTime)
	}
}

func ProcessNetworkEventResponseReceived(ev *proto.NetworkResponseReceived, cfg StreamNetworkConfig, requestMethods *sync.Map, requestStartTimes *sync.Map, startTime time.Time, networkEntriesMap *sync.Map, handlers *EventHandlers) *NetworkEntry {
	if !ShouldIncludeNetworkEvent(cfg, ev) {
		return nil
	}
	requestID := string(ev.RequestID)
	method, _ := LoadStringFromSyncMap(requestMethods, requestID)
	var requestTiming *proto.NetworkResourceTiming
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

func ProcessNetworkEventLoadingFinished(ev *proto.NetworkLoadingFinished, networkEntriesMap *sync.Map, startTime time.Time, handlers *EventHandlers) {
	requestID := string(ev.RequestID)
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

func ProcessNetworkEventLoadingFailed(ev *proto.NetworkLoadingFailed, requestMethods *sync.Map, requestURLs *sync.Map, handlers *EventHandlers) {
	if ne := HandleLoadingFailedEvent(ev, requestMethods, requestURLs); ne != nil && handlers != nil && handlers.OnNetwork != nil {
		handlers.OnNetwork(*ne)
	}
}
