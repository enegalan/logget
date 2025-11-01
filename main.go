package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	helpers "logget/helpers"
	"logget/helpers/flags"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"

	cdplog "github.com/chromedp/cdproto/log"
	cdpnetwork "github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"github.com/spf13/cobra"
)

type LogEntry = helpers.LogEntry
type NetworkEntry = helpers.NetworkEntry

type OutputData struct {
	URL      string         `json:"url"`
	Logs     []LogEntry     `json:"logs,omitempty"`
	Network  []NetworkEntry `json:"network,omitempty"`
	Duration time.Duration  `json:"duration"`
}

var (
	showLogs         bool
	showNetwork      bool
	jsonOutput       bool
	csvOutput        bool
	timeout          int
	wait             int
	userAgent        flags.UserAgent = "logget/1.0"
	headers          flags.HeaderArray
	cookies          flags.CookieArray
	versionFlag      bool
	verbose          bool
	outputFile       flags.OutputFile
	appendMode       bool
	version          string = "dev"
	responseHeaders  map[string]string
	responseCaptured bool
	requestHeaders   map[string]string
	requestCaptured  bool

	followMode      bool
	filterPattern   flags.RegexPattern
	excludePattern  flags.RegexPattern
	statusPattern   flags.RegexPattern
	domainPattern   flags.RegexPattern
	mimePattern     flags.RegexPattern
	refreshInterval int

	skipSSLVerify bool

	logger        *helpers.Logger
	formatter     *helpers.OutputFormatter
	xhrOnly       bool
	documentOnly  bool
	cssOnly       bool
	scriptOnly    bool
	fontOnly      bool
	imgOnly       bool
	mediaOnly     bool
	manifestOnly  bool
	websocketOnly bool
	noColor       bool
	quiet         bool
	minSize       flags.SizeBytes
	maxSize       flags.SizeBytes
)

func main() {
	log.SetOutput(io.Discard)
	logger = helpers.NewLogger(verbose, true)
	formatter = helpers.NewOutputFormatter(true)
	var rootCmd = &cobra.Command{
		Use:   "logget [flags] <url>",
		Short: "Extract logs and network data from web pages",
		Args:  cobra.ArbitraryArgs,
		Run:   runLogget,
	}
	rootCmd.Flags().BoolVarP(&showLogs, "logs", "L", false, "Show console logs")
	rootCmd.Flags().BoolVarP(&showNetwork, "network", "N", false, "Show network requests")
	rootCmd.Flags().BoolVarP(&jsonOutput, "json", "J", false, "Output in JSON format")
	rootCmd.Flags().BoolVar(&csvOutput, "csv", false, "Output in CSV format")
	rootCmd.Flags().IntVarP(&timeout, "timeout", "T", 60, "Timeout in seconds")
	rootCmd.Flags().IntVarP(&wait, "wait", "W", 3000, "Wait time in milliseconds after page load")
	rootCmd.Flags().VarP(&userAgent, "user-agent", "A", "Set User-Agent header")
	rootCmd.Flags().VarP(&headers, "header", "H", "Add custom headers (format: 'Key: Value') or filename containing headers")
	rootCmd.Flags().VarP(&cookies, "cookie", "C", "Add cookies (format: 'name=value' or 'name=value; domain=example.com') or filename containing cookies")
	rootCmd.Flags().VarP(&outputFile, "output", "o", "Write to file instead of stdout")
	rootCmd.Flags().BoolVarP(&appendMode, "append", "a", false, "Append to file instead of overwriting")
	rootCmd.Flags().BoolVarP(&versionFlag, "version", "v", false, "Show version information")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "V", false, "Show detailed HTTP protocol information")
	rootCmd.Flags().BoolVarP(&followMode, "follow", "f", false, "Stream logs and network requests in real-time")
	rootCmd.Flags().VarP(&filterPattern, "filter", "", "Show only logs/requests matching this regex pattern")
	rootCmd.Flags().VarP(&excludePattern, "exclude", "", "Exclude logs/requests matching this regex pattern")
	rootCmd.Flags().VarP(&statusPattern, "status", "", "Only include requests whose HTTP status code matches this regex pattern")
	rootCmd.Flags().VarP(&domainPattern, "domain", "", "Only include requests whose domain matches this regex pattern")
	rootCmd.Flags().VarP(&mimePattern, "mime", "", "Only include requests whose MIME type matches this regex pattern")
	rootCmd.Flags().IntVar(&refreshInterval, "refresh", 100, "Refresh interval in milliseconds for real-time streaming")
	rootCmd.Flags().BoolVarP(&skipSSLVerify, "insecure", "k", false, "Skip SSL certificate verification (useful for self-signed certificates)")
	rootCmd.Flags().BoolVar(&noColor, "no-color", false, "Disable colored output")
	rootCmd.Flags().BoolVar(&xhrOnly, "xhr", false, "Only include fetch/XHR requests")
	rootCmd.Flags().BoolVar(&documentOnly, "document", false, "Only include Document requests")
	rootCmd.Flags().BoolVar(&cssOnly, "css", false, "Only include CSS requests")
	rootCmd.Flags().BoolVar(&scriptOnly, "script", false, "Only include Script requests")
	rootCmd.Flags().BoolVar(&fontOnly, "font", false, "Only include Font requests")
	rootCmd.Flags().BoolVar(&imgOnly, "img", false, "Only include Image requests")
	rootCmd.Flags().BoolVar(&mediaOnly, "media", false, "Only include Media requests")
	rootCmd.Flags().BoolVar(&manifestOnly, "manifest", false, "Only include Manifest requests")
	rootCmd.Flags().BoolVar(&websocketOnly, "socket", false, "Only include WebSocket requests")
	rootCmd.Flags().VarP(&minSize, "min-size", "", "Only include requests whose size is at least this many bytes")
	rootCmd.Flags().VarP(&maxSize, "max-size", "", "Only include requests whose size is at most this many bytes")
	rootCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "Suppress progress messages, only show data")
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runLogget(cmd *cobra.Command, args []string) {
	logger = helpers.NewLogger(verbose, !noColor)
	logger.SetQuiet(quiet)
	formatter = helpers.NewOutputFormatter(!noColor)
	if versionFlag {
		logger.PrintHeader(version)
		os.Exit(0)
	}
	if len(args) == 0 {
		logger.PrintError(fmt.Errorf("URL is required"))
		logger.PrintUsage()
		os.Exit(1)
	}
	url := args[0]
	processURL(url)
}

