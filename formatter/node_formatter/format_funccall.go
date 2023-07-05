package nodeformatter

import (
	"context"
	"fmt"
	"strings"

	"github.com/Toru-Takagi/gopsqlfmt/fmtconf"
	pg_query "github.com/pganalyze/pg_query_go/v4"
)

func FormatFuncname(ctx context.Context, funcCall *pg_query.Node_FuncCall, conf *fmtconf.Config) (string, error) {
	var bu strings.Builder

	for _, name := range funcCall.FuncCall.Funcname {
		if s, ok := name.Node.(*pg_query.Node_String_); ok {
			switch s.String_.Sval {
			case "now":
				// https://www.postgresql.org/docs/15/functions-datetime.html
				bu.WriteString(convertFuncNameTypeCase("now", "NOW", conf))
				bu.WriteString("(")
			case "count":
				// https://www.postgresql.org/docs/15/functions-aggregate.html
				bu.WriteString(convertFuncNameTypeCase("count", "COUNT", conf))
				bu.WriteString("(")
				bu.WriteString("*")
			case "min":
				// https://www.postgresql.org/docs/15/functions-aggregate.html
				bu.WriteString(convertFuncNameTypeCase("min", "MIN", conf))
				bu.WriteString("(")
			case "gen_random_uuid":
				// https://www.postgresql.org/docs/15/functions-uuid.html
				bu.WriteString(convertFuncNameTypeCase("gen_random_uuid", "GEN_RANDOM_UUID", conf))
				bu.WriteString("(")
			case "current_setting":
				// https://www.postgresql.org/docs/15/functions-admin.html#FUNCTIONS-ADMIN-SET
				bu.WriteString(convertFuncNameTypeCase("current_setting", "CURRENT_SETTING", conf))
				bu.WriteString("(")
			case "set_config":
				// https://www.postgresql.org/docs/15/functions-admin.html#FUNCTIONS-ADMIN-SET
				bu.WriteString(convertFuncNameTypeCase("set_config", "SET_CONFIG", conf))
				bu.WriteString("(")
			case "array_agg":
				// https://www.postgresql.org/docs/15/functions-aggregate.html
				bu.WriteString(convertFuncNameTypeCase("array_agg", "ARRAY_AGG", conf))
				bu.WriteString("(")
			}
		}
	}

	return bu.String(), nil
}

func convertFuncNameTypeCase(lower, upper string, conf *fmtconf.Config) string {
	switch conf.FuncCallConfig.FuncNameTypeCase {
	case fmtconf.FUNC_NAME_TYPE_CASE_LOWER:
		return lower
	case fmtconf.FUNC_NAME_TYPE_CASE_UPPER:
		return upper
	}
	return lower
}

func FormatFuncCallArgs(ctx context.Context, funcCall *pg_query.Node_FuncCall) (string, error) {
	var bu strings.Builder

	for argI, arg := range funcCall.FuncCall.Args {
		if argI != 0 {
			bu.WriteString(",")
			bu.WriteString(" ")
		}

		switch n := arg.Node.(type) {
		case *pg_query.Node_AConst:
			res, err := FormatAConst(ctx, n)
			if err != nil {
				return "", err
			}
			bu.WriteString(res)
		case *pg_query.Node_ParamRef:
			bu.WriteString("$")
			bu.WriteString(fmt.Sprint(n.ParamRef.Number))
		case *pg_query.Node_ColumnRef:
			field, err := FormatColumnRefFields(ctx, n)
			if err != nil {
				return "", err
			}
			bu.WriteString(field)
		}
	}

	return bu.String(), nil
}
