package fmtconf

type IndentType string

const (
	INDENT_TYPE_TAB        IndentType = "TAB"
	INDENT_TYPE_TWO_SPACES IndentType = "TWO_SPACES"
)

type Config struct {
	IndentType     IndentType
	FuncCallConfig FuncCallConfig
	Join           JoinConfig
}

func NewDefaultConfig() *Config {
	return &Config{
		IndentType: INDENT_TYPE_TWO_SPACES,
		FuncCallConfig: FuncCallConfig{
			FuncNameTypeCase: FUNC_NAME_TYPE_CASE_LOWER,
		},
		Join: JoinConfig{
			StartIndentType: JOIN_START_INDENT_TYPE_ONE_SPACE,
			LineBreakType:   JOIN_LINE_BREAK_ON_CLAUSE,
		},
	}
}

func (c *Config) WithIndentTypeTab() *Config {
	c.IndentType = INDENT_TYPE_TAB
	return c
}
