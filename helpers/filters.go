package helpers

import "regexp"

func ShouldShowLine(line string, filterRegex *regexp.Regexp, excludeRegex *regexp.Regexp) bool {
	if excludeRegex != nil && excludeRegex.MatchString(line) {
		return false
	}
	if filterRegex != nil && !filterRegex.MatchString(line) {
		return false
	}
	return true
}
