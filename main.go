package main

import (
	"io"
	"log"
	"logget/src/command"
	"logget/src/flags"
	"os"

	"github.com/spf13/cobra"
)

var (
	showLogs    flags.BoolFlag
	showNetwork flags.BoolFlag
	jsonOutput  flags.BoolFlag
	yamlOutput  flags.BoolFlag
	csvOutput   flags.BoolFlag
	timeout     flags.Milliseconds
	wait        flags.Milliseconds
	userAgent   flags.UserAgent
	headers     flags.HeaderArray
	cookies     flags.CookieArray
	versionFlag flags.BoolFlag
	verbose     flags.BoolFlag
	outputFile  flags.OutputFile
	appendMode  flags.BoolFlag
	version     string = "dev"

	followMode      flags.BoolFlag
	filterPattern   flags.RegexPattern
	excludePattern  flags.RegexPattern
	statusPattern   flags.RegexPattern
	domainPattern   flags.RegexPattern
	mimePattern     flags.RegexPattern
	refreshInterval flags.Milliseconds

	skipSSLVerify        flags.BoolFlag
	noRotateFingerprints flags.BoolFlag
	fingerprintInterval  flags.Milliseconds
	harOutput            flags.BoolFlag

	xhrOnly       flags.BoolFlag
	documentOnly  flags.BoolFlag
	cssOnly       flags.BoolFlag
	scriptOnly    flags.BoolFlag
	fontOnly      flags.BoolFlag
	imgOnly       flags.BoolFlag
	mediaOnly     flags.BoolFlag
	manifestOnly  flags.BoolFlag
	websocketOnly flags.BoolFlag
	noColor       flags.BoolFlag
	quiet         flags.BoolFlag
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
	initFlags(rootCmd)
	rootCmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		os.Stderr.WriteString("logget: " + command.FormatCobraError(err) + "\n")
		os.Exit(1)
		return err
	})
	if err := rootCmd.Execute(); err != nil {
		os.Stderr.WriteString("logget: " + command.FormatCobraError(err) + "\n")
		os.Exit(1)
	}
}

func initFlags(cmd *cobra.Command) {
	allFlags := []flags.FlagConfig{
		{Flag: &showLogs, Name: "logs", Short: "L", Desc: "Show console logs", Value: false},
		{Flag: &showNetwork, Name: "network", Short: "N", Desc: "Show network requests", Value: false},
		{Flag: &jsonOutput, Name: "json", Short: "J", Desc: "Output in JSON format", Value: false},
		{Flag: &yamlOutput, Name: "yaml", Short: "", Desc: "Output in YAML format", Value: false},
		{Flag: &csvOutput, Name: "csv", Short: "", Desc: "Output in CSV format", Value: false},
		{Flag: &versionFlag, Name: "version", Short: "v", Desc: "Show version information", Value: false},
		{Flag: &verbose, Name: "verbose", Short: "V", Desc: "Show detailed HTTP protocol information", Value: false},
		{Flag: &appendMode, Name: "append", Short: "a", Desc: "Append to file instead of overwriting", Value: false},
		{Flag: &followMode, Name: "follow", Short: "f", Desc: "Stream logs and network requests in real-time", Value: false},
		{Flag: &skipSSLVerify, Name: "insecure", Short: "k", Desc: "Skip SSL certificate verification (useful for self-signed certificates)", Value: false},
		{Flag: &noColor, Name: "no-color", Short: "", Desc: "Disable colored output", Value: false},
		{Flag: &xhrOnly, Name: "xhr", Short: "", Desc: "Only include fetch/XHR requests", Value: false},
		{Flag: &documentOnly, Name: "document", Short: "", Desc: "Only include Document requests", Value: false},
		{Flag: &cssOnly, Name: "css", Short: "", Desc: "Only include CSS requests", Value: false},
		{Flag: &scriptOnly, Name: "script", Short: "", Desc: "Only include Script requests", Value: false},
		{Flag: &fontOnly, Name: "font", Short: "", Desc: "Only include Font requests", Value: false},
		{Flag: &imgOnly, Name: "img", Short: "", Desc: "Only include Image requests", Value: false},
		{Flag: &mediaOnly, Name: "media", Short: "", Desc: "Only include Media requests", Value: false},
		{Flag: &manifestOnly, Name: "manifest", Short: "", Desc: "Only include Manifest requests", Value: false},
		{Flag: &websocketOnly, Name: "socket", Short: "", Desc: "Only include WebSocket requests", Value: false},
		{Flag: &quiet, Name: "quiet", Short: "q", Desc: "Suppress progress messages, only show data", Value: false},
		{Flag: &noRotateFingerprints, Name: "no-rotate-fingerprints", Short: "", Desc: "Disable fingerprint rotation (default: enabled)", Value: false},
		{Flag: &harOutput, Name: "har", Short: "", Desc: "Output in HAR (HTTP Archive) format", Value: false},
		{Flag: &timeout, Name: "timeout", Short: "T", Desc: "Timeout in milliseconds", Value: 60000},
		{Flag: &wait, Name: "wait", Short: "W", Desc: "Wait time in milliseconds after page load", Value: 100},
		{Flag: &refreshInterval, Name: "refresh", Short: "", Desc: "Refresh interval in milliseconds for real-time streaming", Value: 100},
		{Flag: &fingerprintInterval, Name: "fingerprint-interval", Short: "", Desc: "Interval in milliseconds for fingerprint rotation", Value: 5000},
		{Flag: &userAgent, Name: "user-agent", Short: "A", Desc: "Set User-Agent header", Value: "logget/1.0"},
		{Flag: &outputFile, Name: "output", Short: "o", Desc: "Write to file instead of stdout", Value: ""},
		{Flag: &filterPattern, Name: "filter", Short: "", Desc: "Show only logs/requests matching this regex pattern", Value: ""},
		{Flag: &excludePattern, Name: "exclude", Short: "", Desc: "Exclude logs/requests matching this regex pattern", Value: ""},
		{Flag: &statusPattern, Name: "status", Short: "", Desc: "Only include requests whose HTTP status code matches this regex pattern", Value: ""},
		{Flag: &domainPattern, Name: "domain", Short: "", Desc: "Only include requests whose domain matches this regex pattern", Value: ""},
		{Flag: &mimePattern, Name: "mime", Short: "", Desc: "Only include requests whose MIME type matches this regex pattern", Value: ""},
		{Flag: &headers, Name: "header", Short: "H", Desc: "Add custom headers (format: 'Key: Value') or filename containing headers", Value: []string{}},
		{Flag: &cookies, Name: "cookie", Short: "C", Desc: "Add cookies (format: 'name=value' or 'name=value; domain=example.com') or filename containing cookies", Value: []string{}},
		{Flag: &minSize, Name: "min-size", Short: "", Desc: "Only include requests whose size is at least this many bytes", Value: int64(0)},
		{Flag: &maxSize, Name: "max-size", Short: "", Desc: "Only include requests whose size is at most this many bytes", Value: int64(0)},
	}
	for _, f := range allFlags {
		flags.RegisterFlag(cmd, f)
	}
}

