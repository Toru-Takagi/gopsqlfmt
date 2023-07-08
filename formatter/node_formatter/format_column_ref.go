package nodeformatter

import (
	"context"
	"strings"

	pg_query "github.com/pganalyze/pg_query_go/v4"
)

func FormatColumnRefFields(ctx context.Context, columnRef *pg_query.Node_ColumnRef) (string, error) {
	var bu strings.Builder

	for fi, f := range columnRef.ColumnRef.Fields {
		switch n := f.Node.(type) {
		case *pg_query.Node_String_:
			if fi != 0 {
				bu.WriteString(".")
			}
			if n.String_.Sval == "excluded" {
				bu.WriteString("EXCLUDED")
			} else {
				bu.WriteString(n.String_.Sval)
			}
		case *pg_query.Node_AStar:
			bu.WriteString("*")
		}
	}

	return bu.String(), nil
}
