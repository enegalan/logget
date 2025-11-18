package chrome

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/go-rod/rod/lib/proto"
)

type LogEntry struct {
	Level   string    `json:"level"`
	Message string    `json:"message"`
	Time    time.Time `json:"time"`
	Source  string    `json:"source"`
}

type EventHandlers struct {
	OnLog               func(LogEntry)
	OnNetwork           func(NetworkEntry)
	OnRequestWillBeSent func(requestID string, method, url string, headers map[string]string, startTime float64)
}

var consoleAPIMessages = sync.Map{}

var exceptionMessages = sync.Map{}

func isJavaScriptException(message string) bool {
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?:^|\s)(https?://[^\s:]+):(\d+):(\d+)`),
		regexp.MustCompile(`at\s+(https?://[^\s:]+):(\d+):(\d+)`),
	}
	for _, re := range patterns {
		if matches := re.FindStringSubmatch(message); len(matches) >= 3 {
			locationKey := matches[1] + ":" + matches[2]
			if _, exists := exceptionMessages.Load(locationKey); exists {
				return true
			}
		}
	}
	return false
}

func isChromeInternalMessage(message string) bool {
	_, exists := consoleAPIMessages.Load(message)
	return !exists
}

func formatFrame(frame *proto.RuntimeCallFrame) string {
	if frame.URL == "" && frame.FunctionName == "" {
		return ""
	}
	if frame.FunctionName != "" {
		if frame.URL != "" {
			return fmt.Sprintf("\n    at %s (%s:%d:%d)", frame.FunctionName, frame.URL, frame.LineNumber+1, frame.ColumnNumber+1)
		}
		return fmt.Sprintf("\n    at %s", frame.FunctionName)
	}
	return fmt.Sprintf("\n    at %s:%d:%d", frame.URL, frame.LineNumber+1, frame.ColumnNumber+1)
}

func storeExceptionLocation(ed *proto.RuntimeExceptionDetails) {
	if ed == nil {
		return
	}
	var locationKey string
	if ed.URL != "" {
		locationKey = fmt.Sprintf("%s:%d", ed.URL, ed.LineNumber+1)
	} else if ed.StackTrace != nil && len(ed.StackTrace.CallFrames) > 0 {
		frame := ed.StackTrace.CallFrames[0]
		if frame.URL != "" {
			locationKey = fmt.Sprintf("%s:%d", frame.URL, frame.LineNumber+1)
		}
	}
	if locationKey != "" {
		exceptionMessages.Store(locationKey, true)
	}
}

func normalizeLogLevel(level string, message string) string {
	if strings.ToUpper(level) == "WARNING" && isJavaScriptException(message) {
		return "error"
	}
	return level
}

func createLogEntry(level, message, source string) LogEntry {
	return LogEntry{Level: level, Message: message, Time: time.Now(), Source: source}
}

func ProcessLogEvent(ev interface{}, handlers *EventHandlers) {
	if handlers == nil || handlers.OnLog == nil {
		return
	}
	if ev, ok := ev.(*proto.RuntimeExceptionThrown); ok {
		var message string
		if ev.ExceptionDetails != nil {
			if ev.ExceptionDetails.Exception != nil {
				if ev.ExceptionDetails.Exception.Description != "" {
					message = ev.ExceptionDetails.Exception.Description
				} else if ev.ExceptionDetails.Exception.ClassName != "" {
					message = ev.ExceptionDetails.Exception.ClassName
					if ev.ExceptionDetails.Text != "" {
						message += ": " + ev.ExceptionDetails.Text
					}
				}
			}
			if message == "" {
				message = ev.ExceptionDetails.Text
			}
			if ev.ExceptionDetails.StackTrace != nil && len(ev.ExceptionDetails.StackTrace.CallFrames) > 0 {
				for i, frame := range ev.ExceptionDetails.StackTrace.CallFrames {
					message += formatFrame(frame)
					if i == len(ev.ExceptionDetails.StackTrace.CallFrames)-1 && ev.ExceptionDetails.StackTrace.Parent != nil {
						parent := ev.ExceptionDetails.StackTrace.Parent
						for _, parentFrame := range parent.CallFrames {
							message += formatFrame(parentFrame)
						}
					}
				}
			} else if ev.ExceptionDetails.URL != "" {
				message += fmt.Sprintf(" at %s:%d:%d", ev.ExceptionDetails.URL, ev.ExceptionDetails.LineNumber+1, ev.ExceptionDetails.ColumnNumber+1)
			}
		}
		if message == "" {
			message = "Unhandled JavaScript exception"
		}
		storeExceptionLocation(ev.ExceptionDetails)
		handlers.OnLog(createLogEntry("error", message, "browser"))
		return
	}
	if ev, ok := ev.(*proto.LogEntryAdded); ok {
		if isChromeInternalMessage(ev.Entry.Text) {
			return
		}
		level := normalizeLogLevel(string(ev.Entry.Level), ev.Entry.Text)
		handlers.OnLog(createLogEntry(level, ev.Entry.Text, "browser"))
		return
	}
	if ev, ok := ev.(*proto.RuntimeConsoleAPICalled); ok {
		var message string
		for _, arg := range ev.Args {
			var str string
			argVal := arg.Value
			argBytes, _ := argVal.MarshalJSON()
			if err := json.Unmarshal(argBytes, &str); err == nil {
				message += str + " "
			} else {
				message += fmt.Sprintf("%v ", argVal)
			}
		}
		message = strings.TrimSpace(message)
		if message != "" {
			consoleAPIMessages.Store(message, true)
		}
		level := normalizeLogLevel(string(ev.Type), message)
		handlers.OnLog(createLogEntry(level, message, "console"))
	}
}
