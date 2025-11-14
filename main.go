package main

import (
	"io"
	"log"
	helpers "logget/src"
	"logget/src/flags"
	"os"

	"github.com/spf13/cobra"
)

var (
	showLogs    bool
	showNetwork bool
	jsonOutput  bool
	yamlOutput  bool
	csvOutput   bool
	timeout     flags.Milliseconds = 60000
	wait        flags.Milliseconds = 100
	userAgent   flags.UserAgent    = "logget/1.0"
	headers     flags.HeaderArray
	cookies     flags.CookieArray
	versionFlag bool
	verbose     bool
	outputFile  flags.OutputFile
	appendMode  bool
	version     string = "dev"

	followMode      bool
	filterPattern   flags.RegexPattern
	excludePattern  flags.RegexPattern
	statusPattern   flags.RegexPattern
	domainPattern   flags.RegexPattern
	mimePattern     flags.RegexPattern
	refreshInterval flags.Milliseconds = 100

	skipSSLVerify        bool
	noRotateFingerprints bool
	fingerprintInterval  flags.Milliseconds = 5000
	harOutput            bool

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
	var rootCmd = &cobra.Command{
		Use:          "logget [flags] <url>",
		Short:        "Extract logs and network data from web pages",
		Args:         cobra.ArbitraryArgs,
		Run:          runLogget,
		SilenceUsage: true,
	}
	rootCmd.Flags().BoolVarP(&showLogs, "logs", "L", false, "Show console logs")
	rootCmd.Flags().BoolVarP(&showNetwork, "network", "N", false, "Show network requests")
	rootCmd.Flags().BoolVarP(&jsonOutput, "json", "J", false, "Output in JSON format")
	rootCmd.Flags().BoolVar(&yamlOutput, "yaml", false, "Output in YAML format")
	rootCmd.Flags().BoolVar(&csvOutput, "csv", false, "Output in CSV format")
	rootCmd.Flags().VarP(&timeout, "timeout", "T", "Timeout in milliseconds")
	rootCmd.Flags().VarP(&wait, "wait", "W", "Wait time in milliseconds after page load")
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
	rootCmd.Flags().Var(&refreshInterval, "refresh", "Refresh interval in milliseconds for real-time streaming")
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
	rootCmd.Flags().BoolVar(&noRotateFingerprints, "no-rotate-fingerprints", false, "Disable fingerprint rotation (default: enabled)")
	rootCmd.Flags().Var(&fingerprintInterval, "fingerprint-interval", "Interval in milliseconds for fingerprint rotation")
	rootCmd.Flags().BoolVar(&harOutput, "har", false, "Output in HAR (HTTP Archive) format")
	rootCmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		os.Stderr.WriteString("logget: " + helpers.FormatCobraError(err) + "\n")
		os.Exit(1)
		return err
	})
	if err := rootCmd.Execute(); err != nil {
		os.Stderr.WriteString("logget: " + helpers.FormatCobraError(err) + "\n")
		os.Exit(1)
	}
}

func runLogget(cmd *cobra.Command, args []string) {
	if len(args) > 0 && args[0] == "-" {
		os.Stderr.WriteString("logget: " + helpers.FormatUnknownFlag("", true) + "\n")
		os.Exit(1)
	}
	url := ""
	if len(args) > 0 {
		url = args[0]
	}
	cmdConfig := helpers.CommandConfig{
		ShowLogs:             showLogs,
		ShowNetwork:          showNetwork,
		JSONOutput:           jsonOutput,
		YAMLOutput:           yamlOutput,
		CSVOutput:            csvOutput,
		Timeout:              timeout.Get(),
		Wait:                 wait.Get(),
		UserAgent:            userAgent.Get(),
		Headers:              []string(headers),
		Cookies:              []string(cookies),
		VersionFlag:          versionFlag,
		Verbose:              verbose,
		OutputFile:           outputFile.Get(),
		AppendMode:           appendMode,
		Version:              version,
		FollowMode:           followMode,
		FilterPattern:        filterPattern.Get(),
		ExcludePattern:       excludePattern.Get(),
		StatusPattern:        statusPattern.Get(),
		DomainPattern:        domainPattern.Get(),
		MimePattern:          mimePattern.Get(),
		RefreshInterval:      refreshInterval.Get(),
		SkipSSLVerify:        skipSSLVerify,
		NoRotateFingerprints: noRotateFingerprints,
		FingerprintInterval:  fingerprintInterval.Get(),
		HAROutput:            harOutput,
		XHROnly:              xhrOnly,
		DocumentOnly:         documentOnly,
		CssOnly:              cssOnly,
		ScriptOnly:           scriptOnly,
		FontOnly:             fontOnly,
		ImgOnly:              imgOnly,
		MediaOnly:            mediaOnly,
		ManifestOnly:         manifestOnly,
		WebSocketOnly:        websocketOnly,
		NoColor:              noColor,
		Quiet:                quiet,
		MinSize:              minSize.Get(),
		MaxSize:              maxSize.Get(),
	}
	helpers.RunLogget(cmdConfig, url)
}
