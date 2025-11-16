package flags

type Milliseconds struct {
	SimpleFlag[int]
}

func (m *Milliseconds) Type() string { return "<milliseconds>" }
