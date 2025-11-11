package helpers

import (
	"context"
	"fmt"
	"strings"

	cdpnetwork "github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

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
	var headers []string
	headers = append(headers, fmt.Sprintf("User-Agent: %s", getHeader("user-agent", cfg.UserAgent)))
	acceptDefault := "*/*"
	if strings.Contains(url, ".json") || strings.Contains(url, "/api/") || strings.Contains(url, "api.") {
		acceptDefault = "application/json,text/plain,*/*"
	} else if strings.Contains(url, ".css") {
		acceptDefault = "text/css,*/*;q=0.1"
	}
	headers = append(headers, fmt.Sprintf("Accept: %s", getHeader("accept", acceptDefault)))
	headers = append(headers, fmt.Sprintf("Accept-Language: %s", getHeader("accept-language", "en-US,en;q=0.5")))
	headers = append(headers, fmt.Sprintf("Accept-Encoding: %s", getHeader("accept-encoding", "gzip, deflate")))
	headers = append(headers, fmt.Sprintf("Connection: %s", getHeader("connection", "keep-alive")))
	if strings.HasPrefix(url, "https://") {
		headers = append(headers, fmt.Sprintf("Upgrade-Insecure-Requests: %s", getHeader("upgrade-insecure-requests", "1")))
		headers = append(headers, fmt.Sprintf("Sec-Fetch-Dest: %s", getHeader("sec-fetch-dest", "document")))
		headers = append(headers, fmt.Sprintf("Sec-Fetch-Mode: %s", getHeader("sec-fetch-mode", "navigate")))
		headers = append(headers, fmt.Sprintf("Sec-Fetch-Site: %s", getHeader("sec-fetch-site", "none")))
	}
	headers = append(headers, fmt.Sprintf("Cache-Control: %s", getHeader("cache-control", "max-age=0")))
	processedKeys := map[string]bool{
		"user-agent": true, "accept": true, "accept-language": true, "accept-encoding": true,
		"connection": true, "upgrade-insecure-requests": true, "sec-fetch-dest": true,
		"sec-fetch-mode": true, "sec-fetch-site": true, "cache-control": true,
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
