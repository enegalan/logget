package helpers

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strings"
	"time"
)

type OutputFormatter struct {
	theme *ColorTheme
}

func NewOutputFormatter(colors bool) *OutputFormatter {
	var theme *ColorTheme
	if colors {
		theme = DefaultTheme()
	} else {
		theme = DisabledTheme()
	}
	return &OutputFormatter{
		theme: theme,
	}
}

func (f *OutputFormatter) FormatHTTPResponse(protocol string, statusCode int, duration time.Duration) string {
	var sb strings.Builder
	// Status line
	statusColor := f.theme.GetStatusColor(statusCode)
	formattedProtocol := f.theme.Bold(protocol)
	formattedStatusCode := f.theme.Colorize(statusColor, fmt.Sprintf("%d", statusCode))
	sb.WriteString(fmt.Sprintf("%s %s\n", formattedProtocol, formattedStatusCode))
	// Duration
	durationText := f.theme.Colorize(f.theme.Timestamp, fmt.Sprintf("Duration: %v", duration))
	sb.WriteString(fmt.Sprintf("%s\n", durationText))
	return sb.String()
}

func (f *OutputFormatter) FormatRequestHeaders(headers []string) string {
	var sb strings.Builder
	sectionTitle := f.theme.FormatHeader("REQUEST HEADERS")
	sb.WriteString(fmt.Sprintf("\n=== %s ===\n", sectionTitle))
	for _, header := range headers {
		formattedHeader := f.theme.Colorize(Yellow, header)
		sb.WriteString(fmt.Sprintf("%s\n", formattedHeader))
	}
	return sb.String()
}

func (f *OutputFormatter) FormatResponseHeaders(headers map[string]string) string {
	var sb strings.Builder
	if len(headers) == 0 {
		return ""
	}
	sectionTitle := f.theme.FormatHeader("RESPONSE HEADERS")
	sb.WriteString(fmt.Sprintf("\n=== %s ===\n", sectionTitle))
	for name, value := range headers {
		formattedName := f.theme.Bold(name)
		sb.WriteString(fmt.Sprintf("%s: %s\n", formattedName, value))
	}
	return sb.String()
}

func (f *OutputFormatter) FormatConsoleLogs(logs []LogEntry) string {
	var sb strings.Builder
	if len(logs) == 0 {
		return ""
	}
	sectionTitle := f.theme.FormatHeader("CONSOLE LOGS")
	sb.WriteString(fmt.Sprintf("\n=== %s ===\n", sectionTitle))
	for _, log := range logs {
		timestamp := log.Time.Format("15:04:05")
		level := strings.ToUpper(log.Level)
		formattedTimestamp := f.theme.FormatTimestamp(timestamp)
		formattedLevel := f.theme.FormatLogLevel(level)
		sb.WriteString(fmt.Sprintf("[%s] %s: %s\n", formattedTimestamp, formattedLevel, log.Message))
	}
	return sb.String()
}

func (f *OutputFormatter) FormatNetworkRequests(network []NetworkEntry) string {
	var sb strings.Builder
	if len(network) == 0 {
		return ""
	}
	sectionTitle := f.theme.FormatHeader("NETWORK REQUESTS")
	sb.WriteString(fmt.Sprintf("\n=== %s ===\n", sectionTitle))
	for _, net := range network {
		formattedMethod := f.theme.Bold(net.Method)
		formattedURL := f.theme.Colorize(Cyan, net.URL)
		statusColor := f.theme.GetStatusColor(net.Status)
		formattedStatus := f.theme.Colorize(statusColor, fmt.Sprintf("%d", net.Status))
		sb.WriteString(fmt.Sprintf("%s %s -> %s\n", formattedMethod, formattedURL, formattedStatus))
		if len(net.Headers) > 0 {
			for k, v := range net.Headers {
				formattedKey := f.theme.Colorize(Yellow, k)
				sb.WriteString(fmt.Sprintf("  %s: %s\n", formattedKey, v))
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
	statusColor := f.GetStatusColor(status)
	formattedTimestamp := f.FormatTimestamp(timestamp)
	formattedPrefix := f.FormatNetworkPrefix()
	formattedSymbol := f.Colorize(methodColor, methodSymbol)
	formattedMethod := f.Colorize(methodColor, method)
	formattedStatus := f.Colorize(statusColor, fmt.Sprintf("%d", status))
	line := fmt.Sprintf("[%s] %s %s %s %s %s\n",
		formattedTimestamp,
		formattedPrefix,
		formattedSymbol,
		formattedMethod,
		url,
		formattedStatus)
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
		_ = w.Write([]string{"timestamp", "method", "url", "status", "resourceType", "mimeType", "size"})
	}
	for _, ne := range entries {
		_ = w.Write([]string{ne.Timestamp.Format(time.RFC3339), ne.Method, ne.URL, fmt.Sprintf("%d", ne.Status), ne.ResourceType, ne.Type, fmt.Sprintf("%d", ne.Size)})
	}
	w.Flush()
	return buf.String()
}

func (f *OutputFormatter) FormatAndOutputLogCSVRow(le LogEntry, cfg Config, includeHeader bool) error {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	if includeHeader {
		if err := w.Write([]string{"timestamp", "level", "source", "message"}); err != nil {
			return fmt.Errorf("failed to write CSV header: %v", err)
		}
	}
	if err := w.Write([]string{le.Time.Format(time.RFC3339), strings.ToUpper(le.Level), le.Source, le.Message}); err != nil {
		return fmt.Errorf("failed to write CSV row: %v", err)
	}
	w.Flush()
	return WriteOutput(cfg, buf.String())
}

func (f *OutputFormatter) FormatAndOutputNetworkCSVRow(ne NetworkEntry, cfg Config, includeHeader bool) error {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	if includeHeader {
		if err := w.Write([]string{"timestamp", "method", "url", "status", "resourceType", "mimeType", "size"}); err != nil {
			return fmt.Errorf("failed to write CSV header: %v", err)
		}
	}
	if err := w.Write([]string{ne.Timestamp.Format(time.RFC3339), ne.Method, ne.URL, fmt.Sprintf("%d", ne.Status), ne.ResourceType, ne.Type, fmt.Sprintf("%d", ne.Size)}); err != nil {
		return fmt.Errorf("failed to write CSV row: %v", err)
	}
	w.Flush()
	return WriteOutput(cfg, buf.String())
}
