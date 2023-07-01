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
		for fi, f := range cRef.ColumnRef.Fields {
			if s, ok := f.Node.(*pg_query.Node_String_); ok {
				if fi != 0 {
					bu.WriteString(".")
				}
				bu.WriteString(s.String_.Sval)
			}
		}
	}

	// output operator
	for _, n := range aeXpr.AExpr.Name {
		if s, ok := n.Node.(*pg_query.Node_String_); ok {
			bu.WriteString(" ")
			if s.String_.Sval == "<>" {
				bu.WriteString("!=")
			} else {
				bu.WriteString(s.String_.Sval)
			}
		}
	}

	switch rexprNode := aeXpr.AExpr.Rexpr.Node.(type) {
	case *pg_query.Node_AConst:
		res, err := FormatAConst(ctx, rexprNode)
		if err != nil {
			return "", err
		}
		bu.WriteString(" ")
		bu.WriteString(res)

	// output parameter (if named parameter)
	case *pg_query.Node_ColumnRef:
		for fi, f := range rexprNode.ColumnRef.Fields {
			if s, ok := f.Node.(*pg_query.Node_String_); ok {
				if fi == 0 {
					bu.WriteString(" ")
				} else {
					bu.WriteString(".")
				}

				bu.WriteString(s.String_.Sval)
			}
		}

	// output parameter (if $1)
	case *pg_query.Node_ParamRef:
		bu.WriteString(" $")
		bu.WriteString(fmt.Sprint(rexprNode.ParamRef.Number))
	}

	return bu.String(), nil
}
