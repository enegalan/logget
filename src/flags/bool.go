package flags

type BoolFlag struct {
	SimpleFlag[bool]
	boolPtr *bool
}

func (b *BoolFlag) Type() string { return "" }

func (b *BoolFlag) GetBoolPtr() *bool {
	if b.boolPtr == nil {
		val := b.SimpleFlag.Get()
		b.boolPtr = &val
	}
	return b.boolPtr
}

func (b *BoolFlag) Get() bool {
	if b.boolPtr != nil {
		return *b.boolPtr
	}
	return b.SimpleFlag.Get()
}
