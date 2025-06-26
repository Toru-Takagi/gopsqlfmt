package formatter

import (
	"github.com/Toru-Takagi/gopsqlfmt/fmtconf"
	"github.com/Toru-Takagi/gopsqlfmt/formatter/enumconv"
	"github.com/Toru-Takagi/gopsqlfmt/formatter/internal"
	nodeformatter "github.com/Toru-Takagi/gopsqlfmt/formatter/node_formatter"

	"context"
	"strings"

	pg_query "github.com/pganalyze/pg_query_go/v5"
)

func formatBoolExpr(ctx context.Context, be *pg_query.Node_BoolExpr, indent int, conf *fmtconf.Config) (string, error) {
	var bu strings.Builder

	for argI, arg := range be.BoolExpr.Args {
		switch n := arg.Node.(type) {
		case *pg_query.Node_AExpr:
			res, err := nodeformatter.FormatAExpr(ctx, n, conf)
			if err != nil {
				return "", err
			}
			if argI != 0 {
				bu.WriteString("\n")
				for i := 0; i <= indent; i++ {
					bu.WriteString(internal.GetIndent(conf))
				}
				boolStr, err := enumconv.BoolExprTypeToString(be.BoolExpr.Boolop)
				if err != nil {
					return "", err
				}
				bu.WriteString(boolStr)
				bu.WriteString(" ")
			}
			bu.WriteString(res)
		case *pg_query.Node_BoolExpr:
			res, err := formatBoolExpr(ctx, n, indent+2, conf)
			if err != nil {
				return "", err
			}
			if argI != 0 {
				bu.WriteString("\n")
				bu.WriteString(internal.GetIndent(conf))
				boolStr, err := enumconv.BoolExprTypeToString(be.BoolExpr.Boolop)
				if err != nil {
					return "", err
				}
				bu.WriteString(boolStr)
			}
			bu.WriteString(" ")
			bu.WriteString("(")
			bu.WriteString("\n")
			bu.WriteString(internal.GetIndent(conf))
			bu.WriteString(internal.GetIndent(conf))
			bu.WriteString(res)
			bu.WriteString("\n")
			bu.WriteString(internal.GetIndent(conf))
			bu.WriteString(")")
		case *pg_query.Node_NullTest:
			res, err := nodeformatter.FormatNullTest(ctx, n)
			if err != nil {
				return "", err
			}
			if argI != 0 {
				bu.WriteString("\n")
				for i := 0; i <= indent; i++ {
					bu.WriteString(internal.GetIndent(conf))
				}
				boolStr, err := enumconv.BoolExprTypeToString(be.BoolExpr.Boolop)
				if err != nil {
					return "", err
				}
				bu.WriteString(boolStr)
				bu.WriteString(" ")
			}
			bu.WriteString(res)
		case *pg_query.Node_SubLink:
			if selectStmt, ok := n.SubLink.Subselect.Node.(*pg_query.Node_SelectStmt); ok {
				res, err := FormatSelectStmt(ctx, selectStmt, indent+1, conf)
				if err != nil {
					return "", err
				}

				boolStr, err := enumconv.BoolExprTypeToString(be.BoolExpr.Boolop)
				if err != nil {
					return "", err
				}
				bu.WriteString(boolStr)
				bu.WriteString(" ")

				sbt, err := enumconv.SubLinkTypeToString(n.SubLink.SubLinkType)
				if err != nil {
					return "", err
				}
				bu.WriteString(sbt)

				bu.WriteString("(\n")
				bu.WriteString(res)
				bu.WriteString("\n")
				bu.WriteString(")")
			}
		}
	}

	return bu.String(), nil
}
