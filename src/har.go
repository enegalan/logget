package helpers

import (
	"encoding/json"
	"fmt"
	"time"

	chrome "logget/src/chrome"
)

type HAR struct {
	Log HARLog `json:"log"`
}

type HARLog struct {
	Version string     `json:"version"`
	Creator HARCreator `json:"creator"`
	Pages   []HARPage  `json:"pages,omitempty"`
	Entries []HAREntry `json:"entries"`
}

type HARCreator struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type HARPage struct {
	StartedDateTime string         `json:"startedDateTime"`
	ID              string         `json:"id"`
	Title           string         `json:"title"`
	PageTimings     HARPageTimings `json:"pageTimings"`
}

type HARPageTimings struct {
	OnContentLoad float64 `json:"onContentLoad,omitempty"`
	OnLoad        float64 `json:"onLoad,omitempty"`
}

type HAREntry struct {
	Pageref         string      `json:"pageref,omitempty"`
	StartedDateTime string      `json:"startedDateTime"`
	Time            float64     `json:"time"`
	Request         HARRequest  `json:"request"`
	Response        HARResponse `json:"response"`
	Timings         HARTimings  `json:"timings"`
	ServerIPAddress string      `json:"serverIPAddress,omitempty"`
	Connection      string      `json:"connection,omitempty"`
}

type HARRequest struct {
	Method      string           `json:"method"`
	URL         string           `json:"url"`
	HTTPVersion string           `json:"httpVersion"`
	Headers     []HARHeader      `json:"headers"`
	QueryString []HARQueryString `json:"queryString"`
	Cookies     []HARCookie      `json:"cookies"`
	HeadersSize int              `json:"headersSize"`
	BodySize    int              `json:"bodySize"`
	PostData    *HARPostData     `json:"postData,omitempty"`
}

type HARResponse struct {
	Status      int         `json:"status"`
	StatusText  string      `json:"statusText"`
	HTTPVersion string      `json:"httpVersion"`
	Headers     []HARHeader `json:"headers"`
	Cookies     []HARCookie `json:"cookies"`
	Content     HARContent  `json:"content"`
	RedirectURL string      `json:"redirectURL,omitempty"`
	HeadersSize int         `json:"headersSize"`
	BodySize    int         `json:"bodySize"`
}

type HARContent struct {
	Size        int64  `json:"size"`
	MimeType    string `json:"mimeType"`
	Text        string `json:"text,omitempty"`
	Compression int64  `json:"compression,omitempty"`
}

type HARTimings struct {
	Blocked float64 `json:"blocked,omitempty"`
	DNS     float64 `json:"dns,omitempty"`
	Connect float64 `json:"connect,omitempty"`
	Send    float64 `json:"send"`
	Wait    float64 `json:"wait"`
	Receive float64 `json:"receive"`
	SSL     float64 `json:"ssl,omitempty"`
	Comment string  `json:"comment,omitempty"`
}

type HARHeader struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type HARQueryString struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type HARCookie struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	Path     string `json:"path,omitempty"`
	Domain   string `json:"domain,omitempty"`
	Expires  string `json:"expires,omitempty"`
	HTTPOnly bool   `json:"httpOnly,omitempty"`
	Secure   bool   `json:"secure,omitempty"`
	SameSite string `json:"sameSite,omitempty"`
}

type HARPostData struct {
	MimeType string         `json:"mimeType"`
	Params   []HARPostParam `json:"params,omitempty"`
	Text     string         `json:"text,omitempty"`
}

type HARPostParam struct {
	Name        string `json:"name"`
	Value       string `json:"value"`
	FileName    string `json:"fileName,omitempty"`
	ContentType string `json:"contentType,omitempty"`
}

func ConvertNetworkEntriesToHAR(entries []chrome.NetworkEntry, pageURL string, startTime time.Time) ([]byte, error) {
	har := HAR{
		Log: HARLog{
			Version: "1.2",
			Creator: HARCreator{
				Name:    "logget",
				Version: "2.0",
			},
			Entries: make([]HAREntry, 0, len(entries)),
		},
	}
	if len(entries) > 0 {
		pageDuration := entries[0].Duration
		if pageDuration < 0 {
			pageDuration = 0
		}
		har.Log.Pages = []HARPage{
			{
				StartedDateTime: startTime.Format(time.RFC3339),
				ID:              "page_1",
				Title:           pageURL,
				PageTimings: HARPageTimings{
					OnContentLoad: pageDuration,
					OnLoad:        pageDuration,
				},
			},
		}
	}
	for _, entry := range entries {
		headers := make([]HARHeader, 0, len(entry.Headers))
		for name, value := range entry.Headers {
			headers = append(headers, HARHeader{
				Name:  name,
				Value: value,
			})
		}
		timings := HARTimings{
			Blocked: 0,
			DNS:     entry.DNSLookupTime,
			Connect: entry.ConnectTime,
			SSL:     entry.SSLTime,
			Send:    entry.SendTime,
			Wait:    entry.WaitTime,
			Receive: entry.ReceiveTime,
		}
		if entry.SSLTime > 0 && entry.ConnectTime > entry.SSLTime {
			timings.Connect = entry.ConnectTime - entry.SSLTime
		}
		httpVersion := "HTTP/1.1"
		if entry.Headers["HTTP-Version"] != "" {
			httpVersion = entry.Headers["HTTP-Version"]
		}
		statusText := fmt.Sprintf("%d", entry.Status)
		if entry.Status == 0 {
			statusText = "Error"
		}
		harEntry := HAREntry{
			Pageref:         "page_1",
			StartedDateTime: entry.Timestamp.Format(time.RFC3339),
			Time:            entry.Duration,
			Request: HARRequest{
				Method:      entry.Method,
				URL:         entry.URL,
				HTTPVersion: httpVersion,
				Headers:     headers,
				QueryString: []HARQueryString{},
				Cookies:     []HARCookie{},
				HeadersSize: -1,
				BodySize:    -1,
			},
			Response: HARResponse{
				Status:      entry.Status,
				StatusText:  statusText,
				HTTPVersion: httpVersion,
				Headers:     headers,
				Cookies:     []HARCookie{},
				Content: HARContent{
					Size:     entry.Size,
					MimeType: entry.Type,
				},
				HeadersSize: -1,
				BodySize:    int(entry.Size),
			},
			Timings: timings,
		}
		har.Log.Entries = append(har.Log.Entries, harEntry)
	}
	harJSON, err := json.MarshalIndent(har, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal HAR: %v", err)
	}
	return harJSON, nil
}
