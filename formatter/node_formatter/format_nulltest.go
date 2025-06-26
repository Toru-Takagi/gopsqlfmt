package nodeformatter

import (
	"context"
	"fmt"
	"strings"

	pg_query "github.com/pganalyze/pg_query_go/v5"
)

func FormatNullTest(ctx context.Context, nt *pg_query.Node_NullTest) (string, error) {
	var bu strings.Builder

	if nt.NullTest.Arg != nil {
		switch arg := nt.NullTest.Arg.Node.(type) {
		case *pg_query.Node_ColumnRef:
			field, err := FormatColumnRefFields(ctx, arg)
			if err != nil {
				return "", err
			}
			bu.WriteString(field)
		case *pg_query.Node_ParamRef:
			bu.WriteString("$")
			bu.WriteString(fmt.Sprint(arg.ParamRef.Number))
		case *pg_query.Node_AConst:
			res, err := FormatAConst(ctx, arg)
			if err != nil {
				return "", err
			}
			bu.WriteString(res)
		}
	}

	switch nt.NullTest.Nulltesttype {
	case pg_query.NullTestType_IS_NULL:
		bu.WriteString(" IS NULL")
	case pg_query.NullTestType_IS_NOT_NULL:
		bu.WriteString(" IS NOT NULL")
	default:
		return "", fmt.Errorf("FormatNullTest: unknown NullTestType")
	}

	return bu.String(), nil
}
