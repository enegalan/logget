package helpers

type Config struct {
	UserAgent      string
	Headers        []string
	Cookies        []string
	OutputFile     string
	OutputDir      string
	AppendMode     bool
	SkipSSLVerify  bool
	ShowNetwork    bool
	ShowLogs       bool
	JSONOutput     bool
	FilterPattern  string
	ExcludePattern string
	StatusPattern  string
	DomainPattern  string
	MimePattern    string
	XHROnly        bool
	DocumentOnly   bool
	CssOnly        bool
	ScriptOnly     bool
	FontOnly       bool
	ImgOnly        bool
	MediaOnly      bool
	ManifestOnly   bool
	WebSocketOnly  bool
}
