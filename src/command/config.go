package command

import (
	"fmt"
	"os"
	"regexp"

	helpers "logget/src"
)

type CommandConfig struct {
	ShowLogs             bool
	ShowNetwork          bool
	JSONOutput           bool
	YAMLOutput           bool
	CSVOutput            bool
	Timeout              int
	Wait                 int
	UserAgent            string
	Headers              []string
	Cookies              []string
	VersionFlag          bool
	Verbose              bool
	OutputFile           string
	AppendMode           bool
	Version              string
	FollowMode           bool
	FilterPattern        string
	ExcludePattern       string
	StatusPattern        string
	DomainPattern        string
	MimePattern          string
	RefreshInterval      int
	SkipSSLVerify        bool
	NoRotateFingerprints bool
	FingerprintInterval  int
	HAROutput            bool
	XHROnly              bool
	DocumentOnly         bool
	CssOnly              bool
	ScriptOnly           bool
	FontOnly             bool
	ImgOnly              bool
	MediaOnly            bool
	ManifestOnly         bool
	WebSocketOnly        bool
	NoColor              bool
	Quiet                bool
	MinSize              int64
	MaxSize              int64
}

func buildConfig(cmdConfig CommandConfig) helpers.Config {
	cfg := helpers.Config{
		UserAgent:           cmdConfig.UserAgent,
		Headers:             cmdConfig.Headers,
		Cookies:             cmdConfig.Cookies,
		OutputFile:          cmdConfig.OutputFile,
		AppendMode:          cmdConfig.AppendMode,
		FollowMode:          cmdConfig.FollowMode,
		SkipSSLVerify:       cmdConfig.SkipSSLVerify,
		ShowNetwork:         cmdConfig.ShowNetwork,
		ShowLogs:            cmdConfig.ShowLogs,
		JSONOutput:          cmdConfig.JSONOutput,
		YAMLOutput:          cmdConfig.YAMLOutput,
		FilterPattern:       cmdConfig.FilterPattern,
		ExcludePattern:      cmdConfig.ExcludePattern,
		StatusPattern:       cmdConfig.StatusPattern,
		DomainPattern:       cmdConfig.DomainPattern,
		MimePattern:         cmdConfig.MimePattern,
		XHROnly:             cmdConfig.XHROnly,
		DocumentOnly:        cmdConfig.DocumentOnly,
		CssOnly:             cmdConfig.CssOnly,
		ScriptOnly:          cmdConfig.ScriptOnly,
		FontOnly:            cmdConfig.FontOnly,
		ImgOnly:             cmdConfig.ImgOnly,
		MediaOnly:           cmdConfig.MediaOnly,
		ManifestOnly:        cmdConfig.ManifestOnly,
		WebSocketOnly:       cmdConfig.WebSocketOnly,
		MinSize:             cmdConfig.MinSize,
		MaxSize:             cmdConfig.MaxSize,
		RotateFingerprints:  !cmdConfig.NoRotateFingerprints,
		FingerprintInterval: cmdConfig.FingerprintInterval,
		HAROutput:           cmdConfig.HAROutput,
	}
	if cmdConfig.StatusPattern != "" {
		if r, err := regexp.Compile(cmdConfig.StatusPattern); err == nil {
			cfg.StatusRegex = r
		}
	}
	if cmdConfig.DomainPattern != "" {
		if r, err := regexp.Compile(cmdConfig.DomainPattern); err == nil {
			cfg.DomainRegex = r
		}
	}
	if cmdConfig.MimePattern != "" {
		if r, err := regexp.Compile(cmdConfig.MimePattern); err == nil {
			cfg.MimeRegex = r
		}
	}
	return cfg
}

func validateOutputFormats(cmdConfig CommandConfig) {
	formatCount := 0
	for _, f := range []bool{cmdConfig.JSONOutput, cmdConfig.YAMLOutput, cmdConfig.CSVOutput, cmdConfig.HAROutput} {
		if f {
			formatCount++
		}
	}
	if formatCount > 1 {
		fmt.Println("logget: Only one output format can be specified at a time")
		os.Exit(1)
	}
}
