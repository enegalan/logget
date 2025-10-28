package helpers

import (
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
		// Method and URL
		formattedMethod := f.theme.Bold(net.Method)
		formattedURL := f.theme.Colorize(Cyan, net.URL)
		statusColor := f.theme.GetStatusColor(net.Status)
		formattedStatus := f.theme.Colorize(statusColor, fmt.Sprintf("%d", net.Status))
		sb.WriteString(fmt.Sprintf("%s %s -> %s\n", formattedMethod, formattedURL, formattedStatus))
		// Headers
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

// GetLogLevelColor returns the color for a log level
func (f *OutputFormatter) GetLogLevelColor(level string) string {
	return f.theme.GetLogLevelColor(level)
}

// GetHTTPMethodColor returns the color for an HTTP method
func (f *OutputFormatter) GetHTTPMethodColor(method string) string {
	return f.theme.GetHTTPMethodColor(method)
}

// GetStatusColor returns the color for an HTTP status code
func (f *OutputFormatter) GetStatusColor(statusCode int) string {
	return f.theme.GetStatusColor(statusCode)
}

// FormatTimestamp formats a timestamp with the theme
func (f *OutputFormatter) FormatTimestamp(timestamp string) string {
	return f.theme.FormatTimestamp(timestamp)
}

// FormatConsolePrefix formats the CONSOLE prefix
func (f *OutputFormatter) FormatConsolePrefix() string {
	return f.theme.FormatConsolePrefix()
}

// FormatNetworkPrefix formats the NETWORK prefix
func (f *OutputFormatter) FormatNetworkPrefix() string {
	return f.theme.FormatNetworkPrefix()
}

// Colorize applies color to text using the theme
func (f *OutputFormatter) Colorize(color, text string) string {
	return f.theme.Colorize(color, text)
}
