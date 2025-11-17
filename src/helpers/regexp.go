package helpers

import "regexp"

func CompileRegexPattern(pattern string) *regexp.Regexp {
	if pattern == "" {
		return nil
	}
	if r, err := regexp.Compile(pattern); err == nil {
		return r
	}
	return nil
}

func CompileRegexPatterns(filterPattern, excludePattern string) (*regexp.Regexp, *regexp.Regexp) {
	return CompileRegexPattern(filterPattern), CompileRegexPattern(excludePattern)
}
