package fmtconf

type (
	JoinConfigStartIndentType string
	JoinConfigLineBreakType   string
)

const (
	JOIN_START_INDENT_TYPE_NONE      JoinConfigStartIndentType = "NONE"
	JOIN_START_INDENT_TYPE_ONE_SPACE JoinConfigStartIndentType = "ONE_SPACE"

	JOIN_LINE_BREAK_OFF       JoinConfigLineBreakType = "OFF"
	JOIN_LINE_BREAK_ON_CLAUSE JoinConfigLineBreakType = "ON_CLAUSE"
)

type JoinConfig struct {
	StartIndentType JoinConfigStartIndentType
	LineBreakType   JoinConfigLineBreakType
}

func (c *Config) WithJoinStartIndentTypeNone() *Config {
	c.Join.StartIndentType = JOIN_START_INDENT_TYPE_NONE
	return c
}

func (c *Config) WithJoinLineBreakOff() *Config {
	c.Join.LineBreakType = JOIN_LINE_BREAK_OFF
	return c
}
