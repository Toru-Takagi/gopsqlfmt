package fmtconf

type FuncNameTypeCase string

const (
	FUNC_NAME_TYPE_CASE_LOWER FuncNameTypeCase = "LOWERCASE"
	FUNC_NAME_TYPE_CASE_UPPER FuncNameTypeCase = "UPPERCASE"
)

type FuncCallConfig struct {
	FuncNameTypeCase
}

func (c *Config) WithFuncNameTypeCaseUpper() *Config {
	c.FuncCallConfig.FuncNameTypeCase = FUNC_NAME_TYPE_CASE_UPPER
	return c
}
