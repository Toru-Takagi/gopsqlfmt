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
	parts := make([]string, 0, len(funcCall.FuncCall.Funcname))

	for _, name := range funcCall.FuncCall.Funcname {
		if s, ok := name.Node.(*pg_query.Node_String_); ok {
			switch strings.ToLower(s.String_.Sval) {
			case "now":
				// https://www.postgresql.org/docs/15/functions-datetime.html
				parts = append(parts, convertFuncNameTypeCase("now", "NOW", conf))
			case "count":
				// https://www.postgresql.org/docs/15/functions-aggregate.html
				parts = append(parts, convertFuncNameTypeCase("count", "COUNT", conf))
			case "min":
				// https://www.postgresql.org/docs/15/functions-aggregate.html
				parts = append(parts, convertFuncNameTypeCase("min", "MIN", conf))
			case "gen_random_uuid":
				// https://www.postgresql.org/docs/15/functions-uuid.html
				parts = append(parts, convertFuncNameTypeCase("gen_random_uuid", "GEN_RANDOM_UUID", conf))
			case "current_setting":
				// https://www.postgresql.org/docs/15/functions-admin.html#FUNCTIONS-ADMIN-SET
				parts = append(parts, convertFuncNameTypeCase("current_setting", "CURRENT_SETTING", conf))
			case "set_config":
				// https://www.postgresql.org/docs/15/functions-admin.html#FUNCTIONS-ADMIN-SET
				parts = append(parts, convertFuncNameTypeCase("set_config", "SET_CONFIG", conf))
			case "array_agg":
				// https://www.postgresql.org/docs/15/functions-aggregate.html
				parts = append(parts, convertFuncNameTypeCase("array_agg", "ARRAY_AGG", conf))
			case "json_agg":
				parts = append(parts, convertFuncNameTypeCase("json_agg", "JSON_AGG", conf))
			case "json_build_object":
				parts = append(parts, convertFuncNameTypeCase("json_build_object", "JSON_BUILD_OBJECT", conf))
			case "array_length":
				parts = append(parts, convertFuncNameTypeCase("array_length", "ARRAY_LENGTH", conf))
			case "cardinality":
				parts = append(parts, convertFuncNameTypeCase("cardinality", "CARDINALITY", conf))
			case "date":
				// https://www.postgresql.org/docs/15/functions-datetime.html
				parts = append(parts, convertFuncNameTypeCase("date", "DATE", conf))
			case "current_timestamp":
				// https://www.postgresql.org/docs/15/functions-datetime.html
				parts = append(parts, convertFuncNameTypeCase("current_timestamp", "CURRENT_TIMESTAMP", conf))
			case "current_date":
				// https://www.postgresql.org/docs/15/functions-datetime.html
				parts = append(parts, convertFuncNameTypeCase("current_date", "CURRENT_DATE", conf))
			case "current_time":
				// https://www.postgresql.org/docs/15/functions-datetime.html
				parts = append(parts, convertFuncNameTypeCase("current_time", "CURRENT_TIME", conf))
			case "localtime":
				// https://www.postgresql.org/docs/15/functions-datetime.html
				parts = append(parts, convertFuncNameTypeCase("localtime", "LOCALTIME", conf))
			case "localtimestamp":
				// https://www.postgresql.org/docs/15/functions-datetime.html
				parts = append(parts, convertFuncNameTypeCase("localtimestamp", "LOCALTIMESTAMP", conf))
			default:
				parts = append(parts, applyFuncNameCase(s.String_.Sval, conf))
			}
		}
	}

	return strings.Join(parts, "."), nil
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

