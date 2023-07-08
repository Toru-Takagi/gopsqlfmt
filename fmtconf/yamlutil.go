package fmtconf

import (
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

type YamlFuncSettings struct {
	NameTypeCase FuncNameTypeCase `yaml:"name-type-case"`
}

type YamlJoinSettings struct {
	StartIndentType JoinConfigStartIndentType `yaml:"start-indent-type"`
	LineBreakType   JoinConfigLineBreakType   `yaml:"line-break-type"`
}

type YamlFormatSettings struct {
	IndentType IndentType       `yaml:"indent-type"`
	Func       YamlFuncSettings `yaml:"func"`
	Join       YamlJoinSettings `yaml:"join"`
}

type YamlConfig struct {
	FormatSettings YamlFormatSettings `yaml:"format-settings"`
}

func LoadYamlConfig() (*Config, error) {
	conf := NewDefaultConfig()

	if data, err := ioutil.ReadFile(".gopsqlfmt.yaml"); err == nil {
		var ymlconf YamlConfig
		if err := yaml.Unmarshal(data, &ymlconf); err == nil {
			switch ymlconf.FormatSettings.IndentType {
			case INDENT_TYPE_TAB:
				conf.IndentType = INDENT_TYPE_TAB
			}

			switch ymlconf.FormatSettings.Func.NameTypeCase {
			case FUNC_NAME_TYPE_CASE_UPPER:
				conf.FuncCallConfig.FuncNameTypeCase = FUNC_NAME_TYPE_CASE_UPPER
			}

			switch ymlconf.FormatSettings.Join.StartIndentType {
			case JOIN_START_INDENT_TYPE_NONE:
				conf.Join.StartIndentType = JOIN_START_INDENT_TYPE_NONE
			}

			switch ymlconf.FormatSettings.Join.LineBreakType {
			case JOIN_LINE_BREAK_OFF:
				conf.Join.LineBreakType = JOIN_LINE_BREAK_OFF
			}
		}
	}

	return conf, nil
}
