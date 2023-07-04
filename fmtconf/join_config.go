package fmtconf

type JoinConfigLineBreakType string

const (
	JOIN_LINE_BREAK_ON_CLAUSE JoinConfigLineBreakType = "ON_CLAUSE"
	JOIN_LINE_BREAK_OFF       JoinConfigLineBreakType = "OFF"
)

type JoinConfig struct {
	LineBreakType JoinConfigLineBreakType
}

func (c *Config) WithJoinLineBreakOff() *Config {
	c.Join.LineBreakType = JOIN_LINE_BREAK_OFF
	return c
}
