package chrome

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/go-rod/rod/lib/proto"
)

type NetworkMaps struct {
	Methods        sync.Map
	URLs           sync.Map
	StartTimes     sync.Map
	NetworkEntries sync.Map
}

func GetNetworkMaps() *NetworkMaps {
	return &NetworkMaps{}
}

type StreamConfig struct {
	SkipSSLVerify       bool
	ShowLogs            bool
	ShowNetwork         bool
	Headers             []string
	Cookies             []string
	UserAgent           string
	RotateFingerprints  bool
	FingerprintInterval int
	XHROnly             bool
	DocumentOnly        bool
	CssOnly             bool
	ScriptOnly          bool
	FontOnly            bool
	ImgOnly             bool
	MediaOnly           bool
	ManifestOnly        bool
	WebSocketOnly       bool
	MimeRegex           *regexp.Regexp
	StatusRegex         *regexp.Regexp
	DomainRegex         *regexp.Regexp
	MinSize             int64
	MaxSize             int64
	ExecuteJS           string
}

func StreamLogsRealTime(cfg StreamConfig, ctx context.Context, url string, onLog func(LogEntry), onNet func(NetworkEntry), setHeaders func(*ChromeContext, string, []string) error, setCookies func(*ChromeContext, string, []string) error, startFingerprintRotation func(*ChromeContext, int) error, executeJavaScript func(*ChromeContext, string) (interface{}, error), onJavaScriptResult func(interface{}, error)) error {
	chromeCtx, cancel, err := CreateChromeContext(ctx, cfg.SkipSSLVerify, false)
	if err != nil {
		return err
	}
	defer cancel()
	maps := GetNetworkMaps()
	pageStartTime := time.Now()
	handlers := &EventHandlers{
		OnLog:     onLog,
		OnNetwork: onNet,
	}
	streamCfg := StreamNetworkConfig{
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
	go chromeCtx.Page.EachEvent(func(ev *proto.NetworkRequestWillBeSent) {
		if cfg.ShowNetwork {
			ProcessNetworkEventRequestWillBeSent(ev, &maps.Methods, &maps.URLs, &maps.StartTimes, pageStartTime, handlers)
		}
	}, func(ev *proto.NetworkResponseReceived) {
		if cfg.ShowNetwork {
			ProcessNetworkEventResponseReceived(ev, streamCfg, &maps.Methods, &maps.StartTimes, pageStartTime, &maps.NetworkEntries, handlers)
		}
	}, func(ev *proto.NetworkLoadingFinished) {
		if cfg.ShowNetwork {
			ProcessNetworkEventLoadingFinished(ev, &maps.NetworkEntries, pageStartTime, handlers)
		}
	}, func(ev *proto.NetworkLoadingFailed) {
		if cfg.ShowNetwork {
			ProcessNetworkEventLoadingFailed(ev, &maps.Methods, &maps.URLs, handlers)
		}
	}, func(ev *proto.LogEntryAdded) {
		if cfg.ShowLogs {
			ProcessLogEvent(ev, handlers)
		}
	}, func(ev *proto.RuntimeConsoleAPICalled) {
		if cfg.ShowLogs {
			ProcessLogEvent(ev, handlers)
		}
	}, func(ev *proto.RuntimeExceptionThrown) {
		if cfg.ShowLogs {
			ProcessLogEvent(ev, handlers)
		}
	})()
	if len(cfg.Headers) > 0 || cfg.UserAgent != "" {
		if err := setHeaders(chromeCtx, cfg.UserAgent, cfg.Headers); err != nil {
			return fmt.Errorf("failed to set headers: %v", err)
		}
	}
	if len(cfg.Cookies) > 0 {
		if err := setCookies(chromeCtx, url, cfg.Cookies); err != nil {
			return fmt.Errorf("failed to set cookies: %v", err)
		}
	}
	if err := chromeCtx.Page.Navigate(url); err != nil {
		return fmt.Errorf("failed to navigate to %s: %v", url, err)
	}
	chromeCtx.Page.MustWaitLoad()
	if cfg.RotateFingerprints {
		if err := startFingerprintRotation(chromeCtx, cfg.FingerprintInterval); err != nil {
			return fmt.Errorf("failed to start fingerprint rotation: %v", err)
		}
	}
	if cfg.ExecuteJS != "" && executeJavaScript != nil {
		result, err := executeJavaScript(chromeCtx, cfg.ExecuteJS)
		if onJavaScriptResult != nil {
			onJavaScriptResult(result, err)
		} else {
			if err != nil {
				handlers.OnLog(createLogEntry("error", err.Error(), "javascript"))
			} else if result != nil {
				handlers.OnLog(createLogEntry("info", fmt.Sprintf("JavaScript result: %v", result), "javascript"))
			} else {
				handlers.OnLog(createLogEntry("info", "JavaScript executed successfully", "javascript"))
			}
		}
	}
	<-ctx.Done()
	return nil
}

func GetInitialResponse(skipSSLVerify bool, userAgent string, headers []string, targetURL string) (string, int, error) {
	transport := &http.Transport{}
	if skipSSLVerify {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	client := &http.Client{
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Timeout: 5 * time.Second,
	}
	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return "", 0, err
	}
	req.Header.Set("User-Agent", userAgent)
	for _, header := range headers {
		if colonIndex := strings.Index(header, ":"); colonIndex != -1 {
			key := strings.TrimSpace(header[:colonIndex])
			value := strings.TrimSpace(header[colonIndex+1:])
			req.Header.Set(key, value)
		}
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()
	return resp.Proto, resp.StatusCode, nil
}
