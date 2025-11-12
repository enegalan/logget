package helpers

import "regexp"

type Config struct {
	UserAgent           string
	Headers             []string
	Cookies             []string
	OutputFile          string
	AppendMode          bool
	FollowMode          bool
	SkipSSLVerify       bool
	ShowNetwork         bool
	ShowLogs            bool
	JSONOutput          bool
	YAMLOutput          bool
	FilterPattern       string
	ExcludePattern      string
	StatusPattern       string
	DomainPattern       string
	MimePattern         string
	StatusRegex         *regexp.Regexp
	DomainRegex         *regexp.Regexp
	MimeRegex           *regexp.Regexp
	OutputWriter        *OutputWriter
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
}
