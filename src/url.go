package helpers

import (
	"crypto/tls"
	"net/http"
	"strings"
	"time"
)

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

func GetInitialResponse(cfg Config, targetURL string) (string, int, error) {
	transport := &http.Transport{}
	if cfg.SkipSSLVerify {
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
	req.Header.Set("User-Agent", cfg.UserAgent)
	for _, header := range cfg.Headers {
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
