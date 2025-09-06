package nodeformatter

import (
	"context"
	"fmt"
	"strings"

	"github.com/Toru-Takagi/gopsqlfmt/fmtconf"
	pg_query "github.com/pganalyze/pg_query_go/v6"
)

// ex) user_uuid = $1
func FormatAExpr(ctx context.Context, aeXpr *pg_query.Node_AExpr, conf *fmtconf.Config) (string, error) {
	var bu strings.Builder

	// output column name or function call
	switch lexprNode := aeXpr.AExpr.Lexpr.Node.(type) {
	case *pg_query.Node_ColumnRef:
		field, err := FormatColumnRefFields(ctx, lexprNode)
		if err != nil {
			return "", err
		}
		bu.WriteString(field)
	case *pg_query.Node_FuncCall:
		funcCall, err := FormatFuncname(ctx, lexprNode, conf)
		if err != nil {
			return "", err
		}
		bu.WriteString(funcCall)
		bu.WriteString("(")

		arg, err := FormatFuncCallArgs(ctx, lexprNode, 0, conf)
		if err != nil {
			return "", err
		}
		bu.WriteString(arg)
		bu.WriteString(")")
	case *pg_query.Node_CoalesceExpr:
		bu.WriteString("COALESCE")
		bu.WriteString("(")

		for argI, arg := range lexprNode.CoalesceExpr.Args {
			if argI != 0 {
				bu.WriteString(",")
				bu.WriteString(" ")
			}
			switch n := arg.Node.(type) {
			case *pg_query.Node_ColumnRef:
				field, err := FormatColumnRefFields(ctx, n)
				if err != nil {
					return "", err
				}
				bu.WriteString(field)
			case *pg_query.Node_AConst:
				aconst, err := FormatAConst(ctx, n)
				if err != nil {
					return "", err
				}
				bu.WriteString(aconst)
			case *pg_query.Node_ParamRef:
				bu.WriteString("$")
				bu.WriteString(fmt.Sprint(n.ParamRef.Number))
			}
		}

		bu.WriteString(")")
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
		field, err := FormatColumnRefFields(ctx, rexprNode)
		if err != nil {
			return "", err
		}
		bu.WriteString(" ")
		bu.WriteString(field)

	// output parameter (if $1)
	case *pg_query.Node_ParamRef:
		// Handle ANY operator specially
		if aeXpr.AExpr.Kind == pg_query.A_Expr_Kind_AEXPR_OP_ANY {
			bu.WriteString(" ANY($")
			bu.WriteString(fmt.Sprint(rexprNode.ParamRef.Number))
			bu.WriteString(")")
		} else {
			bu.WriteString(" $")
			bu.WriteString(fmt.Sprint(rexprNode.ParamRef.Number))
		}

	case *pg_query.Node_FuncCall:
		funcCall, err := FormatFuncname(ctx, rexprNode, conf)
		if err != nil {
			return "", err
		}
		bu.WriteString(" ")
		bu.WriteString(funcCall)
		bu.WriteString("(")

		arg, err := FormatFuncCallArgs(ctx, rexprNode, 0, conf)
		if err != nil {
			return "", err
		}
		bu.WriteString(arg)
		bu.WriteString(")")

	case *pg_query.Node_TypeCast:
		if rexprNode.TypeCast.TypeName != nil {
			if len(rexprNode.TypeCast.TypeName.ArrayBounds) > 0 {
				bu.WriteString(" ")
				bu.WriteString("ANY")
				bu.WriteString("(")
				if rexprNode.TypeCast.Arg != nil {
					switch arg := rexprNode.TypeCast.Arg.Node.(type) {
					case *pg_query.Node_AConst:
						res, err := FormatAConst(ctx, arg)
						if err != nil {
							return "", err
						}
						bu.WriteString(res)
						bu.WriteString("::")
						switch n := rexprNode.TypeCast.TypeName.Names[0].Node.(type) {
						case *pg_query.Node_String_:
							bu.WriteString(n.String_.Sval)
						}
						bu.WriteString("[]")
					case *pg_query.Node_ColumnRef:
						field, err := FormatColumnRefFields(ctx, arg)
						if err != nil {
							return "", err
						}
						bu.WriteString(field)
						bu.WriteString("::")
						switch n := rexprNode.TypeCast.TypeName.Names[0].Node.(type) {
						case *pg_query.Node_String_:
							bu.WriteString(n.String_.Sval)
						}
						bu.WriteString("[]")
					}
				}
				bu.WriteString(")")
			}
		}

	case *pg_query.Node_CoalesceExpr:
		bu.WriteString(" COALESCE")
		bu.WriteString("(")

		for argI, arg := range rexprNode.CoalesceExpr.Args {
			if argI != 0 {
				bu.WriteString(",")
				bu.WriteString(" ")
			}
			switch n := arg.Node.(type) {
			case *pg_query.Node_ColumnRef:
				field, err := FormatColumnRefFields(ctx, n)
				if err != nil {
					return "", err
				}
				bu.WriteString(field)
			case *pg_query.Node_AConst:
				aconst, err := FormatAConst(ctx, n)
				if err != nil {
					return "", err
				}
				bu.WriteString(aconst)
			case *pg_query.Node_ParamRef:
				bu.WriteString("$")
				bu.WriteString(fmt.Sprint(n.ParamRef.Number))
			}
		}

		bu.WriteString(")")
	}

	return bu.String(), nil
}
