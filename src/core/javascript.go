package core

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	chrome "logget/src/chrome"
)

func ExecuteJavaScript(ctx *chrome.ChromeContext, codeOrFile string) (interface{}, error) {
	code, err := readJavaScriptCode(codeOrFile)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(code) == "" {
		return nil, fmt.Errorf("empty JavaScript code")
	}
	trimmedCode := strings.TrimSpace(code)
	var codeToExecute string
	if strings.Contains(trimmedCode, ";") || strings.Contains(trimmedCode, "\n") {
		codeToExecute = fmt.Sprintf("() => { %s }", trimmedCode)
	} else {
		codeToExecute = fmt.Sprintf("() => { return (%s); }", trimmedCode)
	}
	result, err := ctx.Page.Eval(codeToExecute)
	if err != nil {
		errMsg := "Uncaught " + strings.TrimPrefix(err.Error(), "eval js error: ")
		return nil, fmt.Errorf("%s", errMsg)
	}
	if result == nil {
		return nil, nil
	}
	var value interface{}
	jsonBytes, err := json.Marshal(result.Value)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %v", err)
	}
	if len(jsonBytes) > 0 && string(jsonBytes) != "null" {
		if err := json.Unmarshal(jsonBytes, &value); err != nil {
			value = string(jsonBytes)
		}
	}
	return value, nil
}

func readJavaScriptCode(codeOrFile string) (string, error) {
	if strings.TrimSpace(codeOrFile) == "" {
		return "", fmt.Errorf("empty JavaScript code")
	}
	fileInfo, err := os.Stat(codeOrFile)
	if err != nil {
		return codeOrFile, nil
	}
	if fileInfo.IsDir() {
		return codeOrFile, nil
	}
	content, err := os.ReadFile(codeOrFile)
	if err != nil {
		return "", fmt.Errorf("failed to read JavaScript file %s: %v", codeOrFile, err)
	}
	return string(content), nil
}
