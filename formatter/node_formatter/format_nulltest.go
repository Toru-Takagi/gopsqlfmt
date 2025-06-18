package nodeformatter

import (
	"context"
	"strings"

	"github.com/Toru-Takagi/gopsqlfmt/formatter/enumconv"
	pg_query "github.com/pganalyze/pg_query_go/v6"
)

func FormatNullTest(ctx context.Context, nullTest *pg_query.Node_NullTest) (string, error) {
	var bu strings.Builder

	// Format the argument (column or expression)
	if nullTest.NullTest.Arg != nil {
		switch argNode := nullTest.NullTest.Arg.Node.(type) {
		case *pg_query.Node_ColumnRef:
			field, err := FormatColumnRefFields(ctx, argNode)
			if err != nil {
				return "", err
			}
			bu.WriteString(field)
		}
	}

	// Format the null test type (IS NULL or IS NOT NULL)
	nullTestStr, err := enumconv.NullTestTypeToString(nullTest.NullTest.Nulltesttype)
	if err != nil {
		return "", err
	}
	bu.WriteString(" ")
	bu.WriteString(nullTestStr)

	return bu.String(), nil
}