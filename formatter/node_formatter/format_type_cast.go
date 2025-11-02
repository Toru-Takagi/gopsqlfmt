package nodeformatter

import (
	"context"
	"fmt"
	"strings"

	"github.com/Toru-Takagi/gopsqlfmt/fmtconf"
	pg_query "github.com/pganalyze/pg_query_go/v6"
)

func FormatTypeCast(ctx context.Context, tc *pg_query.Node_TypeCast, conf *fmtconf.Config) (string, error) {
	var bu strings.Builder

	if tc.TypeCast.Arg != nil {
		switch arg := tc.TypeCast.Arg.Node.(type) {
		case *pg_query.Node_ColumnRef:
			field, err := FormatColumnRefFields(ctx, arg)
			if err != nil {
				return "", err
			}
			bu.WriteString(field)
		case *pg_query.Node_AConst:
			res, err := FormatAConst(ctx, arg)
			if err != nil {
				return "", err
			}
			bu.WriteString(res)
		case *pg_query.Node_TypeCast:
			res, err := FormatTypeCast(ctx, arg, conf)
			if err != nil {
				return "", err
			}
			bu.WriteString(res)
		case *pg_query.Node_FuncCall:
			funcName, err := FormatFuncname(ctx, arg, conf)
			if err != nil {
				return "", err
			}
			bu.WriteString(funcName)
			bu.WriteString("(")
			args, err := FormatFuncCallArgs(ctx, arg, 0, conf)
			if err != nil {
				return "", err
			}
			bu.WriteString(args)
			bu.WriteString(")")
		case *pg_query.Node_ParamRef:
			bu.WriteString("$")
			bu.WriteString(fmt.Sprint(arg.ParamRef.Number))
		case *pg_query.Node_AExpr:
			res, err := FormatAExpr(ctx, arg, conf)
			if err != nil {
				return "", err
			}
			bu.WriteString(res)
		}
	}

	if tc.TypeCast.TypeName != nil {
		bu.WriteString("::")
		if len(tc.TypeCast.TypeName.Names) > 0 {
			if name, ok := tc.TypeCast.TypeName.Names[len(tc.TypeCast.TypeName.Names)-1].Node.(*pg_query.Node_String_); ok {
				bu.WriteString(normalizeTypeName(name.String_.Sval))
			}
		}
		if len(tc.TypeCast.TypeName.ArrayBounds) > 0 {
			bu.WriteString("[]")
		}
	}

	return bu.String(), nil
}

func normalizeTypeName(name string) string {
	switch strings.ToLower(name) {
	case "int2":
		return "smallint"
	case "int4":
		return "integer"
	case "int8":
		return "bigint"
	}
	return name
}
