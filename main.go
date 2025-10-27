package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	cdplog "github.com/chromedp/cdproto/log"
	cdpnetwork "github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"github.com/spf13/cobra"
)

type LogEntry struct {
	Level   string    `json:"level"`
	Message string    `json:"message"`
	Time    time.Time `json:"time"`
	Source  string    `json:"source"`
}

type NetworkEntry struct {
	URL       string            `json:"url"`
	Method    string            `json:"method"`
	Status    int               `json:"status"`
	Headers   map[string]string `json:"headers"`
	Timestamp time.Time         `json:"timestamp"`
	Type      string            `json:"type"`
	Size      int64             `json:"size"`
}

type OutputData struct {
	URL      string         `json:"url"`
	Logs     []LogEntry     `json:"logs,omitempty"`
	Network  []NetworkEntry `json:"network,omitempty"`
	Duration time.Duration  `json:"duration"`
}

var (
	showLogs    bool
	showNetwork bool
	jsonOutput  bool
	timeout     int
	wait        int
	userAgent   string
	headers     []string
	versionFlag bool
	verbose     bool
	outputFile  string
	outputDir   string
	version     string = "dev" // Will be set via ldflags during build
)

func getHostFromURL(url string) string {
	if strings.HasPrefix(url, "http://") {
		url = url[7:]
	} else if strings.HasPrefix(url, "https://") {
		url = url[8:]
	}
	if slashIndex := strings.Index(url, "/"); slashIndex != -1 {
		return url[:slashIndex]
	}
	return url
}

func getPathFromURL(url string) string {
	if strings.HasPrefix(url, "http://") {
		url = url[7:]
	} else if strings.HasPrefix(url, "https://") {
		url = url[8:]
	}
	// Find the first slash to get the path
	if slashIndex := strings.Index(url, "/"); slashIndex != -1 {
		return url[slashIndex:]
	}
	return "/"
}

