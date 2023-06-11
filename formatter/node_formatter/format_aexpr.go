package nodeformatter

import (
	"context"
	"fmt"
	"strings"

	pg_query "github.com/pganalyze/pg_query_go/v4"
)

// ex) user_uuid = $1
func FormatAExpr(ctx context.Context, aeXpr *pg_query.Node_AExpr) (string, error) {
	var bu strings.Builder

	// output column name
	if cRef, ok := aeXpr.AExpr.Lexpr.Node.(*pg_query.Node_ColumnRef); ok {
		for _, f := range cRef.ColumnRef.Fields {
			if s, ok := f.Node.(*pg_query.Node_String_); ok {
				bu.WriteString(s.String_.Sval)
			}
		}
	}

	// output operator
	for _, n := range aeXpr.AExpr.Name {
		if s, ok := n.Node.(*pg_query.Node_String_); ok {
			bu.WriteString(" ")
			bu.WriteString(s.String_.Sval)
		}
	}

	// output parameter
	if pRef, ok := aeXpr.AExpr.Rexpr.Node.(*pg_query.Node_ParamRef); ok {
		bu.WriteString(" $")
		bu.WriteString(fmt.Sprint(pRef.ParamRef.Number))
	}

	return bu.String(), nil
}
