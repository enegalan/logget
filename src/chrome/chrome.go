package chrome

import (
	"context"
	"fmt"
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
					ShowNetwork: cfg.ShowNetwork,
					ShowLogs:    cfg.ShowLogs,
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
