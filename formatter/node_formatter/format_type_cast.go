package nodeformatter

import (
	"context"
	"strings"

	pg_query "github.com/pganalyze/pg_query_go/v5"
)

func FormatTypeCast(ctx context.Context, tc *pg_query.Node_TypeCast) (string, error) {
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
			res, err := FormatTypeCast(ctx, arg)
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
				bu.WriteString(name.String_.Sval)
			}
		}
		if len(tc.TypeCast.TypeName.ArrayBounds) > 0 {
			bu.WriteString("[]")
		}
	}

	return bu.String(), nil
}