func applyFuncNameCase(name string, conf *fmtconf.Config) string {
	switch conf.FuncCallConfig.FuncNameTypeCase {
	case fmtconf.FUNC_NAME_TYPE_CASE_UPPER:
		return strings.ToUpper(name)
	}
	return strings.ToLower(name)
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
		case *pg_query.Node_AExpr:
			expr, err := FormatAExpr(ctx, n, conf)
			if err != nil {
				return "", err
			}
			bu.WriteString(expr)
		case *pg_query.Node_CoalesceExpr:
			bu.WriteString("COALESCE")
			bu.WriteString("(")
			for coalesceArgI, coalesceArg := range n.CoalesceExpr.Args {
				if coalesceArgI != 0 {
					bu.WriteString(",")
					bu.WriteString(" ")
				}
				switch coalesceNode := coalesceArg.Node.(type) {
				case *pg_query.Node_ColumnRef:
					field, err := FormatColumnRefFields(ctx, coalesceNode)
					if err != nil {
						return "", err
					}
					bu.WriteString(field)
				case *pg_query.Node_AConst:
					aconst, err := FormatAConst(ctx, coalesceNode)
					if err != nil {
						return "", err
					}
					bu.WriteString(aconst)
				case *pg_query.Node_ParamRef:
					bu.WriteString("$")
					bu.WriteString(fmt.Sprint(coalesceNode.ParamRef.Number))
				case *pg_query.Node_FuncCall:
					funcName, err := FormatFuncname(ctx, coalesceNode, conf)
					if err != nil {
						return "", err
					}
					bu.WriteString(funcName)
					bu.WriteString("(")
					funcArg, err := FormatFuncCallArgs(ctx, coalesceNode, indent, conf)
					if err != nil {
						return "", err
					}
					bu.WriteString(funcArg)
					bu.WriteString(")")
				case *pg_query.Node_AExpr:
					coalesceExpr, err := FormatAExpr(ctx, coalesceNode, conf)
					if err != nil {
						return "", err
					}
					bu.WriteString(coalesceExpr)
				}
			}
			bu.WriteString(")")
		case *pg_query.Node_TypeCast:
			res, err := FormatTypeCast(ctx, n, conf)
			if err != nil {
				return "", err
			}
			bu.WriteString(res)
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

	// Handle set operations like UNION ALL first (if this is part of a set operation)
	if stmt.SelectStmt.Op != pg_query.SetOperation_SETOP_NONE {
		if stmt.SelectStmt.Larg != nil {
			leftStmt := &pg_query.Node_SelectStmt{SelectStmt: stmt.SelectStmt.Larg}
			leftRes, err := FormatSelectStmtForFuncArg(ctx, leftStmt, indent, conf)
			if err != nil {
				return "", err
			}
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

		return bu.String(), nil
	}

	// Normal SELECT statement processing
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
			bu.WriteString("\n")
			for i := 0; i < indent+1; i++ {
				bu.WriteString(internal.GetIndent(conf))
			}

			// Handle all target types comprehensively
			switch n := res.ResTarget.Val.Node.(type) {
			case *pg_query.Node_ColumnRef:
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

				arg, err := FormatFuncCallArgs(ctx, n, indent+1, conf)
				if err != nil {
					return "", err
				}
				bu.WriteString(arg)
				bu.WriteString(")")
			case *pg_query.Node_AConst:
				constValue, err := FormatAConst(ctx, n)
				if err != nil {
					return "", err
				}
				bu.WriteString(constValue)
			case *pg_query.Node_SubLink:
				// Handle subqueries in target list
				if selectStmt, ok := n.SubLink.Subselect.Node.(*pg_query.Node_SelectStmt); ok {
					subRes, err := FormatSelectStmtForFuncArg(ctx, selectStmt, indent+1, conf)
					if err != nil {
						return "", err
					}
					bu.WriteString("(\n")
					bu.WriteString(subRes)
					bu.WriteString("\n")
					for i := 0; i < indent+1; i++ {
						bu.WriteString(internal.GetIndent(conf))
					}
					bu.WriteString(")")
				}
			default:
				// Handle other expression types that might be in target list
				bu.WriteString("*") // fallback
			}

			if res.ResTarget.Name != "" {
				bu.WriteString(" AS ")
				bu.WriteString(res.ResTarget.Name)
			}
		}
	}

	// output FROM clause
	for _, node := range stmt.SelectStmt.FromClause {
		res, err := FormatSelectStmtFromClauseForFuncArg(ctx, node.Node, indent, conf)
		if err != nil {
			return "", err
		}
		bu.WriteString(res)
	}

	// output WHERE clause - enhanced to handle all WHERE clause types
	if stmt.SelectStmt.WhereClause != nil {
		bu.WriteString("\n")
		for i := 0; i < indent; i++ {
			bu.WriteString(internal.GetIndent(conf))
		}
		bu.WriteString("WHERE ")

		whereRes, err := formatWhereClauseNode(ctx, stmt.SelectStmt.WhereClause, indent, conf)
		if err != nil {
			return "", err
		}
		bu.WriteString(whereRes)
	}

	// output GROUP BY clause
	if len(stmt.SelectStmt.GroupClause) > 0 {
		bu.WriteString("\n")
		for i := 0; i < indent; i++ {
			bu.WriteString(internal.GetIndent(conf))
		}
		bu.WriteString("GROUP BY ")

		for gi, groupItem := range stmt.SelectStmt.GroupClause {
			if gi != 0 {
				bu.WriteString(", ")
			}
			if sortBy, ok := groupItem.Node.(*pg_query.Node_SortBy); ok {
				if colRef, ok := sortBy.SortBy.Node.Node.(*pg_query.Node_ColumnRef); ok {
					field, err := FormatColumnRefFields(ctx, colRef)
					if err != nil {
						return "", err
					}
					bu.WriteString(field)
				}
			}
		}
	}

	// output ORDER BY clause
	if len(stmt.SelectStmt.SortClause) > 0 {
		bu.WriteString("\n")
		for i := 0; i < indent; i++ {
			bu.WriteString(internal.GetIndent(conf))
		}
		bu.WriteString("ORDER BY ")

		for si, sortItem := range stmt.SelectStmt.SortClause {
			if si != 0 {
				bu.WriteString(", ")
			}
			if sortBy, ok := sortItem.Node.(*pg_query.Node_SortBy); ok {
				// Format the sort expression
				if colRef, ok := sortBy.SortBy.Node.Node.(*pg_query.Node_ColumnRef); ok {
					field, err := FormatColumnRefFields(ctx, colRef)
					if err != nil {
						return "", err
					}
					bu.WriteString(field)
				} else if funcCall, ok := sortBy.SortBy.Node.Node.(*pg_query.Node_FuncCall); ok {
					funcName, err := FormatFuncname(ctx, funcCall, conf)
					if err != nil {
						return "", err
					}
					bu.WriteString(funcName)
					bu.WriteString("(")
					arg, err := FormatFuncCallArgs(ctx, funcCall, indent+1, conf)
					if err != nil {
						return "", err
					}
					bu.WriteString(arg)
					bu.WriteString(")")
				}

				// Add sort direction
				sortDir, err := FormatSortByDir(ctx, sortBy)
				if err != nil {
					return "", err
				}
				bu.WriteString(sortDir)
			}
		}
	}

	// output LIMIT clause
	if stmt.SelectStmt.LimitCount != nil {
		bu.WriteString("\n")
		for i := 0; i < indent; i++ {
			bu.WriteString(internal.GetIndent(conf))
		}
		bu.WriteString("LIMIT ")
		if aConst, ok := stmt.SelectStmt.LimitCount.Node.(*pg_query.Node_AConst); ok {
			limitValue, err := FormatAConst(ctx, aConst)
			if err != nil {
				return "", err
			}
			bu.WriteString(limitValue)
		}
	}

	return bu.String(), nil
}

