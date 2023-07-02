package nodeformatter

import (
	"context"
	"strings"

	pg_query "github.com/pganalyze/pg_query_go/v4"
)

func FormatColumnRefFields(ctx context.Context, columnRef *pg_query.Node_ColumnRef) (string, error) {
	var bu strings.Builder

	for fi, f := range columnRef.ColumnRef.Fields {
		if s, ok := f.Node.(*pg_query.Node_String_); ok {
			if fi != 0 {
				bu.WriteString(".")
			}
			if s.String_.Sval == "excluded" {
				bu.WriteString("EXCLUDED")
			} else {
				bu.WriteString(s.String_.Sval)
			}
		}
	}

	return bu.String(), nil
}
