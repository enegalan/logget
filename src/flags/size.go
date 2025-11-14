package flags

import (
	"fmt"
	"strconv"
)

type SizeBytes int64

func (s *SizeBytes) String() string { return fmt.Sprintf("%d", *s) }
func (s *SizeBytes) Set(value string) error {
	size, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid size value: %v", err)
	}
	*s = SizeBytes(size)
	return nil
}
func (s *SizeBytes) Type() string { return "<bytes>" }
func (s *SizeBytes) Get() int64   { return int64(*s) }
func (s *SizeBytes) Empty() bool  { return *s == 0 }
