package flags

import "strings"

type UserAgent string

func (u *UserAgent) String() string {
	return string(*u)
}

func (u *UserAgent) Set(value string) error {
	*u = UserAgent(value)
	return nil
}

func (u *UserAgent) Type() string {
	return "<name>"
}

func (u *UserAgent) Get() string {
	return string(*u)
}

func (u *UserAgent) Empty() bool {
	return strings.TrimSpace(string(*u)) == ""
}
