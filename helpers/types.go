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
	JSONOutput     bool
	FilterPattern  string
	ExcludePattern string
	XHROnly        bool
	DocumentOnly   bool
	CssOnly        bool
	ScriptOnly     bool
	FontOnly       bool
	ImgOnly        bool
	MediaOnly      bool
	ManifestOnly   bool
	WebSocketOnly  bool
	WasmOnly       bool
}
