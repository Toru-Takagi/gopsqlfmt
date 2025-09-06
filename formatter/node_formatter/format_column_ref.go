package nodeformatter

import (
	"context"
	"strings"

	pg_query "github.com/pganalyze/pg_query_go/v6"
)

func FormatColumnRefFields(ctx context.Context, columnRef *pg_query.Node_ColumnRef) (string, error) {
	var bu strings.Builder

	for fi, f := range columnRef.ColumnRef.Fields {
		switch n := f.Node.(type) {
		case *pg_query.Node_String_:
			if fi != 0 {
				bu.WriteString(".")
			}
			switch strings.ToLower(n.String_.Sval) {
			case "excluded":
				bu.WriteString("EXCLUDED")
			case "current_timestamp":
				bu.WriteString("CURRENT_TIMESTAMP")
			case "current_date":
				bu.WriteString("CURRENT_DATE")
			case "current_time":
				bu.WriteString("CURRENT_TIME")
			case "localtime":
				bu.WriteString("LOCALTIME")
			case "localtimestamp":
				bu.WriteString("LOCALTIMESTAMP")
			default:
				bu.WriteString(n.String_.Sval)
			}
		case *pg_query.Node_AStar:
			bu.WriteString("*")
		}
	}

	return bu.String(), nil
}
