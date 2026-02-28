package command

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"
	"time"

	chrome "logget/src/chrome"
	"logget/src/colors"
	"logget/src/io"
)

const separatorWidth = 60

var (
	logLevelSymbols = map[string]string{
		"DEBUG":               "🐛",
		"INFO":                "ℹ️",
		"WARN":                "⚠️",
		"WARNING":             "⚠️",
		"ERROR":               "❌",
		"FATAL":               "💀",
		"LOG":                 "📝",
		"TRACE":               "🔍",
		"STARTGROUP":          "📂",
		"STARTGROUPCOLLAPSED": "📂",
		"ENDGROUP":            "📂",
		"DIR":                 "🗂️",
		"DIRXML":              "🗂️",
		"TABLE":               "📊",
		"TIMEEND":             "⏰",
	}
	defaultLogLevelSymbol = "📋"

	httpMethodSymbols = map[string]string{
		"GET":     "📥",
		"POST":    "📤",
		"HEAD":    "🔍",
		"OPTIONS": "🔍",
		"PUT":     "🔄",
		"DELETE":  "🗑️",
		"PATCH":   "🔧",
	}
	defaultHTTPMethodSymbol = "🌐"
)

type OutputFormatter struct{ theme *colors.ColorTheme }

func NewOutputFormatter(enableColors bool) *OutputFormatter {
	return &OutputFormatter{theme: colors.GetTheme(enableColors)}
}

func (f *OutputFormatter) FormatHTTPResponse(protocol string, statusCode int, duration time.Duration) string {
	sb := strings.Builder{}
	sb.Grow(128)
	sb.WriteString(f.theme.Bold(protocol))
	sb.WriteString(" ")
	sb.WriteString(f.theme.Colorize(f.theme.GetStatusColor(statusCode), strconv.Itoa(statusCode)))
	sb.WriteString("\n")
	sb.WriteString(f.theme.Colorize(f.theme.Timestamp, fmt.Sprintf("Duration: %v", duration)))
	sb.WriteString("\n")
	return sb.String()
}

func (f *OutputFormatter) FormatRequestHeaders(headers []string) string {
	sb := strings.Builder{}
	sb.Grow(len(headers) * 64)
	sb.WriteString("\n=== ")
	sb.WriteString(f.theme.FormatHeader("REQUEST HEADERS"))
	sb.WriteString(" ===\n")
	for _, header := range headers {
		sb.WriteString(f.theme.Colorize(colors.Yellow, header))
		sb.WriteString("\n")
	}
	return sb.String()
}

func (f *OutputFormatter) FormatResponseHeaders(headers map[string]string) string {
	if len(headers) == 0 {
		return ""
	}
	sb := strings.Builder{}
	sb.Grow(len(headers) * 64)
	sb.WriteString("\n=== ")
	sb.WriteString(f.theme.FormatHeader("RESPONSE HEADERS"))
	sb.WriteString(" ===\n")
	for name, value := range headers {
		sb.WriteString(f.theme.Bold(name))
		sb.WriteString(": ")
		sb.WriteString(value)
		sb.WriteString("\n")
	}
	return sb.String()
}

func (f *OutputFormatter) FormatConsoleLogs(logs []chrome.LogEntry) string {
	if len(logs) == 0 {
		return ""
	}
	sb := strings.Builder{}
	sb.Grow(len(logs) * 128)
	sb.WriteString("\n=== ")
	sb.WriteString(f.theme.FormatHeader("CONSOLE LOGS"))
	sb.WriteString(" ===\n")
	for _, log := range logs {
		sb.WriteString("[")
		sb.WriteString(f.theme.FormatTimestamp(log.Time.Format("15:04:05")))
		sb.WriteString("] ")
		sb.WriteString(f.theme.FormatLogLevel(strings.ToUpper(log.Level)))
		sb.WriteString(": ")
		sb.WriteString(log.Message)
		sb.WriteString("\n")
	}
	return sb.String()
}

