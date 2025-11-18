package core

import (
	"fmt"
	"strings"

	chrome "logget/src/chrome"
)

type headerDef struct {
	name       string
	getDefault func() string
	httpsOnly  bool
}

func GenerateHeaders(userAgent string, headers []string, url string) []string {
	customHeaderMap := make(map[string]string)
	for _, header := range headers {
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
	getAcceptDefault := func() string {
		if strings.Contains(url, ".json") || strings.Contains(url, "/api/") || strings.Contains(url, "api.") {
			return "application/json,text/plain,*/*"
		}
		if strings.Contains(url, ".css") {
			return "text/css,*/*;q=0.1"
		}
		return "*/*"
	}
	defaultHeaders := []headerDef{
		{"User-Agent", func() string { return userAgent }, false},
		{"Accept", getAcceptDefault, false},
		{"Accept-Language", func() string { return "en-US,en;q=0.5" }, false},
		{"Accept-Encoding", func() string { return "gzip, deflate" }, false},
		{"Connection", func() string { return "keep-alive" }, false},
		{"Upgrade-Insecure-Requests", func() string { return "1" }, true},
		{"Sec-Fetch-Dest", func() string { return "document" }, true},
		{"Sec-Fetch-Mode", func() string { return "navigate" }, true},
		{"Sec-Fetch-Site", func() string { return "none" }, true},
		{"Cache-Control", func() string { return "max-age=0" }, false},
	}
	var result []string
	processedKeys := make(map[string]bool)
	for _, def := range defaultHeaders {
		if def.httpsOnly && !strings.HasPrefix(url, "https://") {
			continue
		}
		key := strings.ToLower(def.name)
		processedKeys[key] = true
		defaultValue := def.getDefault()
		result = append(result, fmt.Sprintf("%s: %s", def.name, getHeader(key, defaultValue)))
	}
	for _, header := range headers {
		if colonIndex := strings.Index(header, ":"); colonIndex != -1 {
			key := strings.ToLower(strings.TrimSpace(header[:colonIndex]))
			if !processedKeys[key] {
				result = append(result, header)
			}
		}
	}
	return result
}

func SetHeaders(ctx *chrome.ChromeContext, userAgent string, headers []string) error {
	headersList := make([]string, 0)
	hasUserAgent := false
	for _, header := range headers {
		if colonIndex := strings.Index(header, ":"); colonIndex != -1 {
			key := strings.TrimSpace(header[:colonIndex])
			value := strings.TrimSpace(header[colonIndex+1:])
			keyLower := strings.ToLower(key)
			if keyLower == "user-agent" {
				hasUserAgent = true
			}
			headersList = append(headersList, key, value)
		}
	}
	if !hasUserAgent && userAgent != "" {
		headersList = append(headersList, "User-Agent", userAgent)
	}
	if len(headersList) > 0 {
		_, err := ctx.Page.SetExtraHeaders(headersList)
		if err != nil {
			return fmt.Errorf("failed to set extra HTTP headers: %v", err)
		}
	}
	return nil
}