// formatWhereClauseNode handles various WHERE clause node types
func formatWhereClauseNode(ctx context.Context, node *pg_query.Node, indent int, conf *fmtconf.Config) (string, error) {
	switch n := node.Node.(type) {
	case *pg_query.Node_AExpr:
		return FormatAExpr(ctx, n, conf)
	case *pg_query.Node_BoolExpr:
		return formatBoolExprForFunc(ctx, n, indent, conf)
	case *pg_query.Node_NullTest:
		return FormatNullTest(ctx, n)
	case *pg_query.Node_SubLink:
		selectStmt, ok := n.SubLink.Subselect.Node.(*pg_query.Node_SelectStmt)
		if !ok {
			return "", fmt.Errorf("unsupported WHERE clause sublink node type")
		}
		if n.SubLink.SubLinkType != pg_query.SubLinkType_EXISTS_SUBLINK {
			return "", fmt.Errorf("unsupported WHERE clause sublink type")
		}
		subRes, err := FormatSelectStmtForFuncArg(ctx, selectStmt, indent+1, conf)
		if err != nil {
			return "", err
		}
		var bu strings.Builder
		bu.WriteString("EXISTS(\n")
		bu.WriteString(subRes)
		bu.WriteString("\n")
		for i := 0; i < indent; i++ {
			bu.WriteString(internal.GetIndent(conf))
		}
		bu.WriteString(")")
		return bu.String(), nil
	}
	return "", fmt.Errorf("unsupported WHERE clause node type")
}

