package helpers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
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
	if cfg.WasmOnly {
		if ev.Response == nil || strings.ToLower(string(ev.Response.MimeType)) != "application/wasm" {
			return false
		}
	}
	return true
}

func BuildNetworkEntryFromEvent(ev *cdpnetwork.EventResponseReceived) NetworkEntry {
	response := ev.Response
	headers := make(map[string]string)
	for name, value := range response.Headers {
		if str, ok := value.(string); ok {
			headers[name] = str
		} else {
			headers[name] = fmt.Sprintf("%v", value)
		}
	}
	return NetworkEntry{
		URL:          response.URL,
		Method:       "GET",
		Status:       int(response.Status),
		Headers:      headers,
		Timestamp:    time.Now(),
		Type:         string(response.MimeType),
		Size:         int64(response.EncodedDataLength),
		ResourceType: ev.Type.String(),
	}
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
	chromedp.ListenTarget(ctx, func(ev interface{}) {
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
				onNet(BuildNetworkEntryFromEvent(ev))
			}
		}
	})
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
