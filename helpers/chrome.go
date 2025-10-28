package helpers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
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
	URL       string            `json:"url"`
	Method    string            `json:"method"`
	Status    int               `json:"status"`
	Headers   map[string]string `json:"headers"`
	Timestamp time.Time         `json:"timestamp"`
	Type      string            `json:"type"`
	Size      int64             `json:"size"`
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
	if err := chromedp.Run(ctx, cdplog.Enable()); err != nil {
		return fmt.Errorf("failed to enable log domain: %v", err)
	}
	if err := chromedp.Run(ctx, runtime.Enable()); err != nil {
		return fmt.Errorf("failed to enable runtime domain: %v", err)
	}
	if cfg.ShowNetwork {
		if err := chromedp.Run(ctx, cdpnetwork.Enable()); err != nil {
			return fmt.Errorf("failed to enable network domain: %v", err)
		}
	}
	chromedp.ListenTarget(ctx, func(ev interface{}) {
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
		if cfg.ShowNetwork {
			if ev, ok := ev.(*cdpnetwork.EventResponseReceived); ok {
				response := ev.Response
				headers := make(map[string]string)
				for name, value := range response.Headers {
					if str, ok := value.(string); ok {
						headers[name] = str
					} else {
						headers[name] = fmt.Sprintf("%v", value)
					}
				}
				onNet(NetworkEntry{
					URL:       response.URL,
					Method:    "GET",
					Status:    int(response.Status),
					Headers:   headers,
					Timestamp: time.Now(),
					Type:      string(response.MimeType),
					Size:      int64(response.EncodedDataLength),
				})
			}
		}
	})
	if len(cfg.Cookies) > 0 {
		if err := SetCookies(ctx, url, cfg.Cookies); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to set cookies: %v\n", err)
		}
	}
	if err := chromedp.Run(ctx, chromedp.Navigate(url)); err != nil {
		return fmt.Errorf("failed to navigate to %s: %v", url, err)
	}
	<-ctx.Done()
	return nil
}
