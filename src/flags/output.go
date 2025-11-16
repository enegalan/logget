package flags

type OutputFile struct {
	SimpleFlag[string]
}

func (o *OutputFile) Type() string { return "<file>" }
