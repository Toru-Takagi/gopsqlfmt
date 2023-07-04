package nodeformatter

import (
	"context"
	"fmt"
	"strings"

	pg_query "github.com/pganalyze/pg_query_go/v4"
)

func FormatFuncname(ctx context.Context, funcCall *pg_query.Node_FuncCall) (string, error) {
	var bu strings.Builder

	for _, name := range funcCall.FuncCall.Funcname {
		if s, ok := name.Node.(*pg_query.Node_String_); ok {
			switch s.String_.Sval {
			case "now":
				// https://www.postgresql.org/docs/15/functions-datetime.html
				bu.WriteString("now")
				bu.WriteString("(")
			case "count":
				// https://www.postgresql.org/docs/15/functions-aggregate.html
				bu.WriteString("count")
				bu.WriteString("(")
				bu.WriteString("*")
			case "min":
				// https://www.postgresql.org/docs/15/functions-aggregate.html
				bu.WriteString("min")
				bu.WriteString("(")
			case "gen_random_uuid":
				// https://www.postgresql.org/docs/15/functions-uuid.html
				bu.WriteString("gen_random_uuid")
				bu.WriteString("(")
			case "current_setting":
				// https://www.postgresql.org/docs/15/functions-admin.html#FUNCTIONS-ADMIN-SET
				bu.WriteString("current_setting")
				bu.WriteString("(")
			case "set_config":
				// https://www.postgresql.org/docs/15/functions-admin.html#FUNCTIONS-ADMIN-SET
				bu.WriteString("set_config")
				bu.WriteString("(")
			case "array_agg":
				// https://www.postgresql.org/docs/15/functions-aggregate.html
				bu.WriteString("array_agg")
				bu.WriteString("(")
			}
		}
	}

	return bu.String(), nil
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
