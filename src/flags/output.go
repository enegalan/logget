package flags

import "strings"

type OutputFile string

func (o *OutputFile) String() string {
	return string(*o)
}

func (o *OutputFile) Set(value string) error {
	*o = OutputFile(value)
	return nil
}

func (o *OutputFile) Type() string {
	return "<file>"
}

func (o *OutputFile) Get() string {
	return string(*o)
}

func (o *OutputFile) Empty() bool {
	return strings.TrimSpace(string(*o)) == ""
}
