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
	showLogs.SetDefault(false)
	showNetwork.SetDefault(false)
	jsonOutput.SetDefault(false)
	yamlOutput.SetDefault(false)
	csvOutput.SetDefault(false)
	timeout.SetDefault(60000)
	wait.SetDefault(100)
	userAgent.SetDefault("logget/1.0")
	headers.SetDefault([]string{})
	cookies.SetDefault([]string{})
	versionFlag.SetDefault(false)
	verbose.SetDefault(false)
	outputFile.SetDefault("")
	appendMode.SetDefault(false)
	refreshInterval.SetDefault(100)
	followMode.SetDefault(false)
	filterPattern.SetDefault("")
	excludePattern.SetDefault("")
	statusPattern.SetDefault("")
	domainPattern.SetDefault("")
	mimePattern.SetDefault("")
	skipSSLVerify.SetDefault(false)
	noRotateFingerprints.SetDefault(false)
	fingerprintInterval.SetDefault(5000)
	harOutput.SetDefault(false)
	xhrOnly.SetDefault(false)
	documentOnly.SetDefault(false)
	cssOnly.SetDefault(false)
	scriptOnly.SetDefault(false)
	fontOnly.SetDefault(false)
	imgOnly.SetDefault(false)
	mediaOnly.SetDefault(false)
	manifestOnly.SetDefault(false)
	websocketOnly.SetDefault(false)
	noColor.SetDefault(false)
	quiet.SetDefault(false)
	minSize.SetDefault(0)
	maxSize.SetDefault(0)
	cmd.Flags().VarP(&showLogs, "logs", "L", "Show console logs")
	cmd.Flags().VarP(&showNetwork, "network", "N", "Show network requests")
	cmd.Flags().VarP(&jsonOutput, "json", "J", "Output in JSON format")
	cmd.Flags().Var(&yamlOutput, "yaml", "Output in YAML format")
	cmd.Flags().Var(&csvOutput, "csv", "Output in CSV format")
	cmd.Flags().VarP(&timeout, "timeout", "T", "Timeout in milliseconds")
	cmd.Flags().VarP(&wait, "wait", "W", "Wait time in milliseconds after page load")
	cmd.Flags().VarP(&userAgent, "user-agent", "A", "Set User-Agent header")
	cmd.Flags().VarP(&headers, "header", "H", "Add custom headers (format: 'Key: Value') or filename containing headers")
	cmd.Flags().VarP(&cookies, "cookie", "C", "Add cookies (format: 'name=value' or 'name=value; domain=example.com') or filename containing cookies")
	cmd.Flags().VarP(&outputFile, "output", "o", "Write to file instead of stdout")
	cmd.Flags().VarP(&appendMode, "append", "a", "Append to file instead of overwriting")
	cmd.Flags().VarP(&versionFlag, "version", "v", "Show version information")
	cmd.Flags().VarP(&verbose, "verbose", "V", "Show detailed HTTP protocol information")
	cmd.Flags().VarP(&followMode, "follow", "f", "Stream logs and network requests in real-time")
	cmd.Flags().VarP(&filterPattern, "filter", "", "Show only logs/requests matching this regex pattern")
	cmd.Flags().VarP(&excludePattern, "exclude", "", "Exclude logs/requests matching this regex pattern")
	cmd.Flags().VarP(&statusPattern, "status", "", "Only include requests whose HTTP status code matches this regex pattern")
	cmd.Flags().VarP(&domainPattern, "domain", "", "Only include requests whose domain matches this regex pattern")
	cmd.Flags().VarP(&mimePattern, "mime", "", "Only include requests whose MIME type matches this regex pattern")
	cmd.Flags().Var(&refreshInterval, "refresh", "Refresh interval in milliseconds for real-time streaming")
	cmd.Flags().VarP(&skipSSLVerify, "insecure", "k", "Skip SSL certificate verification (useful for self-signed certificates)")
	cmd.Flags().Var(&noColor, "no-color", "Disable colored output")
	cmd.Flags().Var(&xhrOnly, "xhr", "Only include fetch/XHR requests")
	cmd.Flags().Var(&documentOnly, "document", "Only include Document requests")
	cmd.Flags().Var(&cssOnly, "css", "Only include CSS requests")
	cmd.Flags().Var(&scriptOnly, "script", "Only include Script requests")
	cmd.Flags().Var(&fontOnly, "font", "Only include Font requests")
	cmd.Flags().Var(&imgOnly, "img", "Only include Image requests")
	cmd.Flags().Var(&mediaOnly, "media", "Only include Media requests")
	cmd.Flags().Var(&manifestOnly, "manifest", "Only include Manifest requests")
	cmd.Flags().Var(&websocketOnly, "socket", "Only include WebSocket requests")
	cmd.Flags().VarP(&minSize, "min-size", "", "Only include requests whose size is at least this many bytes")
	cmd.Flags().VarP(&maxSize, "max-size", "", "Only include requests whose size is at most this many bytes")
	cmd.Flags().VarP(&quiet, "quiet", "q", "Suppress progress messages, only show data")
	cmd.Flags().Var(&noRotateFingerprints, "no-rotate-fingerprints", "Disable fingerprint rotation (default: enabled)")
	cmd.Flags().Var(&fingerprintInterval, "fingerprint-interval", "Interval in milliseconds for fingerprint rotation")
	cmd.Flags().Var(&harOutput, "har", "Output in HAR (HTTP Archive) format")
}

