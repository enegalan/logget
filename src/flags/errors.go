package flags

import "strings"

func FormatUnknownFlag(flag string, isShort bool) string {
	if isShort {
		return "option -" + flag + ": is unknown"
	}
	return "option --" + flag + ": is unknown"
}

func ExtractFlagName(flagStr string) string {
	flagStr = strings.Trim(flagStr, "'\"")
	if idx := strings.Index(flagStr, " "); idx != -1 {
		flagStr = flagStr[:idx]
	}
	if idx := strings.Index(flagStr, " in"); idx != -1 {
		flagStr = flagStr[:idx]
	}
	return flagStr
}
