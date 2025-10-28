package helpers

import "strings"

const (
	// Reset all formatting
	Reset = "\033[0m"
	// Text colors
	Black   = "\033[30m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"
	Gray    = "\033[90m"
	// Bright colors
	BrightRed     = "\033[91m"
	BrightGreen   = "\033[92m"
	BrightYellow  = "\033[93m"
	BrightBlue    = "\033[94m"
	BrightMagenta = "\033[95m"
	BrightCyan    = "\033[96m"
	BrightWhite   = "\033[97m"
	// Text formatting
	Bold      = "\033[1m"
	Dim       = "\033[2m"
	Italic    = "\033[3m"
	Underline = "\033[4m"
	Blink     = "\033[5m"
	Reverse   = "\033[7m"
	Strike    = "\033[9m"
	// Background colors
	BgBlack   = "\033[40m"
	BgRed     = "\033[41m"
	BgGreen   = "\033[42m"
	BgYellow  = "\033[43m"
	BgBlue    = "\033[44m"
	BgMagenta = "\033[45m"
	BgCyan    = "\033[46m"
	BgWhite   = "\033[47m"
	BgGray    = "\033[100m"
)

type ColorTheme struct {
	// Log levels
	Debug    string
	Info     string
	Warn     string
	Error    string
	Fatal    string
	Success  string
	Progress string
	// UI elements
	Timestamp string
	Header    string
	Separator string
	Border    string
	// HTTP status codes
	StatusSuccess     string // 2xx
	StatusRedirect    string // 3xx
	StatusClientError string // 4xx
	StatusServerError string // 5xx
	// Special elements
	Checkmark string
	Arrow     string
	Cross     string
}

func DefaultTheme() *ColorTheme {
	return &ColorTheme{
		// Log levels
		Debug:    Cyan,
		Info:     Green,
		Warn:     Yellow,
		Error:    Red,
		Fatal:    Magenta,
		Success:  Green,
		Progress: Blue,
		// UI elements
		Timestamp: Gray,
		Header:    Bold + Gray,
		Separator: Gray,
		Border:    Bold + Gray,
		// HTTP status codes
		StatusSuccess:     Green,
		StatusRedirect:    Yellow,
		StatusClientError: Red,
		StatusServerError: Magenta,
		// Special elements
		Checkmark: Green,
		Arrow:     Blue,
		Cross:     Red,
	}
}

func DisabledTheme() *ColorTheme {
	return &ColorTheme{
		Debug:    "",
		Info:     "",
		Warn:     "",
		Error:    "",
		Fatal:    "",
		Success:  "",
		Progress: "",

		Timestamp: "",
		Header:    "",
		Separator: "",
		Border:    "",

		StatusSuccess:     "",
		StatusRedirect:    "",
		StatusClientError: "",
		StatusServerError: "",

		Checkmark: "",
		Arrow:     "",
		Cross:     "",
	}
}

func (ct *ColorTheme) Colorize(color, text string) string {
	if color == "" {
		return text
	}
	return color + text + Reset
}

func (ct *ColorTheme) Bold(text string) string {
	return ct.Colorize(Bold, text)
}

func (ct *ColorTheme) Dim(text string) string {
	return ct.Colorize(Dim, text)
}

func (ct *ColorTheme) GetStatusColor(statusCode int) string {
	if statusCode >= 200 && statusCode < 300 {
		return ct.StatusSuccess
	} else if statusCode >= 300 && statusCode < 400 {
		return ct.StatusRedirect
	} else if statusCode >= 400 && statusCode < 500 {
		return ct.StatusClientError
	} else if statusCode >= 500 {
		return ct.StatusServerError
	}
	return ""
}

func (ct *ColorTheme) GetLogLevelColor(level string) string {
	switch level {
	case "DEBUG", "debug":
		return ct.Debug
	case "INFO", "info":
		return ct.Info
	case "WARN", "WARNING", "warn", "warning":
		return ct.Warn
	case "ERROR", "error":
		return ct.Error
	case "FATAL", "fatal":
		return ct.Fatal
	default:
		return ""
	}
}

func (ct *ColorTheme) FormatLogLevel(level string) string {
	color := ct.GetLogLevelColor(level)
	return ct.Colorize(color, strings.ToUpper(level))
}

func (ct *ColorTheme) FormatTimestamp(timestamp string) string {
	return ct.Colorize(ct.Timestamp, timestamp)
}

func (ct *ColorTheme) FormatSuccess(message string) string {
	checkmark := ct.Colorize(ct.Checkmark, "✓")
	return checkmark + " " + message
}

func (ct *ColorTheme) FormatProgress(message string) string {
	arrow := ct.Colorize(ct.Arrow, "→")
	return arrow + " " + message
}

func (ct *ColorTheme) FormatError(message string) string {
	cross := ct.Colorize(ct.Cross, "✗")
	return cross + " " + message
}

func (ct *ColorTheme) FormatHeader(title string) string {
	return ct.Colorize(ct.Header, strings.ToUpper(title))
}

func (ct *ColorTheme) FormatSeparator(char string, length int) string {
	separator := strings.Repeat(char, length)
	return ct.Colorize(ct.Separator, separator)
}

func (ct *ColorTheme) FormatBorder(char string) string {
	return ct.Colorize(ct.Border, char)
}
