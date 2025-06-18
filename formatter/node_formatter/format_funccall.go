package nodeformatter

import (
	"context"
	"fmt"
	"strings"

	"github.com/Toru-Takagi/gopsqlfmt/fmtconf"
	"github.com/Toru-Takagi/gopsqlfmt/formatter/internal"
	pg_query "github.com/pganalyze/pg_query_go/v6"
)

func FormatFuncname(ctx context.Context, funcCall *pg_query.Node_FuncCall, conf *fmtconf.Config) (string, error) {
	var bu strings.Builder

	for _, name := range funcCall.FuncCall.Funcname {
		if s, ok := name.Node.(*pg_query.Node_String_); ok {
			switch s.String_.Sval {
			case "now":
				// https://www.postgresql.org/docs/15/functions-datetime.html
				bu.WriteString(convertFuncNameTypeCase("now", "NOW", conf))
			case "count":
				// https://www.postgresql.org/docs/15/functions-aggregate.html
				bu.WriteString(convertFuncNameTypeCase("count", "COUNT", conf))
			case "min":
				// https://www.postgresql.org/docs/15/functions-aggregate.html
				bu.WriteString(convertFuncNameTypeCase("min", "MIN", conf))
			case "gen_random_uuid":
				// https://www.postgresql.org/docs/15/functions-uuid.html
				bu.WriteString(convertFuncNameTypeCase("gen_random_uuid", "GEN_RANDOM_UUID", conf))
			case "current_setting":
				// https://www.postgresql.org/docs/15/functions-admin.html#FUNCTIONS-ADMIN-SET
				bu.WriteString(convertFuncNameTypeCase("current_setting", "CURRENT_SETTING", conf))
			case "set_config":
				// https://www.postgresql.org/docs/15/functions-admin.html#FUNCTIONS-ADMIN-SET
				bu.WriteString(convertFuncNameTypeCase("set_config", "SET_CONFIG", conf))
			case "array_agg":
				// https://www.postgresql.org/docs/15/functions-aggregate.html
				bu.WriteString(convertFuncNameTypeCase("array_agg", "ARRAY_AGG", conf))
			case "json_agg":
				bu.WriteString(convertFuncNameTypeCase("json_agg", "JSON_AGG", conf))
			case "json_build_object":
				bu.WriteString(convertFuncNameTypeCase("json_build_object", "JSON_BUILD_OBJECT", conf))
			case "array_length":
				bu.WriteString(convertFuncNameTypeCase("array_length", "ARRAY_LENGTH", conf))
			case "cardinality":
				bu.WriteString(convertFuncNameTypeCase("cardinality", "CARDINALITY", conf))
			}
		}
	}

	return bu.String(), nil
}

func convertFuncNameTypeCase(lower, upper string, conf *fmtconf.Config) string {
	switch conf.FuncCallConfig.FuncNameTypeCase {
	case fmtconf.FUNC_NAME_TYPE_CASE_LOWER:
		return lower
	case fmtconf.FUNC_NAME_TYPE_CASE_UPPER:
		return upper
	}
	return lower
}

func FormatFuncCallArgs(ctx context.Context, funcCall *pg_query.Node_FuncCall, indent int, conf *fmtconf.Config) (string, error) {
	var bu strings.Builder

	for argI, arg := range funcCall.FuncCall.Args {
		if argI != 0 {
			bu.WriteString(",")
			bu.WriteString(" ")
		}

		switch n := arg.Node.(type) {
		case *pg_query.Node_AConst:
			res, err := FormatAConst(ctx, n)
			if err != nil {
				return "", err
			}
			bu.WriteString(res)
		case *pg_query.Node_ParamRef:
			bu.WriteString("$")
			bu.WriteString(fmt.Sprint(n.ParamRef.Number))
		case *pg_query.Node_ColumnRef:
			if funcCall.FuncCall.AggDistinct {
				bu.WriteString("DISTINCT ")
			}

			field, err := FormatColumnRefFields(ctx, n)
			if err != nil {
				return "", err
			}
			bu.WriteString(field)
		case *pg_query.Node_FuncCall:
			funcName, err := FormatFuncname(ctx, n, conf)
			if err != nil {
				return "", err
			}
			bu.WriteString(funcName)
			bu.WriteString("(")

			arg, err := FormatFuncCallArgs(ctx, n, indent, conf)
			if err != nil {
				return "", err
			}
			bu.WriteString(arg)

			bu.WriteString(")")
		case *pg_query.Node_SubLink:
			if selectStmt, ok := n.SubLink.Subselect.Node.(*pg_query.Node_SelectStmt); ok {
				res, err := FormatSelectStmtForFuncArg(ctx, selectStmt, indent+1, conf)
				if err != nil {
					return "", err
				}
				bu.WriteString("(\n")
				bu.WriteString(res)
				bu.WriteString("\n")
				for i := 0; i < indent; i++ {
					bu.WriteString(internal.GetIndent(conf))
				}
				bu.WriteString(")")
			}
		}
	}

	if funcCall.FuncCall.AggStar {
		bu.WriteString("*")
	}

	return bu.String(), nil
}

