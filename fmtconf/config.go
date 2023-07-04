package fmtconf

type Config struct {
	Join JoinConfig
}

type JoinConfigLineBreakType string

const (
	JOIN_LINE_BREAK_ON_CLAUSE JoinConfigLineBreakType = "ON_CLAUSE"
	JOIN_LINE_BREAK_OFF       JoinConfigLineBreakType = "OFF"
)

type JoinConfig struct {
	LineBreakType JoinConfigLineBreakType
}

func NewDefaultConfig() *Config {
	return &Config{
		Join: JoinConfig{
			LineBreakType: JOIN_LINE_BREAK_ON_CLAUSE,
		},
	}
}

func (c *Config) WithJoinLineBreakOff() *Config {
	c.Join.LineBreakType = JOIN_LINE_BREAK_OFF
	return c
}
