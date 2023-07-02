package nodeformatter

import (
	"context"
	"strings"

	pg_query "github.com/pganalyze/pg_query_go/v4"
)

func FormatFuncname(ctx context.Context, funcCall *pg_query.Node_FuncCall) (string, error) {
	var bu strings.Builder

	for _, name := range funcCall.FuncCall.Funcname {
		if s, ok := name.Node.(*pg_query.Node_String_); ok {
			switch s.String_.Sval {
			case "now":
				bu.WriteString("NOW")
				bu.WriteString("(")
			case "count":
				bu.WriteString("COUNT")
				bu.WriteString("(")
				bu.WriteString("*")
			case "min":
				bu.WriteString("MIN")
				bu.WriteString("(")
			case "gen_random_uuid":
				bu.WriteString("GEN_RANDOM_UUID")
				bu.WriteString("(")
			case "current_setting":
				bu.WriteString("CURRENT_SETTING")
				bu.WriteString("(")
			case "set_config":
				bu.WriteString("SET_CONFIG")
				bu.WriteString("(")
			case "array_agg":
				bu.WriteString("ARRAY_AGG")
				bu.WriteString("(")
			}
		}
	}

	return bu.String(), nil
}
