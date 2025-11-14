package flags

import (
	"fmt"
	"strconv"
)

type Milliseconds int

func (m *Milliseconds) String() string { return fmt.Sprintf("%d", *m) }
func (m *Milliseconds) Set(value string) error {
	ms, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid milliseconds value: %v", err)
	}
	*m = Milliseconds(ms)
	return nil
}
func (m *Milliseconds) Type() string { return "<milliseconds>" }
func (m *Milliseconds) Get() int     { return int(*m) }
func (m *Milliseconds) Empty() bool  { return *m == 0 }
