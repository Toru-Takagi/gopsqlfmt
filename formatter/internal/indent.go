package internal

import "github.com/Toru-Takagi/gopsqlfmt/fmtconf"

func GetIndent(conf *fmtconf.Config) string {
	switch conf.IndentType {
	case fmtconf.INDENT_TYPE_TAB:
		return "\t"
	case fmtconf.INDENT_TYPE_TWO_SPACES:
		return "  "
	}
	return "	"
}
