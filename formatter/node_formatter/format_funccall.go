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
			case "count":
				// https://www.postgresql.org/docs/15/functions-aggregate.html
				bu.WriteString(convertFuncNameTypeCase("count", "COUNT", conf))
			case "min":
				// https://www.postgresql.org/docs/15/functions-aggregate.html
				bu.WriteString(convertFuncNameTypeCase("min", "MIN", conf))
			case "gen_random_uuid":
				// https://www.postgresql.org/docs/15/functions-uuid.html
				bu.WriteString(convertFuncNameTypeCase("gen_random_uuid", "GEN_RANDOM_UUID", conf))
			case "current_setting":
				// https://www.postgresql.org/docs/15/functions-admin.html#FUNCTIONS-ADMIN-SET
				bu.WriteString(convertFuncNameTypeCase("current_setting", "CURRENT_SETTING", conf))
			case "set_config":
				// https://www.postgresql.org/docs/15/functions-admin.html#FUNCTIONS-ADMIN-SET
				bu.WriteString(convertFuncNameTypeCase("set_config", "SET_CONFIG", conf))
			case "array_agg":
				// https://www.postgresql.org/docs/15/functions-aggregate.html
				bu.WriteString(convertFuncNameTypeCase("array_agg", "ARRAY_AGG", conf))
			case "json_agg":
				bu.WriteString(convertFuncNameTypeCase("json_agg", "JSON_AGG", conf))
			case "json_build_object":
				bu.WriteString(convertFuncNameTypeCase("json_build_object", "JSON_BUILD_OBJECT", conf))
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

func FormatFuncCallArgs(ctx context.Context, funcCall *pg_query.Node_FuncCall, indent int, conf *fmtconf.Config) (string, error) {
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
		case *pg_query.Node_FuncCall:
			funcName, err := FormatFuncname(ctx, n, conf)
			if err != nil {
				return "", err
			}
			bu.WriteString(funcName)
			bu.WriteString("(")

			arg, err := FormatFuncCallArgs(ctx, n, indent, conf)
			if err != nil {
				return "", err
			}
			bu.WriteString(arg)

			bu.WriteString(")")
		}
	}

	if funcCall.FuncCall.AggStar {
		bu.WriteString("*")
	}

	return bu.String(), nil
}
