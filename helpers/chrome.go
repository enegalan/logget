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
	URL          string            `json:"url"`
	Method       string            `json:"method"`
	Status       int               `json:"status"`
	Headers      map[string]string `json:"headers"`
	Timestamp    time.Time         `json:"timestamp"`
	Type         string            `json:"type"`
	Size         int64             `json:"size"`
	ResourceType string            `json:"resourceType"`
	Error        string            `json:"error,omitempty"`
	ErrorType    string            `json:"errorType,omitempty"`
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
	return true
}

func BuildNetworkEntryFromEvent(ev *cdpnetwork.EventResponseReceived, requestMethod string) NetworkEntry {
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
	return NetworkEntry{
		URL:          response.URL,
		Method:       method,
		Status:       int(response.Status),
		Headers:      headers,
		Timestamp:    time.Now(),
		Type:         string(response.MimeType),
		Size:         int64(response.EncodedDataLength),
		ResourceType: ev.Type.String(),
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
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		if cfg.ShowNetwork {
			if ev, ok := ev.(*cdpnetwork.EventRequestWillBeSent); ok {
				if ev.Request != nil {
					requestMethods.Store(ev.RequestID.String(), ev.Request.Method)
					requestURLs.Store(ev.RequestID.String(), ev.Request.URL)
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
				onNet(BuildNetworkEntryFromEvent(ev, method))
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
	<-ctx.Done()
	return nil
}
