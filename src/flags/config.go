package flags

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type FlagConfig struct {
	Flag  pflag.Value
	Name  string
	Short string
	Desc  string
	Value interface{}
}

func RegisterFlag(cmd *cobra.Command, cfg FlagConfig) {
	switch v := cfg.Value.(type) {
	case bool:
		if boolFlag, ok := cfg.Flag.(*BoolFlag); ok {
			boolFlag.SetDefault(v)
		}
	case int:
		if intFlag, ok := cfg.Flag.(interface{ SetDefault(int) }); ok {
			intFlag.SetDefault(v)
		}
	case int64:
		if int64Flag, ok := cfg.Flag.(interface{ SetDefault(int64) }); ok {
			int64Flag.SetDefault(v)
		}
	case string:
		if stringFlag, ok := cfg.Flag.(interface{ SetDefault(string) }); ok {
			stringFlag.SetDefault(v)
		}
	case []string:
		if arrayFlag, ok := cfg.Flag.(interface{ SetDefault([]string) }); ok {
			arrayFlag.SetDefault(v)
		}
	}
	if cfg.Short != "" {
		cmd.Flags().VarP(cfg.Flag, cfg.Name, cfg.Short, cfg.Desc)
	} else {
		cmd.Flags().Var(cfg.Flag, cfg.Name, cfg.Desc)
	}
	if _, ok := cfg.Flag.(*BoolFlag); ok {
		if flag := cmd.Flags().Lookup(cfg.Name); flag != nil {
			flag.NoOptDefVal = "true"
		}
	}
}
