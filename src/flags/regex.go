package flags

import "strings"

type RegexPattern string

func (r *RegexPattern) String() string         { return string(*r) }
func (r *RegexPattern) Set(value string) error { *r = RegexPattern(value); return nil }
func (r *RegexPattern) Type() string           { return "<regex>" }
func (r *RegexPattern) Get() string            { return string(*r) }
func (r *RegexPattern) Empty() bool            { return strings.TrimSpace(string(*r)) == "" }
