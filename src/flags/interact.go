package flags

type InteractArray struct {
	SimpleFlag[[]string]
}

func (i *InteractArray) Type() string { return "<action:value>" }

func (i *InteractArray) Set(value string) error {
	if i.Value == nil {
		i.Value = []string{}
	}
	i.Value = append(i.Value, value)
	return nil
}

func (i *InteractArray) Get() []string {
	if i.Value == nil {
		return []string{}
	}
	return i.Value
}
