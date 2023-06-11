package formatter

import (
	nodeformatter "Toru-Takagi/sql_formatter_go/formatter/node_formatter"
	"context"
	"errors"
	"strings"

	pg_query "github.com/pganalyze/pg_query_go/v4"
)

func formatBoolExpr(ctx context.Context, be *pg_query.Node_BoolExpr, indent int) (string, error) {
	var bu strings.Builder

	for argI, arg := range be.BoolExpr.Args {
		switch n := arg.Node.(type) {
		case *pg_query.Node_AExpr:
			res, err := nodeformatter.FormatAExpr(ctx, n)
			if err != nil {
				return "", err
			}
			if argI != 0 {
				bu.WriteString("\n")
				for i := 0; i <= indent; i++ {
					bu.WriteString("\t")
				}
				boolStr, err := BoolExprTypeToString(be.BoolExpr.Boolop)
				if err != nil {
					return "", err
				}
				bu.WriteString(boolStr)
				bu.WriteString(" ")
			}
			bu.WriteString(res)
		case *pg_query.Node_BoolExpr:
			res, err := formatBoolExpr(ctx, n, indent+2)
			if err != nil {
				return "", err
			}
			if argI != 0 {
				bu.WriteString("\n")
				bu.WriteString("\t")
				boolStr, err := BoolExprTypeToString(be.BoolExpr.Boolop)
				if err != nil {
					return "", err
				}
				bu.WriteString(boolStr)
			}
			bu.WriteString(" ")
			bu.WriteString("(")
			bu.WriteString("\n")
			bu.WriteString("\t")
			bu.WriteString("\t")
			bu.WriteString(res)
			bu.WriteString("\n")
			bu.WriteString("\t")
			bu.WriteString(")")
		default:
			return "", errors.New("formatBoolExpr: unknown node type")
		}
	}

	return bu.String(), nil
}

func BoolExprTypeToString(bet pg_query.BoolExprType) (string, error) {
	switch bet {
	case pg_query.BoolExprType_AND_EXPR:
		return "AND", nil
	case pg_query.BoolExprType_OR_EXPR:
		return "OR", nil
	case pg_query.BoolExprType_NOT_EXPR:
		return "NOT", nil
	}
	return "", errors.New("BoolExprTypeToString: unknown BoolExprType")
}
