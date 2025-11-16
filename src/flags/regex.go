package flags

type RegexPattern struct {
	SimpleFlag[string]
}

func (r *RegexPattern) Type() string { return "<regex>" }
