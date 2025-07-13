package nodeformatter

import (
	"context"
	"strings"

	"github.com/Toru-Takagi/gopsqlfmt/fmtconf"
	"github.com/Toru-Takagi/gopsqlfmt/formatter/internal"

	pg_query "github.com/pganalyze/pg_query_go/v6"
)

func FormatCaseExpr(ctx context.Context, n *pg_query.Node_CaseExpr, indent int, conf *fmtconf.Config) (string, error) {
	var bu strings.Builder

	bu.WriteString("CASE")

	// Handle CASE ... WHEN test
	if n.CaseExpr.Arg != nil {
		bu.WriteString(" ")
		switch arg := n.CaseExpr.Arg.Node.(type) {
		case *pg_query.Node_ColumnRef:
			field, err := FormatColumnRefFields(ctx, arg)
			if err != nil {
				return "", err
			}
			bu.WriteString(field)
		}
	}

	// Handle WHEN clauses
	for _, when := range n.CaseExpr.Args {
		if whenClause, ok := when.Node.(*pg_query.Node_CaseWhen); ok {
			bu.WriteString(" WHEN ")

			// Handle WHEN condition
			if whenClause.CaseWhen.Expr != nil {
				switch expr := whenClause.CaseWhen.Expr.Node.(type) {
				case *pg_query.Node_AExpr:
					res, err := FormatAExpr(ctx, expr, conf)
					if err != nil {
						return "", err
					}
					bu.WriteString(res)
				case *pg_query.Node_NullTest:
					res, err := FormatNullTest(ctx, expr)
					if err != nil {
						return "", err
					}
					bu.WriteString(res)
				}
			}

			bu.WriteString(" THEN ")

			// Handle THEN result
			if whenClause.CaseWhen.Result != nil {
				switch result := whenClause.CaseWhen.Result.Node.(type) {
				case *pg_query.Node_AConst:
					aconst, err := FormatAConst(ctx, result)
					if err != nil {
						return "", err
					}
					bu.WriteString(aconst)
				case *pg_query.Node_ColumnRef:
					field, err := FormatColumnRefFields(ctx, result)
					if err != nil {
						return "", err
					}
					bu.WriteString(field)
				}
			}
		}
	}

	// Handle ELSE clause
	if n.CaseExpr.Defresult != nil {
		bu.WriteString(" ELSE ")
		switch defResult := n.CaseExpr.Defresult.Node.(type) {
		case *pg_query.Node_AConst:
			aconst, err := FormatAConst(ctx, defResult)
			if err != nil {
				return "", err
			}
			bu.WriteString(aconst)
		case *pg_query.Node_ColumnRef:
			field, err := FormatColumnRefFields(ctx, defResult)
			if err != nil {
				return "", err
			}
			bu.WriteString(field)
		case *pg_query.Node_SubLink:
			// Handle subqueries in ELSE clause
			if selectStmt, ok := defResult.SubLink.Subselect.Node.(*pg_query.Node_SelectStmt); ok {
				subRes, err := FormatSelectStmtForFuncArg(ctx, selectStmt, indent+1, conf)
				if err != nil {
					return "", err
				}
				bu.WriteString("(\n")
				for i := 0; i < indent+1; i++ {
					bu.WriteString(internal.GetIndent(conf))
				}
				bu.WriteString(subRes)
				bu.WriteString("\n")
				for i := 0; i < indent; i++ {
					bu.WriteString(internal.GetIndent(conf))
				}
				bu.WriteString(")")
			}
		}
	}

	bu.WriteString(" END")

	return bu.String(), nil
}
