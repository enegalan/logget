package helpers

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type OutputFormatter struct {
	theme *ColorTheme
}

func NewOutputFormatter(colors bool) *OutputFormatter {
	return &OutputFormatter{
		theme: GetTheme(colors),
	}
}

func (f *OutputFormatter) FormatHTTPResponse(protocol string, statusCode int, duration time.Duration) string {
	sb := strings.Builder{}
	sb.Grow(128)
	statusColor := f.theme.GetStatusColor(statusCode)
	formattedProtocol := f.theme.Bold(protocol)
	formattedStatusCode := f.theme.Colorize(statusColor, strconv.Itoa(statusCode))
	sb.WriteString(formattedProtocol)
	sb.WriteString(" ")
	sb.WriteString(formattedStatusCode)
	sb.WriteString("\n")
	durationText := f.theme.Colorize(f.theme.Timestamp, fmt.Sprintf("Duration: %v", duration))
	sb.WriteString(durationText)
	sb.WriteString("\n")
	return sb.String()
}

func (f *OutputFormatter) FormatRequestHeaders(headers []string) string {
	sb := strings.Builder{}
	sb.Grow(len(headers) * 64)
	sectionTitle := f.theme.FormatHeader("REQUEST HEADERS")
	sb.WriteString("\n=== ")
	sb.WriteString(sectionTitle)
	sb.WriteString(" ===\n")
	for _, header := range headers {
		formattedHeader := f.theme.Colorize(Yellow, header)
		sb.WriteString(formattedHeader)
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
	sectionTitle := f.theme.FormatHeader("RESPONSE HEADERS")
	sb.WriteString("\n=== ")
	sb.WriteString(sectionTitle)
	sb.WriteString(" ===\n")
	for name, value := range headers {
		formattedName := f.theme.Bold(name)
		sb.WriteString(formattedName)
		sb.WriteString(": ")
		sb.WriteString(value)
		sb.WriteString("\n")
	}
	return sb.String()
}

func (f *OutputFormatter) FormatConsoleLogs(logs []LogEntry) string {
	if len(logs) == 0 {
		return ""
	}
	sb := strings.Builder{}
	sb.Grow(len(logs) * 128)
	sectionTitle := f.theme.FormatHeader("CONSOLE LOGS")
	sb.WriteString("\n=== ")
	sb.WriteString(sectionTitle)
	sb.WriteString(" ===\n")
	for _, log := range logs {
		timestamp := log.Time.Format("15:04:05")
		level := strings.ToUpper(log.Level)
		formattedTimestamp := f.theme.FormatTimestamp(timestamp)
		formattedLevel := f.theme.FormatLogLevel(level)
		sb.WriteString("[")
		sb.WriteString(formattedTimestamp)
		sb.WriteString("] ")
		sb.WriteString(formattedLevel)
		sb.WriteString(": ")
		sb.WriteString(log.Message)
		sb.WriteString("\n")
	}
	return sb.String()
}

func (f *OutputFormatter) FormatNetworkRequests(network []NetworkEntry) string {
	if len(network) == 0 {
		return ""
	}
	sb := strings.Builder{}
	sb.Grow(len(network) * 256)
	sectionTitle := f.theme.FormatHeader("NETWORK REQUESTS")
	sb.WriteString("\n=== ")
	sb.WriteString(sectionTitle)
	sb.WriteString(" ===\n")
	for _, net := range network {
		formattedMethod := f.theme.Bold(net.Method)
		formattedURL := f.theme.Colorize(Cyan, net.URL)
		if net.Error != "" {
			errorColor := Red
			switch net.ErrorType {
			case "timeout":
				errorColor = Red
			case "cors":
				errorColor = Yellow
			case "dns":
				errorColor = Magenta
			}
			formattedError := f.theme.Colorize(errorColor, fmt.Sprintf("ERROR: %s (%s)", net.Error, net.ErrorType))
			sb.WriteString(formattedMethod)
			sb.WriteString(" ")
			sb.WriteString(formattedURL)
			sb.WriteString(" -> ")
			sb.WriteString(formattedError)
			sb.WriteString("\n")
		} else {
			statusColor := f.theme.GetStatusColor(net.Status)
			formattedStatus := f.theme.Colorize(statusColor, strconv.Itoa(net.Status))
			sb.WriteString(formattedMethod)
			sb.WriteString(" ")
			sb.WriteString(formattedURL)
			sb.WriteString(" -> ")
			sb.WriteString(formattedStatus)
			sb.WriteString("\n")
		}
		if len(net.Headers) > 0 {
			for k, v := range net.Headers {
				formattedKey := f.theme.Colorize(Yellow, k)
				sb.WriteString("  ")
				sb.WriteString(formattedKey)
				sb.WriteString(": ")
				sb.WriteString(v)
				sb.WriteString("\n")
			}
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

func (f *OutputFormatter) FormatSummary(logCount, networkCount int, duration time.Duration) string {
	var sb strings.Builder
	sectionTitle := f.theme.FormatHeader("SUMMARY")
	sb.WriteString(fmt.Sprintf("\n=== %s ===\n", sectionTitle))
	durationText := f.theme.Colorize(f.theme.Timestamp, fmt.Sprintf("Total Duration: %v", duration))
	logsText := f.theme.Colorize(f.theme.Timestamp, fmt.Sprintf("Console Logs: %d", logCount))
	networkText := f.theme.Colorize(f.theme.Timestamp, fmt.Sprintf("Network Requests: %d", networkCount))
	sb.WriteString(fmt.Sprintf("%s\n", durationText))
	sb.WriteString(fmt.Sprintf("%s\n", logsText))
	sb.WriteString(fmt.Sprintf("%s\n", networkText))
	return sb.String()
}

func (f *OutputFormatter) FormatSeparator() string {
	return f.theme.FormatSeparator("─", 60) + "\n"
}

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

func (f *OutputFormatter) FormatConsolePrefix() string {
	return f.theme.FormatConsolePrefix()
}

func (f *OutputFormatter) FormatNetworkPrefix() string {
	return f.theme.FormatNetworkPrefix()
}

func (f *OutputFormatter) Colorize(color, text string) string {
	return f.theme.Colorize(color, text)
}

func (f *OutputFormatter) FormatAndOutputLog(le LogEntry, cfg Config) error {
	timestamp := le.Time.Format("15:04:05")
	level := strings.ToUpper(le.Level)
	message := le.Message
	levelColor := f.GetLogLevelColor(level)
	var levelSymbol string
	switch level {
	case "DEBUG":
		levelSymbol = "🐛"
	case "INFO":
		levelSymbol = "ℹ️"
	case "WARN", "WARNING":
		levelSymbol = "⚠️"
	case "ERROR":
		levelSymbol = "❌"
	case "FATAL":
		levelSymbol = "💀"
	case "LOG":
		levelSymbol = "📝"
	case "TRACE":
		levelSymbol = "🔍"
	case "STARTGROUP", "STARTGROUPCOLLAPSED", "ENDGROUP":
		levelSymbol = "📂"
	case "DIR", "DIRXML":
		levelSymbol = "🗂️"
	case "TABLE":
		levelSymbol = "📊"
	case "TIMEEND":
		levelSymbol = "⏰"
	default:
		levelSymbol = "📋"
	}
	formattedTimestamp := f.FormatTimestamp(timestamp)
	formattedPrefix := f.FormatConsolePrefix()
	formattedSymbol := f.Colorize(levelColor, levelSymbol)
	formattedLevel := f.Colorize(levelColor, level)
	line := fmt.Sprintf("[%s] %s %s %s: %s\n",
		formattedTimestamp,
		formattedPrefix,
		formattedSymbol,
		formattedLevel,
		message)
	return WriteOutput(cfg, line)
}

func (f *OutputFormatter) FormatAndOutputNetwork(ne NetworkEntry, cfg Config) error {
	timestamp := ne.Timestamp.Format("15:04:05")
	method := ne.Method
	url := ne.URL
	status := ne.Status
	methodColor := f.GetHTTPMethodColor(method)
	var methodSymbol string
	switch method {
	case "GET":
		methodSymbol = "📥"
	case "POST":
		methodSymbol = "📤"
	case "PUT":
		methodSymbol = "🔄"
	case "DELETE":
		methodSymbol = "🗑️"
	case "PATCH":
		methodSymbol = "🔧"
	default:
		methodSymbol = "🌐"
	}
	formattedTimestamp := f.FormatTimestamp(timestamp)
	formattedPrefix := f.FormatNetworkPrefix()
	formattedSymbol := f.Colorize(methodColor, methodSymbol)
	formattedMethod := f.Colorize(methodColor, method)
	var line string
	if ne.Error != "" {
		errorColor := Red
		switch ne.ErrorType {
		case "timeout":
			errorColor = Red
		case "cors":
			errorColor = Yellow
		case "dns":
			errorColor = Magenta
		}
		formattedError := f.Colorize(errorColor, fmt.Sprintf("ERROR: %s (%s)", ne.Error, ne.ErrorType))
		line = fmt.Sprintf("[%s] %s %s %s %s %s\n",
			formattedTimestamp,
			formattedPrefix,
			formattedSymbol,
			formattedMethod,
			url,
			formattedError)
	} else {
		statusColor := f.GetStatusColor(status)
		formattedStatus := f.Colorize(statusColor, strconv.Itoa(status))
		line = fmt.Sprintf("[%s] %s %s %s %s %s\n",
			formattedTimestamp,
			formattedPrefix,
			formattedSymbol,
			formattedMethod,
			url,
			formattedStatus)
	}
	return WriteOutput(cfg, line)
}

func (f *OutputFormatter) FormatLogsCSV(logs []LogEntry, includeHeader bool) string {
	if len(logs) == 0 {
		return ""
	}
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	if includeHeader {
		_ = w.Write([]string{"timestamp", "level", "source", "message"})
	}
	for _, le := range logs {
		_ = w.Write([]string{le.Time.Format(time.RFC3339), strings.ToUpper(le.Level), le.Source, le.Message})
	}
	w.Flush()
	return buf.String()
}

func (f *OutputFormatter) FormatNetworkCSV(entries []NetworkEntry, includeHeader bool) string {
	if len(entries) == 0 {
		return ""
	}
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	if includeHeader {
		_ = w.Write([]string{"timestamp", "method", "url", "status", "resourceType", "mimeType", "size", "duration", "ttfb", "connectTime", "dnsTime", "sslTime", "sendTime", "waitTime", "receiveTime", "error", "errorType"})
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
		_ = w.Write(header)
	}
	_ = w.Write(row)
	w.Flush()
	return WriteOutput(cfg, buf.String())
}

func (f *OutputFormatter) FormatAndOutputLogCSVRow(le LogEntry, cfg Config, includeHeader bool) error {
	return f.formatAndOutputCSVRow(
		[]string{"timestamp", "level", "source", "message"},
		[]string{le.Time.Format(time.RFC3339), strings.ToUpper(le.Level), le.Source, le.Message},
		includeHeader, cfg)
}

func (f *OutputFormatter) FormatAndOutputNetworkCSVRow(ne NetworkEntry, cfg Config, includeHeader bool) error {
	return f.formatAndOutputCSVRow(
		[]string{"timestamp", "method", "url", "status", "resourceType", "mimeType", "size", "duration", "ttfb", "connectTime", "dnsTime", "sslTime", "sendTime", "waitTime", "receiveTime", "error", "errorType"},
		[]string{
			ne.Timestamp.Format(time.RFC3339), ne.Method, ne.URL, strconv.Itoa(ne.Status),
			ne.ResourceType, ne.Type, strconv.FormatInt(ne.Size, 10), strconv.FormatFloat(ne.Duration, 'f', 2, 64),
			strconv.FormatFloat(ne.TimeToFirstByte, 'f', 2, 64), strconv.FormatFloat(ne.ConnectTime, 'f', 2, 64),
			strconv.FormatFloat(ne.DNSLookupTime, 'f', 2, 64), strconv.FormatFloat(ne.SSLTime, 'f', 2, 64),
			strconv.FormatFloat(ne.SendTime, 'f', 2, 64), strconv.FormatFloat(ne.WaitTime, 'f', 2, 64),
			strconv.FormatFloat(ne.ReceiveTime, 'f', 2, 64), ne.Error, ne.ErrorType,
		},
		includeHeader, cfg)
}