func processURL(url string) {
	cfg := helpers.Config{
		UserAgent:      userAgent.Get(),
		Headers:        []string(headers),
		Cookies:        []string(cookies),
		OutputFile:     outputFile.Get(),
		AppendMode:     appendMode,
		FollowMode:     followMode,
		SkipSSLVerify:  skipSSLVerify,
		ShowNetwork:    showNetwork,
		ShowLogs:       showLogs,
		JSONOutput:     jsonOutput,
		FilterPattern:  filterPattern.Get(),
		ExcludePattern: excludePattern.Get(),
		StatusPattern:  statusPattern.Get(),
		DomainPattern:  domainPattern.Get(),
		MimePattern:    mimePattern.Get(),
		XHROnly:        xhrOnly,
		DocumentOnly:   documentOnly,
		CssOnly:        cssOnly,
		ScriptOnly:     scriptOnly,
		FontOnly:       fontOnly,
		ImgOnly:        imgOnly,
		MediaOnly:      mediaOnly,
		ManifestOnly:   manifestOnly,
		WebSocketOnly:  websocketOnly,
		MinSize:        minSize.Get(),
		MaxSize:        maxSize.Get(),
	}
	// Quick check: if no data collection flags are specified, show help immediately
	if !showLogs && !showNetwork && !verbose && !jsonOutput && !followMode {
		logger.PrintUsage()
		os.Exit(0)
	}
	// Validate and normalize URL
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}
	// Create browser context
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("disable-plugins", true),
		chromedp.Flag("disable-web-security", true),
		chromedp.Flag("disable-features", "VizDisplayCompositor"),
		chromedp.Flag("ignore-certificate-errors", true),
		chromedp.Flag("ignore-ssl-errors", true),
		chromedp.Flag("allow-running-insecure-content", true),
		chromedp.Flag("disable-certificate-verification", true),
	)
	if skipSSLVerify {
		opts = append(opts,
			chromedp.Flag("ignore-certificate-errors-spki-list", true),
			chromedp.Flag("ignore-ssl-errors", true),
			chromedp.Flag("ignore-certificate-errors", true),
		)
	}
	allocCtx, allocCancel := chromedp.NewExecAllocator(ctx, opts...)
	defer allocCancel()
	ctx, cancel = chromedp.NewContext(allocCtx)
	defer cancel()
	initialProtocol, initialStatusCode, err := helpers.GetInitialResponse(cfg, url)
	if err != nil {
		logger.Warn("Failed to get initial response: %v", err)
		initialProtocol = "HTTP/1.1"
		initialStatusCode = 200
	}
	// Collect logs and network data
	var logs []LogEntry
	var network []NetworkEntry
	var responseProtocol string = initialProtocol
	var responseStatusCode int = initialStatusCode
	responseHeaders = make(map[string]string)
	responseCaptured = false
	requestHeaders = make(map[string]string)
	requestCaptured = false
	startTime := time.Now()
	logger.Progress("Initializing browser context...")
	if showLogs {
		logger.Progress("Enabling console log monitoring...")
		if err := chromedp.Run(ctx, cdplog.Enable(), runtime.Enable()); err != nil {
			logger.Fatal("Failed to enable log domains: %v", err)
		}
	}
	if showNetwork || verbose {
		logger.Progress("Enabling network monitoring...")
		if err := chromedp.Run(ctx, cdpnetwork.Enable()); err != nil {
			logger.Error("Failed to enable network domain: %v", err)
		}
	}
	requestMethods := sync.Map{}
	requestHeadersMap := sync.Map{}
	requestURLs := sync.Map{}
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		if showNetwork || verbose {
			if ev, ok := ev.(*cdpnetwork.EventRequestWillBeSent); ok {
				if ev.Request != nil {
					requestMethods.Store(ev.RequestID.String(), ev.Request.Method)
					requestURLs.Store(ev.RequestID.String(), ev.Request.URL)
					headers := make(map[string]string)
					for name, value := range ev.Request.Headers {
						if str, ok := value.(string); ok {
							headers[name] = str
						} else {
							headers[name] = fmt.Sprintf("%v", value)
						}
					}
					requestHeadersMap.Store(ev.RequestID.String(), headers)
					if verbose && !requestCaptured && ev.Request.URL == url {
						requestHeaders = headers
						requestCaptured = true
					}
				}
			}
			if ev, ok := ev.(*cdpnetwork.EventRequestWillBeSentExtraInfo); ok {
				if urlVal, ok := requestURLs.Load(ev.RequestID.String()); ok {
					if requestURL, ok := urlVal.(string); ok && requestURL == url {
						headers := make(map[string]string)
						for name, value := range ev.Headers {
							if str, ok := value.(string); ok {
								headers[name] = str
							} else {
								headers[name] = fmt.Sprintf("%v", value)
							}
						}
						if verbose {
							requestHeaders = headers
							requestCaptured = true
						}
					}
				}
			}
		}
		// Browser logs
		if showLogs {
			if ev, ok := ev.(*cdplog.EventEntryAdded); ok {
				logs = append(logs, LogEntry{
					Level:   ev.Entry.Level.String(),
					Message: ev.Entry.Text,
					Time:    time.Now(),
					Source:  "browser",
				})
			}
			// JavaScript console logs
			if ev, ok := ev.(*runtime.EventConsoleAPICalled); ok {
				var message string
				for _, arg := range ev.Args {
					if arg.Value != nil {
						var str string
						if err := json.Unmarshal(arg.Value, &str); err == nil {
							message += str + " "
						} else {
							message += fmt.Sprintf("%v ", arg.Value)
						}
					}
				}
				logs = append(logs, LogEntry{
					Level:   ev.Type.String(),
					Message: strings.TrimSpace(message),
					Time:    time.Now(),
					Source:  "console",
				})
			}
		}
		// Network events
		if showNetwork || verbose {
			if ev, ok := ev.(*cdpnetwork.EventResponseReceived); ok {
				if !helpers.ShouldIncludeNetworkEvent(cfg, ev) {
					return
				}
				var method string
				if methodVal, ok := requestMethods.Load(ev.RequestID.String()); ok {
					if methodStr, ok := methodVal.(string); ok {
						method = methodStr
					}
				}
				ne := helpers.BuildNetworkEntryFromEvent(ev, method)
				if verbose && !responseCaptured {
					responseHeaders = ne.Headers
					responseCaptured = true
				}
				if showNetwork {
					network = append(network, ne)
				}
			}
			// Network loading failures
			if ev, ok := ev.(*cdpnetwork.EventLoadingFailed); ok {
				ne := helpers.HandleLoadingFailedEvent(ev, &requestMethods, &requestURLs)
				if ne != nil && showNetwork {
					network = append(network, *ne)
				}
			}
		}
	})
	if len(headers) > 0 || userAgent.Get() != "" {
		if err := helpers.SetHeaders(ctx, userAgent.Get(), []string(headers)); err != nil {
			logger.Fatal("Failed to set headers: %v", err)
		}
	}
	// Set cookies if provided
	if len(cookies) > 0 {
		logger.Progress("Setting cookies...")
		if err := helpers.SetCookies(ctx, url, []string(cookies)); err != nil {
			logger.Fatal("Failed to set cookies: %v", err)
		}
	}
	// Handle follow mode
	if followMode {
		if cfg.OutputFile != "" {
			filePath := cfg.OutputFile
			var testFile *os.File
			var testErr error
			if cfg.AppendMode {
				testFile, testErr = os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			} else {
				testFile, testErr = os.Create(filePath)
			}
			if testErr != nil {
				logger.Fatal("Failed to create/open output file '%s': %v", filePath, testErr)
			}
			testFile.Close()
			if !cfg.AppendMode {
				if err := os.Truncate(filePath, 0); err != nil {
					logger.Warn("Failed to truncate output file: %v", err)
				}
			}
			logger.Progress("Streaming logs from %s (Press Ctrl+C to stop) -> %s", url, filePath)
		} else {
			logger.Progress("Streaming logs from %s (Press Ctrl+C to stop)", url)
		}
		var filterRegex, excludeRegex *regexp.Regexp
		if !filterPattern.Empty() {
			if r, err := regexp.Compile(filterPattern.Get()); err == nil {
				filterRegex = r
			}
		}
		if !excludePattern.Empty() {
			if r, err := regexp.Compile(excludePattern.Get()); err == nil {
				excludeRegex = r
			}
		}
		headerWrittenLogs := false
		headerWrittenNet := false
		outputErrorCount := 0
		const maxOutputErrors = 5
		onLog := func(le helpers.LogEntry) {
			if !helpers.ShouldShowLine(le.Message, filterRegex, excludeRegex) {
				return
			}
			if jsonOutput {
				b, _ := json.Marshal(le)
				if err := helpers.WriteOutput(cfg, string(b)+"\n"); err != nil {
					outputErrorCount++
					if outputErrorCount <= maxOutputErrors {
						logger.Warn("Failed to write to output file: %v", err)
					}
					if outputErrorCount == maxOutputErrors {
						logger.Warn("Too many output errors, suppressing further warnings...")
					}
				}
				return
			}
			if csvOutput {
				if err := formatter.FormatAndOutputLogCSVRow(le, cfg, !headerWrittenLogs); err != nil {
					outputErrorCount++
					if outputErrorCount <= maxOutputErrors {
						logger.Warn("Failed to write CSV log row: %v", err)
					}
					if outputErrorCount == maxOutputErrors {
						logger.Warn("Too many output errors, suppressing further warnings...")
					}
				} else {
					headerWrittenLogs = true
				}
				return
			}
			if err := formatter.FormatAndOutputLog(le, cfg); err != nil {
				outputErrorCount++
				if outputErrorCount <= maxOutputErrors {
					logger.Warn("Failed to write log entry: %v", err)
				}
				if outputErrorCount == maxOutputErrors {
					logger.Warn("Too many output errors, suppressing further warnings...")
				}
			}
		}
		onNet := func(ne helpers.NetworkEntry) {
			if !helpers.ShouldShowLine(ne.URL, filterRegex, excludeRegex) {
				return
			}
			if cfg.MinSize > 0 && ne.Size < cfg.MinSize {
				return
			}
			if cfg.MaxSize > 0 && ne.Size > cfg.MaxSize {
				return
			}
			if jsonOutput {
				b, _ := json.Marshal(ne)
				if err := helpers.WriteOutput(cfg, string(b)+"\n"); err != nil {
					outputErrorCount++
					if outputErrorCount <= maxOutputErrors {
						logger.Warn("Failed to write to output file: %v", err)
					}
					if outputErrorCount == maxOutputErrors {
						logger.Warn("Too many output errors, suppressing further warnings...")
					}
				}
				return
			}
			if csvOutput {
				if err := formatter.FormatAndOutputNetworkCSVRow(ne, cfg, !headerWrittenNet); err != nil {
					outputErrorCount++
					if outputErrorCount <= maxOutputErrors {
						logger.Warn("Failed to write CSV network row: %v", err)
					}
					if outputErrorCount == maxOutputErrors {
						logger.Warn("Too many output errors, suppressing further warnings...")
					}
				} else {
					headerWrittenNet = true
				}
				return
			}
			if err := formatter.FormatAndOutputNetwork(ne, cfg); err != nil {
				outputErrorCount++
				if outputErrorCount <= maxOutputErrors {
					logger.Warn("Failed to write network entry: %v", err)
				}
				if outputErrorCount == maxOutputErrors {
					logger.Warn("Too many output errors, suppressing further warnings...")
				}
			}
		}
		sigCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()
		if err := helpers.StreamLogsRealTime(cfg, sigCtx, url, onLog, onNet); err != nil {
			logger.Fatal("Error streaming logs: %v", err)
		}
		return
	}
	logger.Progress("Navigating to %s...", url)
	tasks := []chromedp.Action{
		chromedp.Navigate(url),
		chromedp.Sleep(time.Duration(wait) * time.Millisecond),
	}
	if err = chromedp.Run(ctx, tasks...); err != nil {
		if strings.Contains(err.Error(), "ERR_HTTP_RESPONSE_CODE_FAILURE") {
			if verbose {
				fmt.Printf("%s Error (navigation failed)\n", responseProtocol)
				fmt.Printf("Duration: %v\n", time.Since(startTime))
				fmt.Println()
			}
			logger.Fatal("Navigation failed: %v", err)
		} else {
			logger.Fatal("Failed to navigate to %s: %v", url, err)
		}
	}
	logger.Success("Successfully loaded page: %s", url)
	statusCode := responseStatusCode
	duration := time.Since(startTime)
	output := OutputData{
		URL:      url,
		Logs:     logs,
		Network:  network,
		Duration: duration,
	}
	if jsonOutput {
		jsonData, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			logger.Fatal("Failed to marshal JSON: %v", err)
		}
		if err = helpers.WriteOutput(cfg, string(jsonData)+"\n"); err != nil {
			logger.Fatal("Failed to write output: %v", err)
		}
		if cfg.OutputFile != "" {
			helpers.LogOutputFileSuccess(cfg, "JSON output", logger)
		}
	} else {
		var outputContent strings.Builder
		if verbose && !jsonOutput && !csvOutput {
			outputContent.WriteString(formatter.FormatHTTPResponse(responseProtocol, statusCode, duration))
			var requestHeadersList []string
			if requestCaptured && len(requestHeaders) > 0 {
				for name, value := range requestHeaders {
					requestHeadersList = append(requestHeadersList, fmt.Sprintf("%s: %s", name, value))
				}
			} else {
				requestHeadersList = helpers.GenerateHeaders(cfg, url)
			}
			outputContent.WriteString(formatter.FormatRequestHeaders(requestHeadersList))
			outputContent.WriteString(formatter.FormatResponseHeaders(responseHeaders))
		}
		if showLogs && len(logs) > 0 {
			outputContent.WriteString(formatter.FormatConsoleLogs(logs))
		}
		if showNetwork && len(network) > 0 {
			outputContent.WriteString(formatter.FormatNetworkRequests(network))
		}
		if err = helpers.WriteOutput(cfg, outputContent.String()); err != nil {
			logger.Fatal("Failed to write output: %v", err)
		}
		if cfg.OutputFile != "" {
			helpers.LogOutputFileSuccess(cfg, "Output", logger)
		}
	}
}
