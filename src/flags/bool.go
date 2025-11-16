package flags

type BoolFlag struct {
	SimpleFlag[bool]
}

func (b *BoolFlag) Type() string { return "<bool>" }
