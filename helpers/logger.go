package helpers

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

type Logger struct {
	verbose bool
	theme   *ColorTheme
}

func NewLogger(verbose bool, colors bool) *Logger {
	var theme *ColorTheme
	if colors {
		theme = DefaultTheme()
	} else {
		theme = DisabledTheme()
	}
	return &Logger{
		verbose: verbose,
		theme:   theme,
	}
}

func (l *Logger) formatMessage(level LogLevel, format string, args ...interface{}) string {
	timestamp := time.Now().Format("15:04:05")
	levelStr := level.String()
	message := fmt.Sprintf(format, args...)
	formattedTimestamp := l.theme.FormatTimestamp(timestamp)
	formattedLevel := l.theme.FormatLogLevel(levelStr)
	return fmt.Sprintf("[%s] %s: %s", formattedTimestamp, formattedLevel, message)
}

func (l *Logger) Debug(format string, args ...interface{}) {
	if l.verbose {
		fmt.Fprintln(os.Stderr, l.formatMessage(DEBUG, format, args...))
	}
}

func (l *Logger) Info(format string, args ...interface{}) {
	fmt.Fprintln(os.Stderr, l.formatMessage(INFO, format, args...))
}

func (l *Logger) Warn(format string, args ...interface{}) {
	fmt.Fprintln(os.Stderr, l.formatMessage(WARN, format, args...))
}

func (l *Logger) Error(format string, args ...interface{}) {
	fmt.Fprintln(os.Stderr, l.formatMessage(ERROR, format, args...))
}

func (l *Logger) Fatal(format string, args ...interface{}) {
	fmt.Fprintln(os.Stderr, l.formatMessage(FATAL, format, args...))
	os.Exit(1)
}

func (l *Logger) Success(format string, args ...interface{}) {
	timestamp := time.Now().Format("15:04:05")
	message := fmt.Sprintf(format, args...)
	formattedTimestamp := l.theme.FormatTimestamp(timestamp)
	formattedMessage := l.theme.FormatSuccess(message)
	fmt.Fprintf(os.Stderr, "[%s] %s\n", formattedTimestamp, formattedMessage)
}

func (l *Logger) Progress(format string, args ...interface{}) {
	timestamp := time.Now().Format("15:04:05")
	message := fmt.Sprintf(format, args...)
	formattedTimestamp := l.theme.FormatTimestamp(timestamp)
	formattedMessage := l.theme.FormatProgress(message)
	fmt.Fprintf(os.Stderr, "[%s] %s\n", formattedTimestamp, formattedMessage)
}

func (l *Logger) Section(title string) {
	formattedTitle := l.theme.FormatHeader(title)
	fmt.Fprintf(os.Stderr, "\n=== %s ===\n", formattedTitle)
}

func (l *Logger) Separator() {
	separator := l.theme.FormatSeparator("─", 60)
	fmt.Fprintf(os.Stderr, "%s\n", separator)
}

func (l *Logger) PrintHeader(version string) {
	// Top border
	topBorder := l.theme.FormatBorder("┌─────────────────────────────────────────────────────────────┐")
	fmt.Fprintf(os.Stderr, "%s\n", topBorder)
	// Title line
	titleSpaces := strings.Repeat(" ", 51-len(version))
	titleLine := fmt.Sprintf("│ logget v%s%s │", version, titleSpaces)
	formattedTitleLine := l.theme.FormatBorder(titleLine)
	fmt.Fprintf(os.Stderr, "%s\n", formattedTitleLine)
	// Subtitle line
	subtitleSpaces := strings.Repeat(" ", 21)
	subtitleLine := fmt.Sprintf("│ Web Log & Network Data Extraction Tool %s│", subtitleSpaces)
	formattedSubtitleLine := l.theme.FormatBorder(subtitleLine)
	fmt.Fprintf(os.Stderr, "%s\n", formattedSubtitleLine)
	// Bottom border
	bottomBorder := l.theme.FormatBorder("└─────────────────────────────────────────────────────────────┘")
	fmt.Fprintf(os.Stderr, "%s\n", bottomBorder)
	fmt.Fprintf(os.Stderr, "\n")
}

func (l *Logger) PrintUsage() {
	usageText := l.theme.Bold("Usage:")
	fmt.Fprintf(os.Stderr, "%s logget [flags] <url>\n", usageText)
	helpText := l.theme.Dim("Use 'logget --help' for detailed information")
	fmt.Fprintf(os.Stderr, "%s\n", helpText)
}

func (l *Logger) PrintError(err error) {
	errorText := l.theme.Bold("Error:")
	fmt.Fprintf(os.Stderr, "%s %v\n", errorText, err)
}

func (l *Logger) PrintWarning(format string, args ...interface{}) {
	warningText := l.theme.Bold("Warning:")
	message := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "%s %s\n", warningText, message)
}