func runLogget(cmd *cobra.Command, args []string) {
	if len(args) > 0 && args[0] == "-" {
		os.Stderr.WriteString("logget: " + command.FormatUnknownFlag("", true) + "\n")
		os.Exit(1)
	}
	url := ""
	if len(args) > 0 {
		url = args[0]
	}
	cmdConfig := command.CommandConfig{
		ShowLogs:             showLogs.Get(),
		ShowNetwork:          showNetwork.Get(),
		JSONOutput:           jsonOutput.Get(),
		YAMLOutput:           yamlOutput.Get(),
		CSVOutput:            csvOutput.Get(),
		Timeout:              timeout.Get(),
		Wait:                 wait.Get(),
		UserAgent:            userAgent.Get(),
		Headers:              headers.Get(),
		Cookies:              cookies.Get(),
		VersionFlag:          versionFlag.Get(),
		Verbose:              verbose.Get(),
		OutputFile:           outputFile.Get(),
		AppendMode:           appendMode.Get(),
		Version:              version,
		FollowMode:           followMode.Get(),
		FilterPattern:        filterPattern.Get(),
		ExcludePattern:       excludePattern.Get(),
		StatusPattern:        statusPattern.Get(),
		DomainPattern:        domainPattern.Get(),
		MimePattern:          mimePattern.Get(),
		RefreshInterval:      refreshInterval.Get(),
		SkipSSLVerify:        skipSSLVerify.Get(),
		NoRotateFingerprints: noRotateFingerprints.Get(),
		FingerprintInterval:  fingerprintInterval.Get(),
		HAROutput:            harOutput.Get(),
		XHROnly:              xhrOnly.Get(),
		DocumentOnly:         documentOnly.Get(),
		CssOnly:              cssOnly.Get(),
		ScriptOnly:           scriptOnly.Get(),
		FontOnly:             fontOnly.Get(),
		ImgOnly:              imgOnly.Get(),
		MediaOnly:            mediaOnly.Get(),
		ManifestOnly:         manifestOnly.Get(),
		WebSocketOnly:        websocketOnly.Get(),
		NoColor:              noColor.Get(),
		Quiet:                quiet.Get(),
		MinSize:              minSize.Get(),
		MaxSize:              maxSize.Get(),
	}
	command.RunLogget(cmdConfig, url)
}
