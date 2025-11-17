package command

import (
	helpers "logget/src/helpers"
	"regexp"
)

type Config struct {
	UserAgent      string
	Headers        []string
	Cookies        []string
	OutputFile     string
	AppendMode     bool
	FollowMode     bool
	SkipSSLVerify  bool
	ShowNetwork    bool
	ShowLogs       bool
	JSONOutput     bool
	YAMLOutput     bool
	CSVOutput      bool
	FilterPattern  string
	ExcludePattern string
	StatusPattern  string
	DomainPattern  string
	MimePattern    string
	StatusRegex    *regexp.Regexp
	DomainRegex    *regexp.Regexp
	MimeRegex      *regexp.Regexp
	OutputWriter   interface {
		Write(content string) error
		Close() error
	}
	XHROnly             bool
	DocumentOnly        bool
	CssOnly             bool
	ScriptOnly          bool
	FontOnly            bool
	ImgOnly             bool
	MediaOnly           bool
	ManifestOnly        bool
	WebSocketOnly       bool
	MinSize             int64
	MaxSize             int64
	RotateFingerprints  bool
	FingerprintInterval int
	HAROutput           bool
	Timeout             int
	Wait                int
	VersionFlag         bool
	Verbose             bool
	Version             string
	RefreshInterval     int
	NoColor             bool
	Quiet               bool
}

func compileConfigRegexp(cfg Config) Config {
	if cfg.StatusPattern != "" {
		cfg.StatusRegex = helpers.CompileRegexPattern(cfg.StatusPattern)
	}
	if cfg.DomainPattern != "" {
		cfg.DomainRegex = helpers.CompileRegexPattern(cfg.DomainPattern)
	}
	if cfg.MimePattern != "" {
		cfg.MimeRegex = helpers.CompileRegexPattern(cfg.MimePattern)
	}
	return cfg
}
