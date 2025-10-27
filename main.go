package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
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
	cookies     []string
	versionFlag bool
	verbose     bool
	outputFile  string
	outputDir   string
	appendMode  bool
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

// getInitialResponse makes an HTTP request to capture the initial response and redirect information
func getInitialResponse(targetURL string, userAgent string, customHeaders []string) (string, int, error) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Don't follow redirects - we want to capture the redirect response
			return http.ErrUseLastResponse
		},
		Timeout: 5 * time.Second,
	}
	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return "", 0, err
	}
	// Set User-Agent
	req.Header.Set("User-Agent", userAgent)
	// Set custom headers
	for _, header := range customHeaders {
		if colonIndex := strings.Index(header, ":"); colonIndex != -1 {
			key := strings.TrimSpace(header[:colonIndex])
			value := strings.TrimSpace(header[colonIndex+1:])
			req.Header.Set(key, value)
		}
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()
	// Return protocol and status code
	return resp.Proto, resp.StatusCode, nil
}

func generateDynamicHeaders(url string, userAgent string, customHeaders []string) []string {
	customHeaderMap := make(map[string]string)
	for _, header := range customHeaders {
		if colonIndex := strings.Index(header, ":"); colonIndex != -1 {
			key := strings.TrimSpace(header[:colonIndex])
			value := strings.TrimSpace(header[colonIndex+1:])
			customHeaderMap[strings.ToLower(key)] = value
		}
	}
	var headers []string
	headers = append(headers, fmt.Sprintf("Host: %s", getHostFromURL(url)))
	// User-Agent
	if customUA, exists := customHeaderMap["user-agent"]; exists {
		headers = append(headers, fmt.Sprintf("User-Agent: %s", customUA))
	} else {
		headers = append(headers, fmt.Sprintf("User-Agent: %s", userAgent))
	}
	// Accept
	if customAccept, exists := customHeaderMap["accept"]; exists {
		headers = append(headers, fmt.Sprintf("Accept: %s", customAccept))
	} else {
		// Inherit Accept header based on the URL
		if strings.Contains(url, ".json") || strings.Contains(url, "/api/") || strings.Contains(url, "api.") {
			headers = append(headers, "Accept: application/json,text/plain,*/*")
		} else if strings.Contains(url, ".css") {
			headers = append(headers, "Accept: text/css,*/*;q=0.1")
		} else {
			headers = append(headers, "Accept: */*")
		}
	}
	// Accept-Language
	if customLang, exists := customHeaderMap["accept-language"]; exists {
		headers = append(headers, fmt.Sprintf("Accept-Language: %s", customLang))
	} else {
		headers = append(headers, "Accept-Language: en-US,en;q=0.5")
	}
	// Accept-Encoding
	if customEncoding, exists := customHeaderMap["accept-encoding"]; exists {
		headers = append(headers, fmt.Sprintf("Accept-Encoding: %s", customEncoding))
	} else {
		headers = append(headers, "Accept-Encoding: gzip, deflate")
	}
	// Connection
	if customConn, exists := customHeaderMap["connection"]; exists {
		headers = append(headers, fmt.Sprintf("Connection: %s", customConn))
	} else {
		headers = append(headers, "Connection: keep-alive")
	}
	// Security headers
	if strings.HasPrefix(url, "https://") {
		if customUpgrade, exists := customHeaderMap["upgrade-insecure-requests"]; exists {
			headers = append(headers, fmt.Sprintf("Upgrade-Insecure-Requests: %s", customUpgrade))
		} else {
			headers = append(headers, "Upgrade-Insecure-Requests: 1")
		}
		if customDest, exists := customHeaderMap["sec-fetch-dest"]; exists {
			headers = append(headers, fmt.Sprintf("Sec-Fetch-Dest: %s", customDest))
		} else {
			headers = append(headers, "Sec-Fetch-Dest: document")
		}
		if customMode, exists := customHeaderMap["sec-fetch-mode"]; exists {
			headers = append(headers, fmt.Sprintf("Sec-Fetch-Mode: %s", customMode))
		} else {
			headers = append(headers, "Sec-Fetch-Mode: navigate")
		}
		if customSite, exists := customHeaderMap["sec-fetch-site"]; exists {
			headers = append(headers, fmt.Sprintf("Sec-Fetch-Site: %s", customSite))
		} else {
			headers = append(headers, "Sec-Fetch-Site: none")
		}
	}
	// Cache control
	if customCache, exists := customHeaderMap["cache-control"]; exists {
		headers = append(headers, fmt.Sprintf("Cache-Control: %s", customCache))
	} else {
		headers = append(headers, "Cache-Control: max-age=0")
	}
	// Add remaining custom headers
	alreadyProcessedHeaders := map[string]bool{
		"user-agent":                true,
		"accept":                    true,
		"accept-language":           true,
		"accept-encoding":           true,
		"connection":                true,
		"upgrade-insecure-requests": true,
		"sec-fetch-dest":            true,
		"sec-fetch-mode":            true,
		"sec-fetch-site":            true,
		"cache-control":             true,
	}
	for _, header := range customHeaders {
		if colonIndex := strings.Index(header, ":"); colonIndex != -1 {
			key := strings.ToLower(strings.TrimSpace(header[:colonIndex]))
			if !alreadyProcessedHeaders[key] {
				headers = append(headers, header)
			}
		}
	}
	return headers
}

