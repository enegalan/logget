package flags

type SizeBytes struct {
	SimpleFlag[int64]
}

func (s *SizeBytes) Type() string { return "<bytes>" }
