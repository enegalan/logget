package helpers

import (
	"context"
	"encoding/json"
	"fmt"
	neturl "net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	cdplog "github.com/chromedp/cdproto/log"
	cdpnetwork "github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
)

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
	if cfg.XHROnly {
		if !(ev.Type == cdpnetwork.ResourceTypeXHR || ev.Type == cdpnetwork.ResourceTypeFetch) {
			return false
		}
	}
	if cfg.DocumentOnly {
		if ev.Type != cdpnetwork.ResourceTypeDocument {
			return false
		}
	}
	if cfg.CssOnly {
		if ev.Type != cdpnetwork.ResourceTypeStylesheet {
			return false
		}
	}
	if cfg.ScriptOnly {
		if ev.Type != cdpnetwork.ResourceTypeScript {
			return false
		}
	}
	if cfg.FontOnly {
		if ev.Type != cdpnetwork.ResourceTypeFont {
			return false
		}
	}
	if cfg.ImgOnly {
		if ev.Type != cdpnetwork.ResourceTypeImage {
			return false
		}
	}
	if cfg.MediaOnly {
		if ev.Type != cdpnetwork.ResourceTypeMedia {
			return false
		}
	}
	if cfg.ManifestOnly {
		if ev.Type != cdpnetwork.ResourceTypeManifest {
			return false
		}
	}
	if cfg.WebSocketOnly {
		if ev.Type != cdpnetwork.ResourceTypeWebSocket {
			return false
		}
	}
	if cfg.MimePattern != "" {
		if ev.Response == nil {
			return false
		}
		mimeCandidate := strings.ToLower(string(ev.Response.MimeType))
		if ctRaw, ok := ev.Response.Headers["Content-Type"]; ok {
			if s, ok := ctRaw.(string); ok && s != "" {
				mimeCandidate = strings.ToLower(s)
			}
		}
		if r, err := regexp.Compile(cfg.MimePattern); err == nil {
			if !r.MatchString(mimeCandidate) {
				return false
			}
		}
	}
	if cfg.StatusPattern != "" {
		if ev.Response == nil {
			return false
		}
		if r, err := regexp.Compile(cfg.StatusPattern); err == nil {
			if !r.MatchString(fmt.Sprintf("%d", int(ev.Response.Status))) {
				return false
			}
		}
	}
	if cfg.DomainPattern != "" {
		if ev.Response == nil {
			return false
		}
		if u, err := neturl.Parse(ev.Response.URL); err == nil {
			host := u.Hostname()
			if r, err := regexp.Compile(cfg.DomainPattern); err == nil {
				if !r.MatchString(host) {
					return false
				}
			}
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
		return fmt.Sprintf("%.2fs", ms/1000.0)
	}
	return fmt.Sprintf("%.2fms", ms)
}

var navigationStartTime float64 = -1

var responseStartTimesMap = sync.Map{}

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

