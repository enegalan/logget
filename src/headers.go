package helpers

import (
	"context"
	"fmt"
	"strings"

	cdpnetwork "github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

type headerDef struct {
	name         string
	getDefault   func(cfg Config, url string) string
	httpsOnly    bool
}

func GenerateHeaders(cfg Config, url string) []string {
	customHeaderMap := make(map[string]string)
	for _, header := range cfg.Headers {
		if colonIndex := strings.Index(header, ":"); colonIndex != -1 {
			key := strings.TrimSpace(header[:colonIndex])
			value := strings.TrimSpace(header[colonIndex+1:])
			customHeaderMap[strings.ToLower(key)] = value
		}
	}
	getHeader := func(key, defaultValue string) string {
		if val, exists := customHeaderMap[strings.ToLower(key)]; exists {
			return val
		}
		return defaultValue
	}
	getAcceptDefault := func(url string) string {
		if strings.Contains(url, ".json") || strings.Contains(url, "/api/") || strings.Contains(url, "api.") {
			return "application/json,text/plain,*/*"
		}
		if strings.Contains(url, ".css") {
			return "text/css,*/*;q=0.1"
		}
		return "*/*"
	}
	defaultHeaders := []headerDef{
		{"User-Agent", func(cfg Config, url string) string { return cfg.UserAgent }, false},
		{"Accept", func(cfg Config, url string) string { return getAcceptDefault(url) }, false},
		{"Accept-Language", func(cfg Config, url string) string { return "en-US,en;q=0.5" }, false},
		{"Accept-Encoding", func(cfg Config, url string) string { return "gzip, deflate" }, false},
		{"Connection", func(cfg Config, url string) string { return "keep-alive" }, false},
		{"Upgrade-Insecure-Requests", func(cfg Config, url string) string { return "1" }, true},
		{"Sec-Fetch-Dest", func(cfg Config, url string) string { return "document" }, true},
		{"Sec-Fetch-Mode", func(cfg Config, url string) string { return "navigate" }, true},
		{"Sec-Fetch-Site", func(cfg Config, url string) string { return "none" }, true},
		{"Cache-Control", func(cfg Config, url string) string { return "max-age=0" }, false},
	}
	var headers []string
	processedKeys := make(map[string]bool)
	for _, def := range defaultHeaders {
		if def.httpsOnly && !strings.HasPrefix(url, "https://") {
			continue
		}
		key := strings.ToLower(def.name)
		processedKeys[key] = true
		defaultValue := def.getDefault(cfg, url)
		headers = append(headers, fmt.Sprintf("%s: %s", def.name, getHeader(key, defaultValue)))
	}
	for _, header := range cfg.Headers {
		if colonIndex := strings.Index(header, ":"); colonIndex != -1 {
			key := strings.ToLower(strings.TrimSpace(header[:colonIndex]))
			if !processedKeys[key] {
				headers = append(headers, header)
			}
		}
	}
	return headers
}

func SetHeaders(ctx context.Context, userAgent string, headers []string) error {
	headersMap := make(cdpnetwork.Headers)
	hasUserAgent := false
	for _, header := range headers {
		if colonIndex := strings.Index(header, ":"); colonIndex != -1 {
			key := strings.TrimSpace(header[:colonIndex])
			value := strings.TrimSpace(header[colonIndex+1:])
			keyLower := strings.ToLower(key)
			if keyLower == "user-agent" {
				hasUserAgent = true
			}
			headersMap[key] = value
		}
	}
	if !hasUserAgent && userAgent != "" {
		headersMap["User-Agent"] = userAgent
	}
	if len(headersMap) > 0 {
		if err := chromedp.Run(ctx, cdpnetwork.SetExtraHTTPHeaders(headersMap)); err != nil {
			return fmt.Errorf("failed to set extra HTTP headers: %v", err)
		}
	}
	return nil
}
