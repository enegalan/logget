package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	helpers "logget/helpers"
	"os"
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
)

func main() {
	log.SetOutput(io.Discard)
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
		fmt.Printf("logget %s\n", version)
		fmt.Printf("A command-line tool for extracting browser logs and network data from web pages\n")
		os.Exit(0)
	}
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Error: URL is required\n")
		fmt.Fprintf(os.Stderr, "Usage: logget [flags] <url>\n")
		fmt.Fprintf(os.Stderr, "Use 'logget --help' for more information\n")
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
		fmt.Println("logget: try 'logget --help' or 'logget -h' for more information")
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
	// Create context
	ctx, cancel = chromedp.NewContext(allocCtx)
	defer cancel()
	// Get initial response to capture redirect information
	initialProtocol, initialStatusCode, err := helpers.GetInitialResponse(cfg, url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to get initial response: %v\n", err)
		// Fallback to defaults
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
	// Enable CDP domains and set up event listeners
	if showLogs {
		// Enable the log domain for browser logs
		err := chromedp.Run(ctx, cdplog.Enable())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to enable log domain: %v\n", err)
			os.Exit(1)
		}
		// Enable runtime domain for JavaScript console logs
		err = chromedp.Run(ctx, runtime.Enable())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to enable runtime domain: %v\n", err)
			os.Exit(1)
		}
	}
	if showNetwork || verbose {
		// Enable network domain for network monitoring or protocol detection
		err := chromedp.Run(ctx, cdpnetwork.Enable())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to enable network domain: %v\n", err)
		}
	}
	// Set up event listeners for both logs and network
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
		err := helpers.SetCookies(ctx, url, cookies)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to set cookies: %v\n", err)
			os.Exit(1)
		}
	}
	// Navigate to the page
	if followMode {
		fmt.Fprintf(os.Stderr, "Streaming logs from %s (Press Ctrl+C to stop)\n", url)
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
			line := fmt.Sprintf("[%s] %s: %s\n", le.Time.Format("15:04:05"), strings.ToUpper(le.Level), le.Message)
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
			line := fmt.Sprintf("[%s] %s %s -> %d\n", ne.Timestamp.Format("15:04:05"), ne.Method, ne.URL, ne.Status)
			_ = helpers.WriteOutput(cfg, line)
		}
		if err := helpers.StreamLogsRealTime(cfg, ctx, url, onLog, onNet); err != nil {
			fmt.Fprintf(os.Stderr, "Error streaming logs: %v\n", err)
			os.Exit(1)
		}
		return
	}
	tasks := []chromedp.Action{
		chromedp.Navigate(url),
		chromedp.Sleep(time.Duration(wait) * time.Second), // Wait for the page to load
	}
	// Execute tasks
	err = chromedp.Run(ctx, tasks...)
	if err != nil {
		// For HTTP error responses, show basic info before failing
		if strings.Contains(err.Error(), "ERR_HTTP_RESPONSE_CODE_FAILURE") {
			if verbose {
				fmt.Printf("%s Error (navigation failed)\n", responseProtocol)
				fmt.Printf("Duration: %v\n", time.Since(startTime))
				fmt.Println()
			}
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		} else {
			fmt.Fprintf(os.Stderr, "Failed to navigate to %s: %v\n", url, err)
			os.Exit(1)
		}
	}
	// Use the status code we captured from the initial response
	statusCode := responseStatusCode
	duration := time.Since(startTime)
	// Prepare output data
	output := OutputData{
		URL:      url,
		Logs:     logs,
		Network:  network,
		Duration: duration,
	}
	// Output results
	if jsonOutput {
		jsonData, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to marshal JSON: %v\n", err)
			os.Exit(1)
		}
		err = helpers.WriteOutput(cfg, string(jsonData)+"\n")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write output: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Human-readable output
		var outputContent strings.Builder
		if verbose {
			outputContent.WriteString(fmt.Sprintf("%s %d\n", responseProtocol, statusCode))
			outputContent.WriteString(fmt.Sprintf("Duration: %v\n", output.Duration))
			outputContent.WriteString("\n")
			outputContent.WriteString("=== REQUEST HEADERS ===\n")
			dynamicHeaders := helpers.GenerateDynamicHeaders(cfg, url)
			for _, header := range dynamicHeaders {
				outputContent.WriteString(fmt.Sprintf("%s\n", header))
			}
			outputContent.WriteString("\n")
			if len(responseHeaders) > 0 {
				outputContent.WriteString("=== RESPONSE HEADERS ===\n")
				for name, value := range responseHeaders {
					outputContent.WriteString(fmt.Sprintf("%s: %s\n", name, value))
				}
				outputContent.WriteString("\n")
			}
		}
		if showLogs && len(logs) > 0 {
			outputContent.WriteString("=== CONSOLE LOGS ===\n")
			for _, log := range logs {
				outputContent.WriteString(fmt.Sprintf("[%s] %s: %s\n", log.Time.Format("15:04:05"), strings.ToUpper(log.Level), log.Message))
			}
			outputContent.WriteString("\n")
		}
		if showNetwork && len(network) > 0 {
			outputContent.WriteString("=== NETWORK REQUESTS ===\n")
			for _, net := range network {
				outputContent.WriteString(fmt.Sprintf("%s %s -> %d\n", net.Method, net.URL, net.Status))
				if len(net.Headers) > 0 {
					for k, v := range net.Headers {
						outputContent.WriteString(fmt.Sprintf("  %s: %s\n", k, v))
					}
				}
				outputContent.WriteString("\n")
			}
		}
		err := helpers.WriteOutput(cfg, outputContent.String())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write output: %v\n", err)
			os.Exit(1)
		}
	}
}