func runLogget(cmd *cobra.Command, args []string) {
	if len(args) > 0 && args[0] == "-" {
		os.Stderr.WriteString("logget: " + flags.FormatUnknownFlag("", true) + "\n")
		os.Exit(1)
	}
	url := ""
	if len(args) > 0 {
		url = args[0]
	}
	cfg := command.Config{
		ShowLogs:            showLogs.Get(),
		ShowNetwork:         showNetwork.Get(),
		JSONOutput:          jsonOutput.Get(),
		YAMLOutput:          yamlOutput.Get(),
		CSVOutput:           csvOutput.Get(),
		Timeout:             timeout.Get(),
		Wait:                wait.Get(),
		UserAgent:           userAgent.Get(),
		Headers:             headers.Get(),
		Cookies:             cookies.Get(),
		VersionFlag:         versionFlag.Get(),
		Verbose:             verbose.Get(),
		OutputFile:          outputFile.Get(),
		AppendMode:          appendMode.Get(),
		Version:             version,
		FollowMode:          followMode.Get(),
		FilterPattern:       filterPattern.Get(),
		ExcludePattern:      excludePattern.Get(),
		StatusPattern:       statusPattern.Get(),
		DomainPattern:       domainPattern.Get(),
		MimePattern:         mimePattern.Get(),
		RefreshInterval:     refreshInterval.Get(),
		SkipSSLVerify:       skipSSLVerify.Get(),
		RotateFingerprints:  !noRotateFingerprints.Get(),
		FingerprintInterval: fingerprintInterval.Get(),
		HAROutput:           harOutput.Get(),
		XHROnly:             xhrOnly.Get(),
		DocumentOnly:        documentOnly.Get(),
		CssOnly:             cssOnly.Get(),
		ScriptOnly:          scriptOnly.Get(),
		FontOnly:            fontOnly.Get(),
		ImgOnly:             imgOnly.Get(),
		MediaOnly:           mediaOnly.Get(),
		ManifestOnly:        manifestOnly.Get(),
		WebSocketOnly:       websocketOnly.Get(),
		NoColor:             noColor.Get(),
		Quiet:               quiet.Get(),
		MinSize:             minSize.Get(),
		MaxSize:             maxSize.Get(),
	}
	command.RunLogget(cfg, url)
}
