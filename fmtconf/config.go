package fmtconf

type Config struct {
	FuncCallConfig FuncCallConfig
	Join           JoinConfig
}

func NewDefaultConfig() *Config {
	return &Config{
		FuncCallConfig: FuncCallConfig{
			FuncNameTypeCase: FUNC_NAME_TYPE_CASE_LOWER,
		},
		Join: JoinConfig{
			LineBreakType: JOIN_LINE_BREAK_ON_CLAUSE,
		},
	}
}