func (f *OutputFormatter) FormatNetworkRequests(network []chrome.NetworkEntry) string {
	if len(network) == 0 {
		return ""
	}
	sb := strings.Builder{}
	sb.Grow(len(network) * 256)
	sb.WriteString("\n=== ")
	sb.WriteString(f.theme.FormatHeader("NETWORK REQUESTS"))
	sb.WriteString(" ===\n")
	for _, net := range network {
		sb.WriteString(f.theme.Bold(net.Method))
		sb.WriteString(" ")
		sb.WriteString(f.theme.Colorize(colors.Cyan, net.URL))
		sb.WriteString(" -> ")
		if net.Error != "" {
			errorColor := colors.Red
			switch net.ErrorType {
			case "timeout":
				errorColor = colors.Red
			case "cors":
				errorColor = colors.Yellow
			case "dns":
				errorColor = colors.Magenta
			}
			sb.WriteString(f.theme.Colorize(errorColor, fmt.Sprintf("ERROR: %s (%s)", net.Error, net.ErrorType)))
		} else {
			sb.WriteString(f.theme.Colorize(f.theme.GetStatusColor(net.Status), strconv.Itoa(net.Status)))
		}
		sb.WriteString("\n")
		for k, v := range net.Headers {
			sb.WriteString("  ")
			sb.WriteString(f.theme.Colorize(colors.Yellow, k))
			sb.WriteString(": ")
			sb.WriteString(v)
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

func (f *OutputFormatter) FormatSummary(logCount, networkCount int, duration time.Duration) string {
	var sb strings.Builder
	sb.WriteString("\n=== ")
	sb.WriteString(f.theme.FormatHeader("SUMMARY"))
	sb.WriteString(" ===\n")
	sb.WriteString(f.theme.Colorize(f.theme.Timestamp, fmt.Sprintf("Total Duration: %v", duration)))
	sb.WriteString("\n")
	sb.WriteString(f.theme.Colorize(f.theme.Timestamp, fmt.Sprintf("Console Logs: %d", logCount)))
	sb.WriteString("\n")
	sb.WriteString(f.theme.Colorize(f.theme.Timestamp, fmt.Sprintf("Network Requests: %d", networkCount)))
	sb.WriteString("\n")
	return sb.String()
}

func (f *OutputFormatter) FormatSeparator() string { return f.theme.FormatSeparator("─", separatorWidth) + "\n" }

func (f *OutputFormatter) FormatSuccessMessage(message string) string {
	return f.theme.FormatSuccess(message) + "\n"
}

func (f *OutputFormatter) FormatErrorMessage(message string) string {
	return f.theme.FormatError(message) + "\n"
}

func (f *OutputFormatter) GetLogLevelColor(level string) string {
	return f.theme.GetLogLevelColor(level)
}

func (f *OutputFormatter) GetHTTPMethodColor(method string) string {
	return f.theme.GetHTTPMethodColor(method)
}

func (f *OutputFormatter) GetStatusColor(statusCode int) string {
	return f.theme.GetStatusColor(statusCode)
}

func (f *OutputFormatter) FormatTimestamp(timestamp string) string {
	return f.theme.FormatTimestamp(timestamp)
}

func (f *OutputFormatter) FormatConsolePrefix() string { return f.theme.FormatConsolePrefix() }

func (f *OutputFormatter) FormatNetworkPrefix() string { return f.theme.FormatNetworkPrefix() }

func (f *OutputFormatter) Colorize(color, text string) string { return f.theme.Colorize(color, text) }

func (f *OutputFormatter) FormatAndOutputLog(le chrome.LogEntry, cfg Config) error {
	level := strings.ToUpper(le.Level)
	levelColor := f.GetLogLevelColor(level)
	levelSymbol := logLevelSymbols[level]
	if levelSymbol == "" {
		levelSymbol = defaultLogLevelSymbol
	}
	return io.WriteOutput(io.WriteConfig{
		OutputWriter: cfg.OutputWriter,
		OutputFile:   cfg.OutputFile,
		AppendMode:   cfg.AppendMode,
		FollowMode:   cfg.FollowMode,
	}, fmt.Sprintf("[%s] %s %s %s: %s\n",
		f.FormatTimestamp(le.Time.Format("15:04:05")), f.FormatConsolePrefix(),
		f.Colorize(levelColor, levelSymbol), f.Colorize(levelColor, level), le.Message))
}

func (f *OutputFormatter) FormatAndOutputNetwork(ne chrome.NetworkEntry, cfg Config) error {
	methodColor := f.GetHTTPMethodColor(ne.Method)
	methodSymbol := httpMethodSymbols[ne.Method]
	if methodSymbol == "" {
		methodSymbol = defaultHTTPMethodSymbol
	}
	var statusOrError string
	if ne.Error != "" {
		errorColor := colors.Red
		switch ne.ErrorType {
		case "timeout":
			errorColor = colors.Red
		case "cors":
			errorColor = colors.Yellow
		case "dns":
			errorColor = colors.Magenta
		}
		statusOrError = f.Colorize(errorColor, fmt.Sprintf("ERROR: %s (%s)", ne.Error, ne.ErrorType))
	} else {
		statusOrError = f.Colorize(f.GetStatusColor(ne.Status), strconv.Itoa(ne.Status))
	}
	return io.WriteOutput(io.WriteConfig{
		OutputWriter: cfg.OutputWriter,
		OutputFile:   cfg.OutputFile,
		AppendMode:   cfg.AppendMode,
		FollowMode:   cfg.FollowMode,
	}, fmt.Sprintf("[%s] %s %s %s %s %s\n",
		f.FormatTimestamp(ne.Timestamp.Format("15:04:05")), f.FormatNetworkPrefix(),
		f.Colorize(methodColor, methodSymbol), f.Colorize(methodColor, ne.Method), ne.URL, statusOrError))
}

func (f *OutputFormatter) FormatLogsCSV(logs []chrome.LogEntry, includeHeader bool) string {
	if len(logs) == 0 {
		return ""
	}
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	if includeHeader {
		w.Write([]string{"timestamp", "level", "source", "message"})
	}
	for _, le := range logs {
		w.Write([]string{le.Time.Format(time.RFC3339), strings.ToUpper(le.Level), le.Source, le.Message})
	}
	w.Flush()
	return buf.String()
}

func (f *OutputFormatter) FormatNetworkCSV(entries []chrome.NetworkEntry, includeHeader bool) string {
	if len(entries) == 0 {
		return ""
	}
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	if includeHeader {
		w.Write([]string{"timestamp", "method", "url", "status", "resourceType", "mimeType", "size", "duration", "ttfb", "connectTime", "dnsTime", "sslTime", "sendTime", "waitTime", "receiveTime", "error", "errorType"})
	}
	for _, ne := range entries {
		_ = w.Write([]string{
			ne.Timestamp.Format(time.RFC3339),
			ne.Method,
			ne.URL,
			strconv.Itoa(ne.Status),
			ne.ResourceType,
			ne.Type,
			strconv.FormatInt(ne.Size, 10),
			strconv.FormatFloat(ne.Duration, 'f', 2, 64),
			strconv.FormatFloat(ne.TimeToFirstByte, 'f', 2, 64),
			strconv.FormatFloat(ne.ConnectTime, 'f', 2, 64),
			strconv.FormatFloat(ne.DNSLookupTime, 'f', 2, 64),
			strconv.FormatFloat(ne.SSLTime, 'f', 2, 64),
			strconv.FormatFloat(ne.SendTime, 'f', 2, 64),
			strconv.FormatFloat(ne.WaitTime, 'f', 2, 64),
			strconv.FormatFloat(ne.ReceiveTime, 'f', 2, 64),
			ne.Error,
			ne.ErrorType,
		})
	}
	w.Flush()
	return buf.String()
}

func (f *OutputFormatter) formatAndOutputCSVRow(header []string, row []string, includeHeader bool, cfg Config) error {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	if includeHeader {
		w.Write(header)
	}
	w.Write(row)
	w.Flush()
	return io.WriteOutput(io.WriteConfig{
		OutputWriter: cfg.OutputWriter,
		OutputFile:   cfg.OutputFile,
		AppendMode:   cfg.AppendMode,
		FollowMode:   cfg.FollowMode,
	}, buf.String())
}

func (f *OutputFormatter) FormatAndOutputLogCSVRow(le chrome.LogEntry, cfg Config, includeHeader bool) error {
	return f.formatAndOutputCSVRow([]string{"timestamp", "level", "source", "message"},
		[]string{le.Time.Format(time.RFC3339), strings.ToUpper(le.Level), le.Source, le.Message}, includeHeader, cfg)
}

func (f *OutputFormatter) FormatAndOutputNetworkCSVRow(ne chrome.NetworkEntry, cfg Config, includeHeader bool) error {
	return f.formatAndOutputCSVRow([]string{"timestamp", "method", "url", "status", "resourceType", "mimeType", "size", "duration", "ttfb", "connectTime", "dnsTime", "sslTime", "sendTime", "waitTime", "receiveTime", "error", "errorType"},
		[]string{ne.Timestamp.Format(time.RFC3339), ne.Method, ne.URL, strconv.Itoa(ne.Status), ne.ResourceType, ne.Type,
			strconv.FormatInt(ne.Size, 10), strconv.FormatFloat(ne.Duration, 'f', 2, 64), strconv.FormatFloat(ne.TimeToFirstByte, 'f', 2, 64),
			strconv.FormatFloat(ne.ConnectTime, 'f', 2, 64), strconv.FormatFloat(ne.DNSLookupTime, 'f', 2, 64), strconv.FormatFloat(ne.SSLTime, 'f', 2, 64),
			strconv.FormatFloat(ne.SendTime, 'f', 2, 64), strconv.FormatFloat(ne.WaitTime, 'f', 2, 64), strconv.FormatFloat(ne.ReceiveTime, 'f', 2, 64),
			ne.Error, ne.ErrorType}, includeHeader, cfg)
}
