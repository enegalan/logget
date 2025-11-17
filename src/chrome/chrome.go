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

	cdpnetwork "github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
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
}

func StreamLogsRealTime(cfg StreamConfig, ctx context.Context, url string, onLog func(LogEntry), onNet func(NetworkEntry), setHeaders func(context.Context, string, []string) error, setCookies func(context.Context, string, []string) error, startFingerprintRotation func(context.Context, int) error) error {
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
				ProcessNetworkEventResponseReceived(evResp, streamCfg, &maps.Methods, &maps.StartTimes, pageStartTime, &maps.NetworkEntries, handlers)
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
		if err := setHeaders(chromeCtx, cfg.UserAgent, cfg.Headers); err != nil {
			return fmt.Errorf("failed to set headers: %v", err)
		}
	}
	if len(cfg.Cookies) > 0 {
		if err := setCookies(chromeCtx, url, cfg.Cookies); err != nil {
			return fmt.Errorf("failed to set cookies: %v", err)
		}
	}
	if err := chromedp.Run(chromeCtx, chromedp.Navigate(url)); err != nil {
		return fmt.Errorf("failed to navigate to %s: %v", url, err)
	}
	if cfg.RotateFingerprints {
		if err := startFingerprintRotation(chromeCtx, cfg.FingerprintInterval); err != nil {
			return fmt.Errorf("failed to start fingerprint rotation: %v", err)
		}
	}
	<-chromeCtx.Done()
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
