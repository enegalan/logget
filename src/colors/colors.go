package colors

import (
	_ "embed"
	"strings"
)

//go:embed colors.def
var colorsDefData string
var (
	Reset         string
	Black         string
	Red           string
	Green         string
	Yellow        string
	Blue          string
	Magenta       string
	Cyan          string
	White         string
	Gray          string
	BrightRed     string
	BrightGreen   string
	BrightYellow  string
	BrightBlue    string
	BrightMagenta string
	BrightCyan    string
	BrightWhite   string
	Bold          string
	Dim           string
	Italic        string
	Underline     string
	Blink         string
	Reverse       string
	Strike        string
	BgBlack       string
	BgRed         string
	BgGreen       string
	BgYellow      string
	BgBlue        string
	BgMagenta     string
	BgCyan        string
	BgWhite       string
	BgGray        string
)
var colorMap map[string]string

func init() {
	colorMap = make(map[string]string)
	lines := strings.Split(colorsDefData, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		value = strings.ReplaceAll(value, "\\033", "\x1b")
		colorMap[key] = value
	}
	Reset = colorMap["Reset"]
	Black = colorMap["Black"]
	Red = colorMap["Red"]
	Green = colorMap["Green"]
	Yellow = colorMap["Yellow"]
	Blue = colorMap["Blue"]
	Magenta = colorMap["Magenta"]
	Cyan = colorMap["Cyan"]
	White = colorMap["White"]
	Gray = colorMap["Gray"]
	BrightRed = colorMap["BrightRed"]
	BrightGreen = colorMap["BrightGreen"]
	BrightYellow = colorMap["BrightYellow"]
	BrightBlue = colorMap["BrightBlue"]
	BrightMagenta = colorMap["BrightMagenta"]
	BrightCyan = colorMap["BrightCyan"]
	BrightWhite = colorMap["BrightWhite"]
	Bold = colorMap["Bold"]
	Dim = colorMap["Dim"]
	Italic = colorMap["Italic"]
	Underline = colorMap["Underline"]
	Blink = colorMap["Blink"]
	Reverse = colorMap["Reverse"]
	Strike = colorMap["Strike"]
	BgBlack = colorMap["BgBlack"]
	BgRed = colorMap["BgRed"]
	BgGreen = colorMap["BgGreen"]
	BgYellow = colorMap["BgYellow"]
	BgBlue = colorMap["BgBlue"]
	BgMagenta = colorMap["BgMagenta"]
	BgCyan = colorMap["BgCyan"]
	BgWhite = colorMap["BgWhite"]
	BgGray = colorMap["BgGray"]
}

type ColorTheme struct {
	Enabled           bool
	Debug             string
	Info              string
	Warn              string
	Error             string
	Fatal             string
	Success           string
	Progress          string
	Timestamp         string
	Header            string
	Separator         string
	Border            string
	StatusSuccess     string // 2xx
	StatusRedirect    string // 3xx
	StatusClientError string // 4xx
	StatusServerError string // 5xx
	Checkmark         string
	Arrow             string
	Cross             string
}

func GetTheme(colors bool) *ColorTheme {
	if colors {
		return DefaultTheme()
	}
	return DisabledTheme()
}

func DefaultTheme() *ColorTheme {
	return &ColorTheme{
		Enabled:           true,
		Debug:             Cyan,
		Info:              Green,
		Warn:              Yellow,
		Error:             Red,
		Fatal:             Magenta,
		Success:           Green,
		Progress:          Blue,
		Timestamp:         Gray,
		Header:            Bold + Gray,
		Separator:         Gray,
		Border:            Bold + Gray,
		StatusSuccess:     Green,
		StatusRedirect:    Yellow,
		StatusClientError: Red,
		StatusServerError: Magenta,
		Checkmark:         Green,
		Arrow:             Blue,
		Cross:             Red,
	}
}

func DisabledTheme() *ColorTheme {
	return &ColorTheme{
		Enabled:           false,
		Debug:             "",
		Info:              "",
		Warn:              "",
		Error:             "",
		Fatal:             "",
		Success:           "",
		Progress:          "",
		Timestamp:         "",
		Header:            "",
		Separator:         "",
		Border:            "",
		StatusSuccess:     "",
		StatusRedirect:    "",
		StatusClientError: "",
		StatusServerError: "",
		Checkmark:         "",
		Arrow:             "",
		Cross:             "",
	}
}

func (ct *ColorTheme) Colorize(color, text string) string {
	if !ct.Enabled || color == "" {
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

func (ct *ColorTheme) FormatBorder(char string) string { return ct.Colorize(ct.Border, char) }

func (ct *ColorTheme) GetHTTPMethodColor(method string) string {
	switch method {
	case "GET":
		return ct.Success
	case "POST":
		return ct.Warn
	case "PUT":
		return ct.Progress
	case "DELETE":
		return ct.Error
	case "PATCH":
		return ct.Fatal
	default:
		return ct.Debug
	}
}

func (ct *ColorTheme) FormatHTTPMethod(method string) string {
	color := ct.GetHTTPMethodColor(method)
	return ct.Colorize(color, method)
}

func (ct *ColorTheme) FormatNetworkPrefix() string { return ct.Colorize(Bold+Cyan, "NETWORK") }

func (ct *ColorTheme) FormatConsolePrefix() string { return ct.Colorize(Bold+Magenta, "CONSOLE") }