// formatBoolExprForFunc handles BoolExpr nodes specifically for function arguments
func formatBoolExprForFunc(ctx context.Context, be *pg_query.Node_BoolExpr, subqueryIndent int, conf *fmtconf.Config) (string, error) {
	var bu strings.Builder

	for argI, arg := range be.BoolExpr.Args {
		if argI != 0 {
			bu.WriteString(" ")
			switch be.BoolExpr.Boolop {
			case pg_query.BoolExprType_AND_EXPR:
				bu.WriteString("AND")
			case pg_query.BoolExprType_OR_EXPR:
				bu.WriteString("OR")
			case pg_query.BoolExprType_NOT_EXPR:
				bu.WriteString("NOT")
			}
			bu.WriteString(" ")
		}

		switch n := arg.Node.(type) {
		case *pg_query.Node_AExpr:
			res, err := FormatAExpr(ctx, n, conf)
			if err != nil {
				return "", err
			}
			if be.BoolExpr.Boolop == pg_query.BoolExprType_NOT_EXPR {
				bu.WriteString("NOT (")
				bu.WriteString(res)
				bu.WriteString(")")
				continue
			}
			bu.WriteString(res)
		case *pg_query.Node_BoolExpr:
			res, err := formatBoolExprForFunc(ctx, n, subqueryIndent, conf)
			if err != nil {
				return "", err
			}
			if be.BoolExpr.Boolop == pg_query.BoolExprType_NOT_EXPR {
				bu.WriteString("NOT ")
			}
			bu.WriteString("(")
			bu.WriteString(res)
			bu.WriteString(")")
		case *pg_query.Node_NullTest:
			res, err := FormatNullTest(ctx, n)
			if err != nil {
				return "", err
			}
			if be.BoolExpr.Boolop == pg_query.BoolExprType_NOT_EXPR {
				bu.WriteString("NOT (")
				bu.WriteString(res)
				bu.WriteString(")")
				continue
			}
			bu.WriteString(res)
		case *pg_query.Node_ColumnRef:
			res, err := FormatColumnRefFields(ctx, n)
			if err != nil {
				return "", err
			}
			if be.BoolExpr.Boolop == pg_query.BoolExprType_NOT_EXPR {
				bu.WriteString("NOT (")
				bu.WriteString(res)
				bu.WriteString(")")
				continue
			}
			bu.WriteString(res)
		case *pg_query.Node_SubLink:
			selectStmt, ok := n.SubLink.Subselect.Node.(*pg_query.Node_SelectStmt)
			if !ok {
				return "", fmt.Errorf("unsupported sublink node type")
			}
			if n.SubLink.SubLinkType != pg_query.SubLinkType_EXISTS_SUBLINK {
				return "", fmt.Errorf("unsupported boolean expression sublink type")
			}
			res, err := FormatSelectStmtForFuncArg(ctx, selectStmt, subqueryIndent+1, conf)
			if err != nil {
				return "", err
			}
			if be.BoolExpr.Boolop == pg_query.BoolExprType_NOT_EXPR {
				bu.WriteString("NOT ")
			}
			bu.WriteString("EXISTS")
			bu.WriteString("(\n")
			bu.WriteString(res)
			bu.WriteString("\n")
			for i := 0; i < subqueryIndent; i++ {
				bu.WriteString(internal.GetIndent(conf))
			}
			bu.WriteString(")")
		default:
			return "", fmt.Errorf("unsupported boolean expression node type")
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
		bu.WriteString("FROM ")
		bu.WriteString(tName)
	case *pg_query.Node_RangeSubselect:
		// Handle subqueries in FROM clause
		if selectStmt, ok := n.RangeSubselect.Subquery.Node.(*pg_query.Node_SelectStmt); ok {
			subRes, err := FormatSelectStmtForFuncArg(ctx, selectStmt, indent+1, conf)
			if err != nil {
				return "", err
			}

			bu.WriteString("\n")
			for i := 0; i < indent; i++ {
				bu.WriteString(internal.GetIndent(conf))
			}
			bu.WriteString("FROM (")
			bu.WriteString("\n")
			bu.WriteString(subRes)
			bu.WriteString("\n")
			for i := 0; i < indent; i++ {
				bu.WriteString(internal.GetIndent(conf))
			}
			bu.WriteString(")")

			if n.RangeSubselect.Alias != nil {
				bu.WriteString(" ")
				bu.WriteString(n.RangeSubselect.Alias.Aliasname)
			}
		}
	case *pg_query.Node_JoinExpr:
		// Handle JOIN expressions
		joinRes, err := formatJoinExprForFunc(ctx, n, indent, conf)
		if err != nil {
			return "", err
		}
		bu.WriteString(joinRes)
	}

	return bu.String(), nil
}

// formatJoinExprForFunc handles JOIN expressions for function arguments
func formatJoinExprForFunc(ctx context.Context, join *pg_query.Node_JoinExpr, indent int, conf *fmtconf.Config) (string, error) {
	var bu strings.Builder

	// Left side of join
	if join.JoinExpr.Larg != nil {
		leftRes, err := FormatSelectStmtFromClauseForFuncArg(ctx, join.JoinExpr.Larg.Node, indent, conf)
		if err != nil {
			return "", err
		}
		bu.WriteString(leftRes)
	}

	// Join type and right side
	bu.WriteString("\n")
	for i := 0; i < indent+1; i++ {
		bu.WriteString(internal.GetIndent(conf))
	}

	switch join.JoinExpr.Jointype {
	case pg_query.JoinType_JOIN_INNER:
		bu.WriteString("INNER JOIN")
	case pg_query.JoinType_JOIN_LEFT:
		bu.WriteString("LEFT JOIN")
	case pg_query.JoinType_JOIN_RIGHT:
		bu.WriteString("RIGHT JOIN")
	case pg_query.JoinType_JOIN_FULL:
		bu.WriteString("FULL JOIN")
	default:
		bu.WriteString("JOIN")
	}

	if join.JoinExpr.Rarg != nil {
		if rangeVar, ok := join.JoinExpr.Rarg.Node.(*pg_query.Node_RangeVar); ok {
			bu.WriteString(" ")
			bu.WriteString(rangeVar.RangeVar.Relname)
			if rangeVar.RangeVar.Alias != nil {
				bu.WriteString(" ")
				bu.WriteString(rangeVar.RangeVar.Alias.Aliasname)
			}
		}
	}

	// Join condition
	if join.JoinExpr.Quals != nil {
		bu.WriteString("\n")
		for i := 0; i < indent+2; i++ {
			bu.WriteString(internal.GetIndent(conf))
		}
		bu.WriteString("ON ")

		qualRes, err := formatWhereClauseNode(ctx, join.JoinExpr.Quals, indent+2, conf)
		if err != nil {
			return "", err
		}
		bu.WriteString(qualRes)
	}

	return bu.String(), nil
}
