package flags

type UserAgent struct {
	SimpleFlag[string]
}

func (u *UserAgent) Type() string { return "<name>" }