func FormatSelectStmtForFuncArg(ctx context.Context, stmt *pg_query.Node_SelectStmt, indent int, conf *fmtconf.Config) (string, error) {
	if len(stmt.SelectStmt.TargetList) == 0 {
		return "", nil
	}

	var bu strings.Builder
	for i := 0; i < indent; i++ {
		bu.WriteString(internal.GetIndent(conf))
	}
	bu.WriteString("SELECT")

	// output column name
	for ti, node := range stmt.SelectStmt.TargetList {
		if ti != 0 {
			bu.WriteString(",")
		}
		if res, ok := node.Node.(*pg_query.Node_ResTarget); ok {
			switch n := res.ResTarget.Val.Node.(type) {
			case *pg_query.Node_ColumnRef:
				field, err := FormatColumnRefFields(ctx, n)
				if err != nil {
					return "", err
				}
				bu.WriteString("\n")
				for i := 0; i < indent+1; i++ {
					bu.WriteString(internal.GetIndent(conf))
				}
				bu.WriteString(field)
			case *pg_query.Node_FuncCall:
				bu.WriteString("\n")
				for i := 0; i < indent+1; i++ {
					bu.WriteString(internal.GetIndent(conf))
				}
				funcName, err := FormatFuncname(ctx, n, conf)
				if err != nil {
					return "", err
				}
				bu.WriteString(funcName)
				bu.WriteString("(")

				arg, err := FormatFuncCallArgs(ctx, n, indent+1, conf)
				if err != nil {
					return "", err
				}
				bu.WriteString(arg)
				bu.WriteString(")")
			}
			if res.ResTarget.Name != "" {
				bu.WriteString(" AS ")
				bu.WriteString(res.ResTarget.Name)
			}
		}
	}

	// output table name
	for _, node := range stmt.SelectStmt.FromClause {
		res, err := FormatSelectStmtFromClauseForFuncArg(ctx, node.Node, indent, conf)
		if err != nil {
			return "", err
		}
		bu.WriteString(res)
	}

	// output where clause
	if stmt.SelectStmt.WhereClause != nil {
		switch n := stmt.SelectStmt.WhereClause.Node.(type) {
		case *pg_query.Node_AExpr:
			res, err := FormatAExpr(ctx, n, conf)
			if err != nil {
				return "", err
			}
			bu.WriteString("\n")
			for i := 0; i < indent; i++ {
				bu.WriteString(internal.GetIndent(conf))
			}
			bu.WriteString("WHERE")
			bu.WriteString(" ")
			bu.WriteString(res)
		}
	}

	// Handle set operations like UNION ALL
	if stmt.SelectStmt.Op != pg_query.SetOperation_SETOP_NONE {
		if stmt.SelectStmt.Larg != nil {
			leftStmt := &pg_query.Node_SelectStmt{SelectStmt: stmt.SelectStmt.Larg}
			leftRes, err := FormatSelectStmtForFuncArg(ctx, leftStmt, indent, conf)
			if err != nil {
				return "", err
			}
			// Reset bu and start with left side
			bu.Reset()
			bu.WriteString(leftRes)
		}

		// Add set operation
		bu.WriteString("\n")
		for i := 0; i < indent; i++ {
			bu.WriteString(internal.GetIndent(conf))
		}
		switch stmt.SelectStmt.Op {
		case pg_query.SetOperation_SETOP_UNION:
			if stmt.SelectStmt.All {
				bu.WriteString("UNION ALL")
			} else {
				bu.WriteString("UNION")
			}
		case pg_query.SetOperation_SETOP_INTERSECT:
			if stmt.SelectStmt.All {
				bu.WriteString("INTERSECT ALL")
			} else {
				bu.WriteString("INTERSECT")
			}
		case pg_query.SetOperation_SETOP_EXCEPT:
			if stmt.SelectStmt.All {
				bu.WriteString("EXCEPT ALL")
			} else {
				bu.WriteString("EXCEPT")
			}
		}

		// Add right side
		if stmt.SelectStmt.Rarg != nil {
			rightStmt := &pg_query.Node_SelectStmt{SelectStmt: stmt.SelectStmt.Rarg}
			rightRes, err := FormatSelectStmtForFuncArg(ctx, rightStmt, indent, conf)
			if err != nil {
				return "", err
			}
			bu.WriteString("\n")
			bu.WriteString(rightRes)
		}
	}

	return bu.String(), nil
}

func FormatSelectStmtFromClauseForFuncArg(ctx context.Context, node any, indent int, conf *fmtconf.Config) (string, error) {
	var bu strings.Builder

	switch n := node.(type) {
	case *pg_query.Node_RangeVar:
		tName := ""
		if n.RangeVar.Schemaname != "" {
			tName += n.RangeVar.Schemaname
			tName += "."
		}
		tName += n.RangeVar.Relname
		if n.RangeVar.Alias != nil {
			tName += " "
			tName += n.RangeVar.Alias.Aliasname
		}

		bu.WriteString("\n")
		for i := 0; i < indent; i++ {
			bu.WriteString(internal.GetIndent(conf))
		}
		bu.WriteString("FROM")
		bu.WriteString(" ")
		bu.WriteString(tName)
	}

	return bu.String(), nil
}