func generateDynamicHeaders(url string, userAgent string, customHeaders []string) []string {
	var headers []string
	// Basic headers that are always present
	headers = append(headers, fmt.Sprintf("Host: %s", getHostFromURL(url)))
	headers = append(headers, fmt.Sprintf("User-Agent: %s", userAgent))
	// Accept header based on what we're requesting
	if strings.Contains(url, ".json") || strings.Contains(url, "/api/") || strings.Contains(url, "api.") {
		headers = append(headers, "Accept: application/json,text/plain,*/*")
	} else if strings.Contains(url, ".css") {
		headers = append(headers, "Accept: text/css,*/*;q=0.1")
	} else if strings.Contains(url, ".js") {
		headers = append(headers, "Accept: */*")
	} else {
		headers = append(headers, "Accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	}
	// Language based on system locale (simplified)
	headers = append(headers, "Accept-Language: en-US,en;q=0.5")
	// Encoding based on what the browser supports
	headers = append(headers, "Accept-Encoding: gzip, deflate")
	// Connection type
	headers = append(headers, "Connection: keep-alive")
	// Security headers for HTTPS
	if strings.HasPrefix(url, "https://") {
		headers = append(headers, "Upgrade-Insecure-Requests: 1")
		headers = append(headers, "Sec-Fetch-Dest: document")
		headers = append(headers, "Sec-Fetch-Mode: navigate")
		headers = append(headers, "Sec-Fetch-Site: none")
	}
	// Cache control
	headers = append(headers, "Cache-Control: max-age=0")
	// Add custom headers
	headers = append(headers, customHeaders...)
	return headers
}

func writeOutput(content string) error {
	if outputFile != "" {
		// Determine the full file path
		var filePath string
		if outputDir != "" {
			// Create the output directory if it doesn't exist
			err := os.MkdirAll(outputDir, 0755)
			if err != nil {
				return fmt.Errorf("failed to create output directory: %v", err)
			}
			// Join directory and filename
			filePath = filepath.Join(outputDir, outputFile)
		} else {
			filePath = outputFile
		}

		// Write to file
		file, err := os.Create(filePath)
		if err != nil {
			return fmt.Errorf("failed to create output file: %v", err)
		}
		defer file.Close()

		_, err = file.WriteString(content)
		if err != nil {
			return fmt.Errorf("failed to write to output file: %v", err)
		}
		fmt.Fprintf(os.Stderr, "Output written to: %s\n", filePath)
	} else {
		// Write to stdout
		fmt.Print(content)
	}
	return nil
}

func main() {
	log.SetOutput(io.Discard)
	var rootCmd = &cobra.Command{
		Use:   "logget <url>",
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
	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Write to file instead of stdout")
	rootCmd.Flags().StringVar(&outputDir, "output-dir", "", "Directory to save files in")
	rootCmd.Flags().BoolVarP(&versionFlag, "version", "v", false, "Show version information")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "V", false, "Show detailed HTTP protocol information")
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
		fmt.Fprintf(os.Stderr, "Usage: logget <url> [flags]\n")
		fmt.Fprintf(os.Stderr, "Use 'logget --help' for more information\n")
		os.Exit(1)
	}
	// Quick check: if no data collection flags are specified, show help immediately
	if !showLogs && !showNetwork && !verbose && !jsonOutput {
		fmt.Println("logget: try 'logget --help' or 'logget -h' for more information")
		os.Exit(0)
	}
	url := args[0]
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
	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()
	// Create context
	ctx, cancel = chromedp.NewContext(allocCtx)
	defer cancel()
	// Collect logs and network data
	var logs []LogEntry
	var network []NetworkEntry
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
	if showNetwork {
		// Enable network domain for network monitoring
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
		if showNetwork {
			// Network response received
			if ev, ok := ev.(*cdpnetwork.EventResponseReceived); ok {
				// Get response details
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
				// Determine request method (default to GET)
				method := "GET"
				if ev.RequestID != "" {
					// Try to get method from request, but CDP doesn't always provide it
					// We'll use GET as default for most cases
				}
				network = append(network, NetworkEntry{
					URL:       response.URL,
					Method:    method,
					Status:    int(response.Status),
					Headers:   headers,
					Timestamp: time.Now(),
					Type:      string(response.MimeType),
					Size:      int64(response.EncodedDataLength),
				})
			}
		}
	})
	// Navigate to the page
	tasks := []chromedp.Action{
		chromedp.Navigate(url),
		chromedp.Sleep(time.Duration(wait) * time.Second), // Wait for the page to load
	}
	// Execute tasks
	err := chromedp.Run(ctx, tasks...)
	if err != nil {
		// For HTTP error responses, show basic info before failing
		if strings.Contains(err.Error(), "ERR_HTTP_RESPONSE_CODE_FAILURE") {
			if verbose {
				fmt.Printf("URL: %s\n", url)
				fmt.Printf("Status: HTTP Error (navigation failed)\n")
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
	// Always collect basic network info (status code)
	var statusCode int
	err = chromedp.Run(ctx, chromedp.Evaluate(`
		// Try to get the status code from the current page
		if (window.performance && window.performance.getEntriesByType) {
			const entries = window.performance.getEntriesByType('navigation');
			if (entries.length > 0) entries[0].responseStatus || 200;
			else 200;
		} else 200;
	`, &statusCode))
	if err != nil {
		statusCode = 200 // Default to 200 if we can't determine
	}
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
		err = writeOutput(string(jsonData) + "\n")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write output: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Human-readable output
		var outputContent strings.Builder

		if verbose {
			outputContent.WriteString(fmt.Sprintf("URL: %s\n", output.URL))
			outputContent.WriteString(fmt.Sprintf("Status: %d\n", statusCode))
			outputContent.WriteString(fmt.Sprintf("Duration: %v\n", output.Duration))
			outputContent.WriteString("\n")
			// Show detailed HTTP protocol information
			outputContent.WriteString("=== HTTP REQUEST ===\n")
			outputContent.WriteString(fmt.Sprintf("GET %s HTTP/1.1\n", getPathFromURL(url)))
			dynamicHeaders := generateDynamicHeaders(url, userAgent, headers)
			for _, header := range dynamicHeaders {
				outputContent.WriteString(fmt.Sprintf("%s\n", header))
			}
			outputContent.WriteString("\n")
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

		err := writeOutput(outputContent.String())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write output: %v\n", err)
			os.Exit(1)
		}
	}
}
