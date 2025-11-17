package helpers

import (
	chrome "logget/src/chrome"
	"regexp"
)

func ShouldShowLine(line string, filterRegex *regexp.Regexp, excludeRegex *regexp.Regexp) bool {
	if excludeRegex != nil && excludeRegex.MatchString(line) {
		return false
	}
	if filterRegex != nil && !filterRegex.MatchString(line) {
		return false
	}
	return true
}

func shouldIncludeNetworkEntry(ne chrome.NetworkEntry, filterRegex *regexp.Regexp, excludeRegex *regexp.Regexp, minSize, maxSize int64) bool {
	if !ShouldShowLine(ne.URL, filterRegex, excludeRegex) {
		return false
	}
	if minSize > 0 && ne.Size < minSize {
		return false
	}
	if maxSize > 0 && ne.Size > maxSize {
		return false
	}
	return true
}

func FilterLog(log chrome.LogEntry, filterRegex *regexp.Regexp, excludeRegex *regexp.Regexp) bool {
	if filterRegex == nil && excludeRegex == nil {
		return true
	}
	return ShouldShowLine(log.Message, filterRegex, excludeRegex)
}

func FilterLogs(logs []chrome.LogEntry, filterRegex *regexp.Regexp, excludeRegex *regexp.Regexp) []chrome.LogEntry {
	var filtered []chrome.LogEntry
	for _, log := range logs {
		if FilterLog(log, filterRegex, excludeRegex) {
			filtered = append(filtered, log)
		}
	}
	return filtered
}

func FilterNetworkEntry(ne chrome.NetworkEntry, filterRegex *regexp.Regexp, excludeRegex *regexp.Regexp, minSize, maxSize int64) bool {
	if filterRegex == nil && excludeRegex == nil && minSize == 0 && maxSize == 0 {
		return true
	}
	return shouldIncludeNetworkEntry(ne, filterRegex, excludeRegex, minSize, maxSize)
}

func FilterNetworkEntries(network []chrome.NetworkEntry, filterRegex *regexp.Regexp, excludeRegex *regexp.Regexp, minSize, maxSize int64) []chrome.NetworkEntry {
	var filtered []chrome.NetworkEntry
	for _, ne := range network {
		if FilterNetworkEntry(ne, filterRegex, excludeRegex, minSize, maxSize) {
			filtered = append(filtered, ne)
		}
	}
	return filtered
}