func BuildNetworkEntryFromEvent(ev *cdpnetwork.EventResponseReceived, requestMethod string, requestTiming *cdpnetwork.ResourceTiming, responseTime float64, requestTime float64) NetworkEntry {
	response := ev.Response
	headers := make(map[string]string)
	for name, value := range response.Headers {
		if str, ok := value.(string); ok {
			headers[name] = str
		} else {
			headers[name] = fmt.Sprintf("%v", value)
		}
	}
	method := requestMethod
	if method == "" {
		method = "GET"
	}
	var duration, ttfb, connectTime, dnsTime, sslTime, receiveTime, sendTime, waitTime float64
	var requestStartTime, responseStartTime float64
	if requestTiming != nil {
		if requestTiming.RequestTime >= 0 {
			requestStartTime = requestTiming.RequestTime
		}
		if requestTiming.DNSStart >= 0 && requestTiming.DNSEnd >= 0 {
			dnsTime = requestTiming.DNSEnd - requestTiming.DNSStart
		}
		if requestTiming.ConnectStart >= 0 && requestTiming.ConnectEnd >= 0 {
			connectTime = requestTiming.ConnectEnd - requestTiming.ConnectStart
		}
		if requestTiming.SslStart >= 0 && requestTiming.SslEnd >= 0 {
			sslTime = requestTiming.SslEnd - requestTiming.SslStart
		}
		if requestTiming.SendStart >= 0 && requestTiming.SendEnd >= 0 {
			sendTime = requestTiming.SendEnd - requestTiming.SendStart
		}
		if requestTiming.SendEnd >= 0 && requestTiming.ReceiveHeadersStart >= 0 {
			waitTime = requestTiming.ReceiveHeadersStart - requestTiming.SendEnd
			if requestTiming.SendStart >= 0 && requestTiming.ReceiveHeadersStart >= requestTiming.SendStart {
				ttfb = requestTiming.ReceiveHeadersStart - requestTiming.SendStart
			}
		}
		if requestTiming.ReceiveHeadersStart >= 0 && requestTiming.ReceiveHeadersEnd >= 0 {
			receiveTime = requestTiming.ReceiveHeadersEnd - requestTiming.ReceiveHeadersStart
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

	return NetworkEntry{
		URL:                          response.URL,
		Method:                       method,
		Status:                       int(response.Status),
		Headers:                      headers,
		Timestamp:                    time.Now(),
		Type:                         string(response.MimeType),
		Size:                         int64(response.EncodedDataLength),
		ResourceType:                 ev.Type.String(),
		Duration:                     duration,
		DurationFormatted:            FormatTiming(duration),
		TimeToFirstByte:              ttfb,
		TimeToFirstByteFormatted:     FormatTiming(ttfb),
		ConnectTime:                  connectTime,
		ConnectTimeFormatted:         FormatTiming(connectTime),
		DNSLookupTime:                dnsTime,
		DNSLookupTimeFormatted:       FormatTiming(dnsTime),
		SSLTime:                      sslTime,
		SSLTimeFormatted:             FormatTiming(sslTime),
		ReceiveTime:                  receiveTime,
		ReceiveTimeFormatted:         FormatTiming(receiveTime),
		ContentDownloadTime:          0,
		ContentDownloadTimeFormatted: "",
		SendTime:                     sendTime,
		SendTimeFormatted:            FormatTiming(sendTime),
		WaitTime:                     waitTime,
		WaitTimeFormatted:            FormatTiming(waitTime),
		RequestStartTime:             requestStartTime,
		RequestStartTimeFormatted:    FormatTiming(requestStartTime),
		ResponseStartTime:            responseStartTime,
		ResponseStartTimeFormatted:   FormatTiming(responseStartTime),
		QueuedTime:                   queuedTime,
		QueuedTimeFormatted:          FormatTiming(queuedTime),
		Total:                        total,
		TotalFormatted:               FormatTiming(total),
	}
}

func CategorizeError(errorText string) string {
	errorTextLower := strings.ToLower(errorText)
	if strings.Contains(errorTextLower, "timeout") || strings.Contains(errorText, "ERR_TIMED_OUT") || strings.Contains(errorText, "ERR_NET_TIMED_OUT") {
		return "timeout"
	}
	if strings.Contains(errorTextLower, "dns") || strings.Contains(errorText, "ERR_NAME_NOT_RESOLVED") || strings.Contains(errorText, "ERR_NAME_RESOLUTION_FAILED") || strings.Contains(errorText, "ERR_DNS_") {
		return "dns"
	}
	if strings.Contains(errorTextLower, "cors") || strings.Contains(errorTextLower, "cross-origin") || strings.Contains(errorText, "ERR_BLOCKED_BY_CLIENT") || (strings.Contains(errorTextLower, "blocked") && strings.Contains(errorTextLower, "origin")) {
		return "cors"
	}
	if strings.Contains(errorText, "ERR_CONNECTION_REFUSED") {
		return "connection_refused"
	}
	if strings.Contains(errorText, "ERR_CONNECTION_RESET") {
		return "connection_reset"
	}
	if strings.Contains(errorText, "ERR_CONNECTION_CLOSED") {
		return "connection_closed"
	}
	if strings.Contains(errorText, "ERR_NETWORK_CHANGED") {
		return "network_changed"
	}
	if strings.Contains(errorText, "ERR_SSL_") {
		return "ssl"
	}
	if strings.Contains(errorText, "ERR_CERT_") {
		return "certificate"
	}
	return "unknown"
}

func BuildNetworkEntryFromErrorEvent(ev *cdpnetwork.EventLoadingFailed, requestMethod string, requestURL string) NetworkEntry {
	method := requestMethod
	if method == "" {
		method = "GET"
	}
	errorText := ev.ErrorText
	errorType := CategorizeError(errorText)
	return NetworkEntry{
		URL:          requestURL,
		Method:       method,
		Status:       0,
		Headers:      make(map[string]string),
		Timestamp:    time.Now(),
		Type:         "",
		Size:         0,
		ResourceType: ev.Type.String(),
		Error:        errorText,
		ErrorType:    errorType,
	}
}

func HandleLoadingFailedEvent(ev *cdpnetwork.EventLoadingFailed, requestMethods *sync.Map, requestURLs *sync.Map) *NetworkEntry {
	var method string
	var requestURL string
	if methodVal, ok := requestMethods.Load(ev.RequestID.String()); ok {
		if methodStr, ok := methodVal.(string); ok {
			method = methodStr
		}
	}
	if urlVal, ok := requestURLs.Load(ev.RequestID.String()); ok {
		if urlStr, ok := urlVal.(string); ok {
			requestURL = urlStr
		}
	}
	if requestURL == "" {
		return nil
	}
	ne := BuildNetworkEntryFromErrorEvent(ev, method, requestURL)
	return &ne
}

func StreamLogsRealTime(cfg Config, ctx context.Context, url string, onLog func(LogEntry), onNet func(NetworkEntry)) error {
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
	)
	if cfg.SkipSSLVerify {
		opts = append(opts,
			chromedp.Flag("ignore-certificate-errors-spki-list", true),
			chromedp.Flag("ignore-ssl-errors", true),
			chromedp.Flag("ignore-certificate-errors", true),
		)
	}
	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()
	ctx, cancel = chromedp.NewContext(allocCtx)
	defer cancel()
	if cfg.ShowLogs {
		if err := chromedp.Run(ctx, cdplog.Enable()); err != nil {
			return fmt.Errorf("failed to enable log domain: %v", err)
		}
		if err := chromedp.Run(ctx, runtime.Enable()); err != nil {
			return fmt.Errorf("failed to enable runtime domain: %v", err)
		}
	}
	if cfg.ShowNetwork {
		if err := chromedp.Run(ctx, cdpnetwork.Enable()); err != nil {
			return fmt.Errorf("failed to enable network domain: %v", err)
		}
	}
	requestMethods := sync.Map{}
	requestURLs := sync.Map{}
	requestStartTimes := sync.Map{}
	networkEntriesMap := sync.Map{}
	pageStartTime := time.Now()
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		if cfg.ShowNetwork {
			if ev, ok := ev.(*cdpnetwork.EventRequestWillBeSent); ok {
				if ev.Request != nil {
					requestMethods.Store(ev.RequestID.String(), ev.Request.Method)
					requestURLs.Store(ev.RequestID.String(), ev.Request.URL)
					requestStartTimes.Store(ev.RequestID.String(), float64(time.Since(pageStartTime).Nanoseconds())/1e6)
				}
			}
		}
		if cfg.ShowLogs {
			if ev, ok := ev.(*cdplog.EventEntryAdded); ok {
				onLog(LogEntry{
					Level:   ev.Entry.Level.String(),
					Message: ev.Entry.Text,
					Time:    time.Now(),
					Source:  "browser",
				})
			}
			if ev, ok := ev.(*runtime.EventConsoleAPICalled); ok {
				var message string
				for _, arg := range ev.Args {
					if arg.Value != nil {
						var str string
						if err := json.Unmarshal(arg.Value, &str); err == nil {
							message += str + " "
						} else {
							message += fmt.Sprintf("%v ", arg.Value)
						}
					}
				}
				onLog(LogEntry{
					Level:   ev.Type.String(),
					Message: strings.TrimSpace(message),
					Time:    time.Now(),
					Source:  "console",
				})
			}
		}
		if cfg.ShowNetwork {
			if ev, ok := ev.(*cdpnetwork.EventResponseReceived); ok {
				if !ShouldIncludeNetworkEvent(cfg, ev) {
					return
				}
				var method string
				if methodVal, ok := requestMethods.Load(ev.RequestID.String()); ok {
					if methodStr, ok := methodVal.(string); ok {
						method = methodStr
					}
				}
				var requestTiming *cdpnetwork.ResourceTiming
				var requestStartTime, responseTime float64
				if ev.Response != nil && ev.Response.Timing != nil {
					requestTiming = ev.Response.Timing
				}
				if timeVal, ok := requestStartTimes.Load(ev.RequestID.String()); ok {
					if timeFloat, ok := timeVal.(float64); ok {
						requestStartTime = timeFloat
					}
				}
				responseTime = float64(time.Since(pageStartTime).Nanoseconds()) / 1e6
				ne := BuildNetworkEntryFromEvent(ev, method, requestTiming, responseTime, requestStartTime)
				if ne.ResponseStartTime > 0 {
					responseStartTimesMap.Store(ev.RequestID.String(), responseTime)
				}
				onNet(ne)
				networkEntriesMap.Store(ev.RequestID.String(), ne)
			}
			if ev, ok := ev.(*cdpnetwork.EventLoadingFinished); ok {
				if entryVal, ok := networkEntriesMap.Load(ev.RequestID.String()); ok {
					if entry, ok := entryVal.(NetworkEntry); ok {
						loadingFinishedTime := float64(time.Since(pageStartTime).Nanoseconds()) / 1e6
						if responseStartTimeVal, ok := responseStartTimesMap.Load(ev.RequestID.String()); ok {
							if responseStartTime, ok := responseStartTimeVal.(float64); ok {
								entry.ContentDownloadTime = loadingFinishedTime - responseStartTime
								entry.ContentDownloadTimeFormatted = FormatTiming(entry.ContentDownloadTime)
								if entry.Duration > 0 && entry.ContentDownloadTime > 0 {
									entry.Duration = entry.Duration + entry.ContentDownloadTime
									entry.DurationFormatted = FormatTiming(entry.Duration)
								}
								entry.Total = entry.Total + entry.ContentDownloadTime
								entry.TotalFormatted = FormatTiming(entry.Total)
								networkEntriesMap.Store(ev.RequestID.String(), entry)
								onNet(entry)
							}
						}
					}
				}
			}
			// Network loading failures
			if ev, ok := ev.(*cdpnetwork.EventLoadingFailed); ok {
				ne := HandleLoadingFailedEvent(ev, &requestMethods, &requestURLs)
				if ne != nil {
					onNet(*ne)
				}
			}
		}
	})
	if len(cfg.Headers) > 0 || cfg.UserAgent != "" {
		if err := SetHeaders(ctx, cfg.UserAgent, cfg.Headers); err != nil {
			return fmt.Errorf("failed to set headers: %v", err)
		}
	}
	if len(cfg.Cookies) > 0 {
		if err := SetCookies(ctx, url, cfg.Cookies); err != nil {
			return fmt.Errorf("failed to set cookies: %v", err)
		}
	}
	if err := chromedp.Run(ctx, chromedp.Navigate(url)); err != nil {
		return fmt.Errorf("failed to navigate to %s: %v", url, err)
	}
	if cfg.RotateFingerprints {
		if err := StartFingerprintRotation(ctx, cfg.FingerprintInterval); err != nil {
			return fmt.Errorf("failed to start fingerprint rotation: %v", err)
		}
	}
	<-ctx.Done()
	return nil
}
