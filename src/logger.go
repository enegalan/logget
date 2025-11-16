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
	quiet   bool
	theme   *ColorTheme
}

func NewLogger(verbose bool, colors bool) *Logger {
	return &Logger{
		verbose: verbose,
		quiet:   false,
		theme:   GetTheme(colors),
	}
}

func (l *Logger) SetQuiet(quiet bool) { l.quiet = quiet }

func (l *Logger) formatMessage(level LogLevel, format string, args ...interface{}) string {
	timestamp := time.Now().Format("15:04:05")
	levelStr := level.String()
	message := fmt.Sprintf(format, args...)
	formattedTimestamp := l.theme.FormatTimestamp(timestamp)
	formattedLevel := l.theme.FormatLogLevel(levelStr)
	prefix := ""
	if level == WARN || level == ERROR || level == FATAL {
		prefix = "logget: "
	}
	return fmt.Sprintf("[%s] %s: %s%s", formattedTimestamp, formattedLevel, prefix, message)
}

func (l *Logger) Debug(format string, args ...interface{}) {
	if l.verbose {
		fmt.Fprintln(os.Stderr, l.formatMessage(DEBUG, format, args...))
	}
}

func (l *Logger) Info(format string, args ...interface{}) {
	if !l.quiet {
		fmt.Fprintln(os.Stderr, l.formatMessage(INFO, format, args...))
	}
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
	if !l.quiet {
		timestamp := time.Now().Format("15:04:05")
		message := fmt.Sprintf(format, args...)
		formattedTimestamp := l.theme.FormatTimestamp(timestamp)
		formattedMessage := l.theme.FormatSuccess(message)
		fmt.Fprintf(os.Stderr, "[%s] %s\n", formattedTimestamp, formattedMessage)
	}
}

func (l *Logger) Progress(format string, args ...interface{}) {
	if !l.quiet {
		timestamp := time.Now().Format("15:04:05")
		message := fmt.Sprintf(format, args...)
		formattedTimestamp := l.theme.FormatTimestamp(timestamp)
		formattedMessage := l.theme.FormatProgress(message)
		fmt.Fprintf(os.Stderr, "[%s] %s\n", formattedTimestamp, formattedMessage)
	}
}

func (l *Logger) Section(title string) {
	if !l.quiet {
		formattedTitle := l.theme.FormatHeader(title)
		fmt.Fprintf(os.Stderr, "\n=== %s ===\n", formattedTitle)
	}
}

func (l *Logger) Separator() {
	if !l.quiet {
		separator := l.theme.FormatSeparator("─", 60)
		fmt.Fprintf(os.Stderr, "%s\n", separator)
	}
}

func (l *Logger) PrintHeader(version string) {
	topBorder := l.theme.FormatBorder("┌─────────────────────────────────────────────────────────────┐")
	fmt.Fprintf(os.Stderr, "%s\n", topBorder)
	titleSpaces := strings.Repeat(" ", 51-len(version))
	titleLine := fmt.Sprintf("│ logget v%s%s │", version, titleSpaces)
	formattedTitleLine := l.theme.FormatBorder(titleLine)
	fmt.Fprintf(os.Stderr, "%s\n", formattedTitleLine)
	subtitleSpaces := strings.Repeat(" ", 21)
	subtitleLine := fmt.Sprintf("│ Web Log & Network Data Extraction Tool %s│", subtitleSpaces)
	formattedSubtitleLine := l.theme.FormatBorder(subtitleLine)
	fmt.Fprintf(os.Stderr, "%s\n", formattedSubtitleLine)
	bottomBorder := l.theme.FormatBorder("└─────────────────────────────────────────────────────────────┘")
	fmt.Fprintf(os.Stderr, "%s\n", bottomBorder)
	fmt.Fprintf(os.Stderr, "\n")
}

func (l *Logger) PrintUsage() {
	fmt.Fprintf(os.Stderr, "logget: try 'logget --help' for more information\n")
}

func (l *Logger) PrintError(err error) { fmt.Fprintf(os.Stderr, "logget: %v\n", err) }
