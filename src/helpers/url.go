package helpers

import "strings"

func NormalizeURL(url string) string {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return "https://" + url
	}
	return url
}

func GetHostFromURL(rawURL string) string {
	if strings.HasPrefix(rawURL, "http://") {
		rawURL = rawURL[7:]
	} else if strings.HasPrefix(rawURL, "https://") {
		rawURL = rawURL[8:]
	}
	if idx := strings.Index(rawURL, "/"); idx != -1 {
		return rawURL[:idx]
	}
	return rawURL
}
