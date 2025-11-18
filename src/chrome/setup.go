package chrome

import (
	"context"
	"fmt"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

func GetChromeOptions(skipSSLVerify bool) []string {
	args := []string{
		"--headless",
		"--disable-gpu",
		"--disable-dev-shm-usage",
		"--no-sandbox",
		"--disable-extensions",
		"--disable-plugins",
		"--disable-web-security",
		"--disable-features=VizDisplayCompositor",
		"--ignore-certificate-errors",
		"--ignore-ssl-errors",
		"--allow-running-insecure-content",
		"--disable-certificate-verification",
		"--disable-background-networking",
		"--disable-background-timer-throttling",
		"--disable-breakpad",
		"--disable-client-side-phishing-detection",
		"--disable-default-apps",
		"--disable-hang-monitor",
		"--disable-popup-blocking",
		"--disable-prompt-on-repost",
		"--disable-sync",
		"--disable-translate",
		"--metrics-recording-only",
		"--no-first-run",
		"--safebrowsing-disable-auto-update",
		"--enable-automation=false",
		"--password-store=basic",
		"--use-mock-keychain",
	}
	if skipSSLVerify {
		args = append(args,
			"--ignore-certificate-errors-spki-list",
			"--ignore-ssl-errors",
			"--ignore-certificate-errors",
		)
	}
	return args
}

func ConvertEventHeaders(headersMap map[string]interface{}) map[string]string {
	headers := make(map[string]string, len(headersMap))
	for name, value := range headersMap {
		if str, ok := value.(string); ok {
			headers[name] = str
		} else {
			headers[name] = fmt.Sprintf("%v", value)
		}
	}
	return headers
}

type ChromeContext struct {
	Browser *rod.Browser
	Page    *rod.Page
	Cancel  context.CancelFunc
	Ctx     context.Context
}

func CreateChromeContext(ctx context.Context, skipSSLVerify bool) (*ChromeContext, context.CancelFunc, error) {
	launcher := launcher.New().
		Headless(true).
		NoSandbox(true).
		Set("disable-gpu").
		Set("disable-dev-shm-usage").
		Set("disable-extensions").
		Set("disable-plugins").
		Set("disable-web-security").
		Set("disable-features", "VizDisplayCompositor").
		Set("ignore-certificate-errors").
		Set("ignore-ssl-errors").
		Set("allow-running-insecure-content").
		Set("disable-certificate-verification").
		Set("disable-background-networking").
		Set("disable-background-timer-throttling").
		Set("disable-breakpad").
		Set("disable-client-side-phishing-detection").
		Set("disable-default-apps").
		Set("disable-hang-monitor").
		Set("disable-popup-blocking").
		Set("disable-prompt-on-repost").
		Set("disable-sync").
		Set("disable-translate").
		Set("metrics-recording-only").
		Set("no-first-run").
		Set("safebrowsing-disable-auto-update").
		Set("enable-automation", "false").
		Set("password-store", "basic").
		Set("use-mock-keychain")
	if skipSSLVerify {
		launcher = launcher.
			Set("ignore-certificate-errors-spki-list").
			Set("ignore-ssl-errors").
			Set("ignore-certificate-errors")
	}
	url, err := launcher.Launch()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to launch browser: %v", err)
	}
	browser := rod.New().ControlURL(url).Context(ctx)
	if err := browser.Connect(); err != nil {
		return nil, nil, fmt.Errorf("failed to connect to browser: %v", err)
	}
	page, err := browser.Page(proto.TargetCreateTarget{})
	if err != nil {
		browser.Close()
		return nil, nil, fmt.Errorf("failed to create page: %v", err)
	}
	page.MustEval("() => {}")
	cancel := func() {
		page.Close()
		browser.Close()
		launcher.Cleanup()
	}
	return &ChromeContext{
		Browser: browser,
		Page:    page,
		Cancel:  cancel,
		Ctx:     ctx,
	}, cancel, nil
}
