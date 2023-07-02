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
			for fi, f := range n.ColumnRef.Fields {
				if s, ok := f.Node.(*pg_query.Node_String_); ok {
					if fi != 0 {
						bu.WriteString(".")
					}
					bu.WriteString(s.String_.Sval)
				}
			}
		}
	}

	return bu.String(), nil
}
