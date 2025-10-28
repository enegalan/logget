package helpers

import (
	"fmt"
	"strings"
)

func GenerateDynamicHeaders(cfg Config, url string) []string {
	customHeaderMap := make(map[string]string)
	for _, header := range cfg.Headers {
		if colonIndex := strings.Index(header, ":"); colonIndex != -1 {
			key := strings.TrimSpace(header[:colonIndex])
			value := strings.TrimSpace(header[colonIndex+1:])
			customHeaderMap[strings.ToLower(key)] = value
		}
	}
	var headers []string
	headers = append(headers, fmt.Sprintf("Host: %s", GetHostFromURL(url)))
	if customUA, exists := customHeaderMap["user-agent"]; exists {
		headers = append(headers, fmt.Sprintf("User-Agent: %s", customUA))
	} else {
		headers = append(headers, fmt.Sprintf("User-Agent: %s", cfg.UserAgent))
	}
	if customAccept, exists := customHeaderMap["accept"]; exists {
		headers = append(headers, fmt.Sprintf("Accept: %s", customAccept))
	} else {
		if strings.Contains(url, ".json") || strings.Contains(url, "/api/") || strings.Contains(url, "api.") {
			headers = append(headers, "Accept: application/json,text/plain,*/*")
		} else if strings.Contains(url, ".css") {
			headers = append(headers, "Accept: text/css,*/*;q=0.1")
		} else {
			headers = append(headers, "Accept: */*")
		}
	}
	if customLang, exists := customHeaderMap["accept-language"]; exists {
		headers = append(headers, fmt.Sprintf("Accept-Language: %s", customLang))
	} else {
		headers = append(headers, "Accept-Language: en-US,en;q=0.5")
	}
	if customEncoding, exists := customHeaderMap["accept-encoding"]; exists {
		headers = append(headers, fmt.Sprintf("Accept-Encoding: %s", customEncoding))
	} else {
		headers = append(headers, "Accept-Encoding: gzip, deflate")
	}
	if customConn, exists := customHeaderMap["connection"]; exists {
		headers = append(headers, fmt.Sprintf("Connection: %s", customConn))
	} else {
		headers = append(headers, "Connection: keep-alive")
	}
	if strings.HasPrefix(url, "https://") {
		if customUpgrade, exists := customHeaderMap["upgrade-insecure-requests"]; exists {
			headers = append(headers, fmt.Sprintf("Upgrade-Insecure-Requests: %s", customUpgrade))
		} else {
			headers = append(headers, "Upgrade-Insecure-Requests: 1")
		}
		if customDest, exists := customHeaderMap["sec-fetch-dest"]; exists {
			headers = append(headers, fmt.Sprintf("Sec-Fetch-Dest: %s", customDest))
		} else {
			headers = append(headers, "Sec-Fetch-Dest: document")
		}
		if customMode, exists := customHeaderMap["sec-fetch-mode"]; exists {
			headers = append(headers, fmt.Sprintf("Sec-Fetch-Mode: %s", customMode))
		} else {
			headers = append(headers, "Sec-Fetch-Mode: navigate")
		}
		if customSite, exists := customHeaderMap["sec-fetch-site"]; exists {
			headers = append(headers, fmt.Sprintf("Sec-Fetch-Site: %s", customSite))
		} else {
			headers = append(headers, "Sec-Fetch-Site: none")
		}
	}
	if customCache, exists := customHeaderMap["cache-control"]; exists {
		headers = append(headers, fmt.Sprintf("Cache-Control: %s", customCache))
	} else {
		headers = append(headers, "Cache-Control: max-age=0")
	}
	alreadyProcessedHeaders := map[string]bool{
		"user-agent":                true,
		"accept":                    true,
		"accept-language":           true,
		"accept-encoding":           true,
		"connection":                true,
		"upgrade-insecure-requests": true,
		"sec-fetch-dest":            true,
		"sec-fetch-mode":            true,
		"sec-fetch-site":            true,
		"cache-control":             true,
	}
	for _, header := range cfg.Headers {
		if colonIndex := strings.Index(header, ":"); colonIndex != -1 {
			key := strings.ToLower(strings.TrimSpace(header[:colonIndex]))
			if !alreadyProcessedHeaders[key] {
				headers = append(headers, header)
			}
		}
	}
	return headers
}