func setCookies(ctx context.Context, targetURL string, cookies []string) error {
	if len(cookies) == 0 {
		return nil
	}
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return fmt.Errorf("failed to parse URL: %v", err)
	}
	domain := parsedURL.Host
	if !strings.Contains(domain, ".") {
		domain = "." + domain
	}
	for _, cookieStr := range cookies {
		// Parse cookie string (format: "name=value" or "name=value; domain=example.com")
		parts := strings.Split(cookieStr, ";")
		nameValue := strings.TrimSpace(parts[0])
		if !strings.Contains(nameValue, "=") {
			return fmt.Errorf("invalid cookie format: %s (expected 'name=value')", cookieStr)
		}
		nameValueParts := strings.SplitN(nameValue, "=", 2)
		name := strings.TrimSpace(nameValueParts[0])
		value := strings.TrimSpace(nameValueParts[1])
		// Parse additional attributes
		cookieDomain := domain
		path := "/"
		secure := parsedURL.Scheme == "https"
		httpOnly := false
		sameSite := ""
		var expires *time.Time
		var maxAge *int64
		for i := 1; i < len(parts); i++ {
			attr := strings.TrimSpace(parts[i])
			attrLower := strings.ToLower(attr)
			switch {
			case strings.HasPrefix(attrLower, "domain="):
				cookieDomain = strings.TrimSpace(attr[7:])
			case strings.HasPrefix(attrLower, "path="):
				path = strings.TrimSpace(attr[5:])
			case attrLower == "secure":
				secure = true
			case attrLower == "httponly":
				httpOnly = true
			case strings.HasPrefix(attrLower, "samesite="):
				sameSite = strings.TrimSpace(attr[9:])
			case strings.HasPrefix(attrLower, "expires="):
				expiresStr := strings.TrimSpace(attr[8:])
				// Try parsing as RFC1123, if it fails, try as RFC1123Z
				if parsedExpires, err := time.Parse(time.RFC1123, expiresStr); err == nil {
					expires = &parsedExpires
				} else {
					parsedExpires, err := time.Parse(time.RFC1123Z, expiresStr)
					if err == nil {
						expires = &parsedExpires
					} else {
						return fmt.Errorf("invalid expires format: %s", expiresStr)
					}
				}
			case strings.HasPrefix(attrLower, "max-age="):
				if maxAgeVal, err := fmt.Sscanf(strings.TrimSpace(attr[8:]), "%d", &maxAge); err == nil && maxAgeVal == 1 {
					age := int64(0)
					fmt.Sscanf(strings.TrimSpace(attr[8:]), "%d", &age)
					maxAge = &age
				}
			}
		}
		// Build cookie command
		cookieCmd := cdpnetwork.SetCookie(name, value).
			WithDomain(cookieDomain).
			WithPath(path).
			WithSecure(secure).
			WithHTTPOnly(httpOnly)
		// Add SameSite if specified
		if sameSite != "" {
			switch strings.ToLower(sameSite) {
			case "strict":
				cookieCmd = cookieCmd.WithSameSite(cdpnetwork.CookieSameSiteStrict)
			case "lax":
				cookieCmd = cookieCmd.WithSameSite(cdpnetwork.CookieSameSiteLax)
			case "none":
				cookieCmd = cookieCmd.WithSameSite(cdpnetwork.CookieSameSiteNone)
			}
		}
		// Add expiration if specified
		if expires != nil {
			expiresTime := cdp.TimeSinceEpoch(*expires)
			cookieCmd = cookieCmd.WithExpires(&expiresTime)
		}
		// Execute cookie command
		err := chromedp.Run(ctx, cookieCmd)
		if err != nil {
			return fmt.Errorf("failed to set cookie %s: %v", name, err)
		}
	}
	return nil
}

func writeOutput(content string) error {
	if outputFile != "" { // Determine the full file path
		var filePath string
		if outputDir != "" { // Create the output directory if it doesn't exist
			err := os.MkdirAll(outputDir, 0755)
			if err != nil {
				return fmt.Errorf("failed to create output directory: %v", err)
			}
			filePath = filepath.Join(outputDir, outputFile)
		} else {
			filePath = outputFile
		}
		var file *os.File
		var err error
		if appendMode { // Open file in append mode, create if it doesn't exist
			file, err = os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		} else { // Create new file (overwrite existing)
			file, err = os.Create(filePath)
		}
		if err != nil {
			return fmt.Errorf("failed to open output file: %v", err)
		}
		defer file.Close()
		_, err = file.WriteString(content)
		if err != nil {
			return fmt.Errorf("failed to write to output file: %v", err)
		}
		if appendMode {
			fmt.Fprintf(os.Stderr, "Output appended to: %s\n", filePath)
		} else {
			fmt.Fprintf(os.Stderr, "Output written to: %s\n", filePath)
		}
	} else {
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
	rootCmd.Flags().StringArrayVarP(&cookies, "cookie", "C", []string{}, "Add cookies (format: 'name=value' or 'name=value; domain=example.com')")
	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Write to file instead of stdout")
	rootCmd.Flags().StringVar(&outputDir, "output-dir", "", "Directory to save files in")
	rootCmd.Flags().BoolVarP(&appendMode, "append", "a", false, "Append to file instead of overwriting")
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

	// Get initial response to capture redirect information
	initialProtocol, initialStatusCode, err := getInitialResponse(url, userAgent, headers)
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
		if showNetwork {
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
	})
	// Set cookies if provided
	if len(cookies) > 0 {
		err := setCookies(ctx, url, cookies)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to set cookies: %v\n", err)
			os.Exit(1)
		}
	}
	// Navigate to the page
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
		err = writeOutput(string(jsonData) + "\n")
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
