package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	helpers "logget/helpers"
	"os"
	"path/filepath"
	"regexp"
	"strings"
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
	timeout          int
	wait             int
	userAgent        string
	headers          []string
	cookies          []string
	versionFlag      bool
	verbose          bool
	outputFile       string
	outputDir        string
	appendMode       bool
	version          string = "dev"
	responseHeaders  map[string]string
	responseCaptured bool

	followMode      bool
	filterPattern   string
	excludePattern  string
	refreshInterval int

	skipSSLVerify bool

	logger    *helpers.Logger
	formatter *helpers.OutputFormatter
)

func main() {
	log.SetOutput(io.Discard)
	logger = helpers.NewLogger(verbose, true)
	formatter = helpers.NewOutputFormatter(true)
	var rootCmd = &cobra.Command{
		Use:   "logget [flags] <url>",
		Short: "Extract logs and network data from web pages",
		Long:  ``,
		Args:  cobra.ArbitraryArgs,
		Run:   runLogget,
	}
	rootCmd.Flags().BoolVarP(&showLogs, "logs", "L", false, "Show console logs")
	rootCmd.Flags().BoolVarP(&showNetwork, "network", "N", false, "Show network requests")
	rootCmd.Flags().BoolVarP(&jsonOutput, "json", "J", false, "Output in JSON format")
	rootCmd.Flags().IntVarP(&timeout, "timeout", "T", 60, "Timeout in seconds")
	rootCmd.Flags().IntVarP(&wait, "wait", "W", 3, "Wait time in seconds after page load")
	rootCmd.Flags().StringVarP(&userAgent, "user-agent", "A", "logget/1.0", "Set User-Agent header")
	rootCmd.Flags().StringArrayVarP(&headers, "header", "H", []string{}, "Add custom headers (format: 'Key: Value')")
	rootCmd.Flags().StringArrayVarP(&cookies, "cookie", "C", []string{}, "Add cookies (format: 'name=value' or 'name=value; domain=example.com')")
	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Write to file instead of stdout")
	rootCmd.Flags().StringVar(&outputDir, "output-dir", "", "Directory to save files in")
	rootCmd.Flags().BoolVarP(&appendMode, "append", "a", false, "Append to file instead of overwriting")
	rootCmd.Flags().BoolVarP(&versionFlag, "version", "v", false, "Show version information")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "V", false, "Show detailed HTTP protocol information")
	rootCmd.Flags().BoolVarP(&followMode, "follow", "f", false, "Stream logs and network requests in real-time")
	rootCmd.Flags().StringVar(&filterPattern, "filter", "", "Show only logs/requests matching this regex pattern")
	rootCmd.Flags().StringVar(&excludePattern, "exclude", "", "Exclude logs/requests matching this regex pattern")
	rootCmd.Flags().IntVar(&refreshInterval, "refresh", 100, "Refresh interval in milliseconds for real-time streaming")
	rootCmd.Flags().BoolVarP(&skipSSLVerify, "insecure", "k", false, "Skip SSL certificate verification (useful for self-signed certificates)")
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runLogget(cmd *cobra.Command, args []string) {
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
		UserAgent:      userAgent,
		Headers:        headers,
		Cookies:        cookies,
		OutputFile:     outputFile,
		OutputDir:      outputDir,
		AppendMode:     appendMode,
		SkipSSLVerify:  skipSSLVerify,
		ShowNetwork:    showNetwork,
		JSONOutput:     jsonOutput,
		FilterPattern:  filterPattern,
		ExcludePattern: excludePattern,
	}
	// Quick check: if no data collection flags are specified, show help immediately
	if !showLogs && !showNetwork && !verbose && !jsonOutput && !followMode {
		logger.PrintUsage()
		os.Exit(0)
	}
	// Validate URL
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()
	// Create chromedp context
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
	// SSL bypass flags
	if skipSSLVerify {
		opts = append(opts,
			chromedp.Flag("ignore-certificate-errors-spki-list", true),
			chromedp.Flag("ignore-ssl-errors", true),
			chromedp.Flag("ignore-certificate-errors", true),
		)
	}
	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()
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
	responseHeaders = make(map[string]string) // Initialize response headers map
	responseCaptured = false                  // Reset flag
	startTime := time.Now()
	logger.Progress("Initializing browser context...")
	if showLogs {
		logger.Progress("Enabling console log monitoring...")
		err := chromedp.Run(ctx, cdplog.Enable())
		if err != nil {
			logger.Fatal("Failed to enable log domain: %v", err)
		}
		err = chromedp.Run(ctx, runtime.Enable())
		if err != nil {
			logger.Fatal("Failed to enable runtime domain: %v", err)
		}
	}
	if showNetwork || verbose {
		logger.Progress("Enabling network monitoring...")
		err := chromedp.Run(ctx, cdpnetwork.Enable())
		if err != nil {
			logger.Error("Failed to enable network domain: %v", err)
		}
	}
	chromedp.ListenTarget(ctx, func(ev interface{}) {
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
						// Try to unmarshal as string first
						var str string
						if err := json.Unmarshal(arg.Value, &str); err == nil {
							message += str + " "
						} else { // If not a string, convert to string representation
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
				response := ev.Response
				// Convert headers to map
				headers := make(map[string]string)
				for name, value := range response.Headers {
					if str, ok := value.(string); ok {
						headers[name] = str
					} else {
						headers[name] = fmt.Sprintf("%v", value)
					}
				}
				if verbose && !responseCaptured {
					responseHeaders = headers
					responseCaptured = true
				}
				if showNetwork {
					network = append(network, NetworkEntry{
						URL:       response.URL,
						Method:    "GET",
						Status:    int(response.Status),
						Headers:   headers,
						Timestamp: time.Now(),
						Type:      string(response.MimeType),
						Size:      int64(response.EncodedDataLength),
					})
				}
			}
		}
	})
	// Set cookies if provided
	if len(cookies) > 0 {
		logger.Progress("Setting cookies...")
		err := helpers.SetCookies(ctx, url, cookies)
		if err != nil {
			logger.Fatal("Failed to set cookies: %v", err)
		}
	}
	// Navigate to the page
	if followMode {
		logger.Progress("Streaming logs from %s (Press Ctrl+C to stop)", url)
		var filterRegex *regexp.Regexp
		var excludeRegex *regexp.Regexp
		if filterPattern != "" {
			if r, err := regexp.Compile(filterPattern); err == nil {
				filterRegex = r
			}
		}
		if excludePattern != "" {
			if r, err := regexp.Compile(excludePattern); err == nil {
				excludeRegex = r
			}
		}
		onLog := func(le helpers.LogEntry) {
			if !helpers.ShouldShowLine(le.Message, filterRegex, excludeRegex) {
				return
			}
			if jsonOutput {
				b, _ := json.Marshal(le)
				_ = helpers.WriteOutput(cfg, string(b)+"\n")
				return
			}
			timestamp := le.Time.Format("15:04:05")
			level := strings.ToUpper(le.Level)
			message := le.Message
			levelColor := formatter.GetLogLevelColor(level)
			var levelSymbol string
			switch level {
			case "DEBUG":
				levelSymbol = "🐛"
			case "INFO":
				levelSymbol = "ℹ️"
			case "WARN", "WARNING":
				levelSymbol = "⚠️"
			case "ERROR":
				levelSymbol = "❌"
			case "FATAL":
				levelSymbol = "💀"
			case "LOG":
				levelSymbol = "📝"
			case "TRACE":
				levelSymbol = "🔍"
			default:
				levelSymbol = "📋"
			}
			formattedTimestamp := formatter.FormatTimestamp(timestamp)
			formattedPrefix := formatter.FormatConsolePrefix()
			formattedSymbol := formatter.Colorize(levelColor, levelSymbol)
			formattedLevel := formatter.Colorize(levelColor, level)
			line := fmt.Sprintf("[%s] %s %s %s: %s\n",
				formattedTimestamp,
				formattedPrefix,
				formattedSymbol,
				formattedLevel,
				message)
			_ = helpers.WriteOutput(cfg, line)
		}
		onNet := func(ne helpers.NetworkEntry) {
			if !helpers.ShouldShowLine(ne.URL, filterRegex, excludeRegex) {
				return
			}
			if jsonOutput {
				b, _ := json.Marshal(ne)
				_ = helpers.WriteOutput(cfg, string(b)+"\n")
				return
			}
			timestamp := ne.Timestamp.Format("15:04:05")
			method := ne.Method
			url := ne.URL
			status := ne.Status
			methodColor := formatter.GetHTTPMethodColor(method)
			var methodSymbol string
			switch method {
			case "GET":
				methodSymbol = "📥"
			case "POST":
				methodSymbol = "📤"
			case "PUT":
				methodSymbol = "🔄"
			case "DELETE":
				methodSymbol = "🗑️"
			case "PATCH":
				methodSymbol = "🔧"
			default:
				methodSymbol = "🌐"
			}
			statusColor := formatter.GetStatusColor(status)
			formattedTimestamp := formatter.FormatTimestamp(timestamp)
			formattedPrefix := formatter.FormatNetworkPrefix()
			formattedSymbol := formatter.Colorize(methodColor, methodSymbol)
			formattedMethod := formatter.Colorize(methodColor, method)
			formattedStatus := formatter.Colorize(statusColor, fmt.Sprintf("%d", status))
			line := fmt.Sprintf("[%s] %s %s %s %s %s\n",
				formattedTimestamp,
				formattedPrefix,
				formattedSymbol,
				formattedMethod,
				url,
				formattedStatus)
			_ = helpers.WriteOutput(cfg, line)
		}
		if err := helpers.StreamLogsRealTime(cfg, ctx, url, onLog, onNet); err != nil {
			logger.Fatal("Error streaming logs: %v", err)
		}
		return
	}
	logger.Progress("Navigating to %s...", url)
	tasks := []chromedp.Action{
		chromedp.Navigate(url),
		chromedp.Sleep(time.Duration(wait) * time.Second), // Wait for the page to load
	}
	// Execute tasks
	err = chromedp.Run(ctx, tasks...)
	if err != nil { // For HTTP error responses, show basic info before failing
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
		err = helpers.WriteOutput(cfg, string(jsonData)+"\n")
		if err != nil {
			logger.Fatal("Failed to write output: %v", err)
		}
		if cfg.OutputFile != "" {
			var filePath string
			if cfg.OutputDir != "" {
				filePath = filepath.Join(cfg.OutputDir, cfg.OutputFile)
			} else {
				filePath = cfg.OutputFile
			}
			if cfg.AppendMode {
				logger.Success("JSON output appended to: %s", filePath)
			} else {
				logger.Success("JSON output written to: %s", filePath)
			}
		}
	} else {
		var outputContent strings.Builder
		// HTTP Response section
		if verbose {
			outputContent.WriteString(formatter.FormatHTTPResponse(responseProtocol, statusCode, duration))
			outputContent.WriteString(formatter.FormatRequestHeaders(helpers.GenerateDynamicHeaders(cfg, url)))
			outputContent.WriteString(formatter.FormatResponseHeaders(responseHeaders))
		}
		// Console logs section
		if showLogs && len(logs) > 0 {
			outputContent.WriteString(formatter.FormatConsoleLogs(logs))
		}
		// Network requests section
		if showNetwork && len(network) > 0 {
			outputContent.WriteString(formatter.FormatNetworkRequests(network))
		}
		err := helpers.WriteOutput(cfg, outputContent.String())
		if err != nil {
			logger.Fatal("Failed to write output: %v", err)
		}
		if cfg.OutputFile != "" {
			var filePath string
			if cfg.OutputDir != "" {
				filePath = filepath.Join(cfg.OutputDir, cfg.OutputFile)
			} else {
				filePath = cfg.OutputFile
			}
			if cfg.AppendMode {
				logger.Success("Output appended to: %s", filePath)
			} else {
				logger.Success("Output written to: %s", filePath)
			}
		}
	}
}
