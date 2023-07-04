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
			StartIndentType: JOIN_START_INDENT_TYPE_ONE_SPACE,
			LineBreakType:   JOIN_LINE_BREAK_ON_CLAUSE,
		},
	}
}
