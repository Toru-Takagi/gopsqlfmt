package formatter

import (
	"context"
	"fmt"
	"strings"

	"github.com/Toru-Takagi/gopsqlfmt/fmtconf"
	"github.com/Toru-Takagi/gopsqlfmt/formatter/enumconv"
	"github.com/Toru-Takagi/gopsqlfmt/formatter/internal"
	nodeformatter "github.com/Toru-Takagi/gopsqlfmt/formatter/node_formatter"

	pg_query "github.com/pganalyze/pg_query_go/v6"
)

const (
	castParamPrefix  = "::"
	namedParamPrefix = ":"
	cpMarkPrefix     = "castpre_"
	npMarkPrefix     = "ttpre_"
)

func Format(sql string, conf *fmtconf.Config) (string, error) {
	ctx := context.Background()
	if conf == nil {
		conf = fmtconf.NewDefaultConfig()
	}

	// support named parameter
	replacedSQL := strings.NewReplacer([]string{
		castParamPrefix, cpMarkPrefix,
		namedParamPrefix, npMarkPrefix,
	}...).Replace(sql)
	replacedSQL = strings.NewReplacer([]string{
		cpMarkPrefix, castParamPrefix,
	}...).Replace(replacedSQL)

	result, err := pg_query.Parse(replacedSQL)
	if err != nil {
		return "", err
	}
	var strBuilder strings.Builder
	strBuilder.WriteString("\n")
	for _, raw := range result.Stmts {
		switch stmt := raw.Stmt.Node.(type) {
		case *pg_query.Node_SelectStmt:
			res, err := FormatSelectStmt(ctx, stmt, 0, conf)
			if err != nil {
				return "", err
			}
			strBuilder.WriteString(res)

		case *pg_query.Node_InsertStmt:
			strBuilder.WriteString("INSERT INTO")

			// output table name
			tableName, err := nodeformatter.FormatRelation(ctx, stmt.InsertStmt.Relation)
			if err != nil {
				return "", err
			}
			strBuilder.WriteString(tableName)

			if len(stmt.InsertStmt.Cols) > 0 {
				strBuilder.WriteString("(")

				// output column name
				for i, col := range stmt.InsertStmt.Cols {
					if target, ok := col.Node.(*pg_query.Node_ResTarget); ok {
						if i != 0 {
							strBuilder.WriteString(",")
						}
						strBuilder.WriteString("\n")
						strBuilder.WriteString(internal.GetIndent(conf))
						strBuilder.WriteString(target.ResTarget.Name)
					}
				}

				strBuilder.WriteString("\n")
				strBuilder.WriteString(") ")
			} else {
				strBuilder.WriteString("\n")
			}

			// output parameter
			if stmt.InsertStmt.SelectStmt != nil {
				if sNode, ok := stmt.InsertStmt.SelectStmt.Node.(*pg_query.Node_SelectStmt); ok {
					for _, value := range sNode.SelectStmt.ValuesLists {
						strBuilder.WriteString("VALUES (")
						if list, ok := value.Node.(*pg_query.Node_List); ok {
							for itemI, item := range list.List.Items {
								if itemI != 0 {
									strBuilder.WriteString(",")
								}
								switch v := item.Node.(type) {
								case *pg_query.Node_ParamRef:
									strBuilder.WriteString("\n")
									strBuilder.WriteString(internal.GetIndent(conf))
									strBuilder.WriteString("$")
									strBuilder.WriteString(fmt.Sprint(v.ParamRef.Number))
								case *pg_query.Node_ColumnRef:
									strBuilder.WriteString("\n")
									strBuilder.WriteString(internal.GetIndent(conf))
									field, err := nodeformatter.FormatColumnRefFields(ctx, v)
									if err != nil {
										return "", err
									}
									strBuilder.WriteString(field)
								case *pg_query.Node_AConst:
									aconst, err := nodeformatter.FormatAConst(ctx, v)
									if err != nil {
										return "", err
									}
									strBuilder.WriteString("\n")
									strBuilder.WriteString(internal.GetIndent(conf))
									strBuilder.WriteString(aconst)
								case *pg_query.Node_FuncCall:
									funcName, err := nodeformatter.FormatFuncname(ctx, v, conf)
									if err != nil {
										return "", err
									}
									strBuilder.WriteString("\n")
									strBuilder.WriteString(internal.GetIndent(conf))
									strBuilder.WriteString(funcName)
									strBuilder.WriteString("(")

									arg, err := nodeformatter.FormatFuncCallArgs(ctx, v, 0, conf)
									if err != nil {
										return "", err
									}
									strBuilder.WriteString(arg)

									strBuilder.WriteString(")")
								}
							}
						}
						strBuilder.WriteString("\n")
						strBuilder.WriteString(")")
					}

					res, err := FormatSelectStmt(ctx, sNode, 0, conf)
					if err != nil {
						return "", err
					}
					strBuilder.WriteString(res)
				}
			}

			// output on conflict
			if stmt.InsertStmt.OnConflictClause != nil {
				strBuilder.WriteString("\n")
				strBuilder.WriteString("ON CONFLICT")
				if stmt.InsertStmt.OnConflictClause.Infer != nil {
					if len(stmt.InsertStmt.OnConflictClause.Infer.IndexElems) > 0 {
						strBuilder.WriteString("(")
					}
					for i, elm := range stmt.InsertStmt.OnConflictClause.Infer.IndexElems {
						if idxElm, ok := elm.Node.(*pg_query.Node_IndexElem); ok {
							if i > 0 {
								strBuilder.WriteString(", ")
							}
							strBuilder.WriteString(idxElm.IndexElem.Name)
						}
					}
					if len(stmt.InsertStmt.OnConflictClause.Infer.IndexElems) > 0 {
						strBuilder.WriteString(")")
					}

					if stmt.InsertStmt.OnConflictClause.Infer.Conname != "" {
						strBuilder.WriteString(" ")
						strBuilder.WriteString("ON CONSTRAINT")
						strBuilder.WriteString(" ")
						strBuilder.WriteString(stmt.InsertStmt.OnConflictClause.Infer.Conname)
					}
				}

				switch stmt.InsertStmt.OnConflictClause.Action {
				case pg_query.OnConflictAction_ONCONFLICT_NOTHING:
					strBuilder.WriteString("\n")
					strBuilder.WriteString("DO NOTHING")
				case pg_query.OnConflictAction_ONCONFLICT_UPDATE:
					strBuilder.WriteString("\n")
					strBuilder.WriteString("DO UPDATE SET")
				}
				for targetI, target := range stmt.InsertStmt.OnConflictClause.TargetList {
					if res, ok := target.Node.(*pg_query.Node_ResTarget); ok {
						if targetI != 0 {
							strBuilder.WriteString(",")
						}
						strBuilder.WriteString("\n")
						strBuilder.WriteString(internal.GetIndent(conf))
						strBuilder.WriteString(res.ResTarget.Name)
						strBuilder.WriteString(" = ")

						if res.ResTarget.Val != nil {
							switch n := res.ResTarget.Val.Node.(type) {
							case *pg_query.Node_ColumnRef:
								field, err := nodeformatter.FormatColumnRefFields(ctx, n)
								if err != nil {
									return "", err
								}
								strBuilder.WriteString(field)
							case *pg_query.Node_FuncCall:
								res, err := nodeformatter.FormatFuncname(ctx, n, conf)
								if err != nil {
									return "", err
								}
								strBuilder.WriteString(res)
								strBuilder.WriteString("(")
								strBuilder.WriteString(")")
							}
						}
					}
				}
			}
		case *pg_query.Node_UpdateStmt:
			strBuilder.WriteString("UPDATE")

			// output table name
			tableName, err := nodeformatter.FormatRelation(ctx, stmt.UpdateStmt.Relation)
			if err != nil {
				return "", err
			}
			strBuilder.WriteString(tableName)

			strBuilder.WriteString("\n")
			strBuilder.WriteString("SET")

			for targetI, target := range stmt.UpdateStmt.TargetList {
				if res, ok := target.Node.(*pg_query.Node_ResTarget); ok {
					if targetI != 0 {
						strBuilder.WriteString(",")
					}
					strBuilder.WriteString("\n")
					strBuilder.WriteString(internal.GetIndent(conf))
					strBuilder.WriteString(res.ResTarget.Name)
					strBuilder.WriteString(" = ")

					if res.ResTarget.Val != nil {
						switch n := res.ResTarget.Val.Node.(type) {
						case *pg_query.Node_ColumnRef:
							field, err := nodeformatter.FormatColumnRefFields(ctx, n)
							if err != nil {
								return "", err
							}
							strBuilder.WriteString(field)
						case *pg_query.Node_ParamRef:
							strBuilder.WriteString("$")
							strBuilder.WriteString(fmt.Sprint(n.ParamRef.Number))
						case *pg_query.Node_FuncCall:
							res, err := nodeformatter.FormatFuncname(ctx, n, conf)
							if err != nil {
								return "", err
							}
							strBuilder.WriteString(res)
							strBuilder.WriteString("(")
							strBuilder.WriteString(")")
						}
					}
				}
			}

			// output where clause
			if stmt.UpdateStmt.WhereClause != nil {
				var (
					res string
					err error
				)
				if n, ok := stmt.UpdateStmt.WhereClause.Node.(*pg_query.Node_AExpr); ok {
					res, err = nodeformatter.FormatAExpr(ctx, n, conf)
				}
				if nBoolExpr, ok := stmt.UpdateStmt.WhereClause.Node.(*pg_query.Node_BoolExpr); ok {
					res, err = formatBoolExpr(ctx, nBoolExpr, 0, conf)
				}
				if nNullTest, ok := stmt.UpdateStmt.WhereClause.Node.(*pg_query.Node_NullTest); ok {
					res, err = nodeformatter.FormatNullTest(ctx, nNullTest)
				}
				if err != nil {
					return "", err
				}
				strBuilder.WriteString("\n")
				strBuilder.WriteString("WHERE")
				strBuilder.WriteString(" ")
				strBuilder.WriteString(res)
			}
		case *pg_query.Node_DeleteStmt:
			strBuilder.WriteString("DELETE FROM")

			// output table name
			tableName, err := nodeformatter.FormatRelation(ctx, stmt.DeleteStmt.Relation)
			if err != nil {
				return "", err
			}
			strBuilder.WriteString(tableName)

			// output where clause
			if stmt.DeleteStmt.WhereClause != nil {
				var (
					res string
					err error
				)
				if n, ok := stmt.DeleteStmt.WhereClause.Node.(*pg_query.Node_AExpr); ok {
					res, err = nodeformatter.FormatAExpr(ctx, n, conf)
				}
				if nBoolExpr, ok := stmt.DeleteStmt.WhereClause.Node.(*pg_query.Node_BoolExpr); ok {
					res, err = formatBoolExpr(ctx, nBoolExpr, 0, conf)
				}
				if nNullTest, ok := stmt.DeleteStmt.WhereClause.Node.(*pg_query.Node_NullTest); ok {
					res, err = nodeformatter.FormatNullTest(ctx, nNullTest)
				}
				if err != nil {
					return "", err
				}
				strBuilder.WriteString("\n")
				strBuilder.WriteString("WHERE")
				strBuilder.WriteString(" ")
				strBuilder.WriteString(res)
			}
		}
	}
	strBuilder.WriteString("\n")
	return strings.NewReplacer([]string{
		npMarkPrefix, namedParamPrefix,
	}...).Replace(strBuilder.String()), nil
}

func FormatSelectStmt(ctx context.Context, stmt *pg_query.Node_SelectStmt, indent int, conf *fmtconf.Config) (string, error) {
	// Handle set operations first, before checking target list
	if stmt.SelectStmt.Op != pg_query.SetOperation_SETOP_NONE {
		var bu strings.Builder

		if stmt.SelectStmt.Larg != nil {
			leftStmt := &pg_query.Node_SelectStmt{SelectStmt: stmt.SelectStmt.Larg}
			leftRes, err := FormatSelectStmt(ctx, leftStmt, indent, conf)
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
			rightRes, err := FormatSelectStmt(ctx, rightStmt, indent, conf)
			if err != nil {
				return "", err
			}
			bu.WriteString("\n")
			bu.WriteString(rightRes)
		}

		return bu.String(), nil
	}

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
				field, err := nodeformatter.FormatColumnRefFields(ctx, n)
				if err != nil {
					return "", err
				}
				bu.WriteString("\n")
				bu.WriteString(internal.GetIndent(conf))
				for i := 0; i < indent; i++ {
					bu.WriteString(internal.GetIndent(conf))
				}
				bu.WriteString(field)
			case *pg_query.Node_FuncCall:
				bu.WriteString("\n")
				bu.WriteString(internal.GetIndent(conf))
				for i := 0; i < indent; i++ {
					bu.WriteString(internal.GetIndent(conf))
				}
				funcName, err := nodeformatter.FormatFuncname(ctx, n, conf)
				if err != nil {
					return "", err
				}
				bu.WriteString(funcName)
				bu.WriteString("(")

				arg, err := nodeformatter.FormatFuncCallArgs(ctx, n, indent+1, conf)
				if err != nil {
					return "", err
				}
				bu.WriteString(arg)

				for sortI, order := range n.FuncCall.AggOrder {
					if sortI == 0 {
						bu.WriteString(" ")
						bu.WriteString("ORDER BY")
						bu.WriteString(" ")
					}
					if sortBy, ok := order.Node.(*pg_query.Node_SortBy); ok {
						if sortBy.SortBy.Node != nil {
							switch n := sortBy.SortBy.Node.Node.(type) {
							case *pg_query.Node_ColumnRef:
								if sortI != 0 {
									bu.WriteString(",")
									bu.WriteString("\n")
									bu.WriteString(internal.GetIndent(conf))
								}
								field, err := nodeformatter.FormatColumnRefFields(ctx, n)
								if err != nil {
									return "", err
								}
								bu.WriteString(field)
								sortBy, err := nodeformatter.FormatSortByDir(ctx, sortBy)
								if err != nil {
									return "", err
								}
								bu.WriteString(sortBy)
							}
						}
					}
				}
				bu.WriteString(")")
				if n.FuncCall.Over != nil {
					bu.WriteString(" OVER()")
				}
			case *pg_query.Node_SubLink:
				if selectStmt, ok := n.SubLink.Subselect.Node.(*pg_query.Node_SelectStmt); ok {
					res, err := FormatSelectStmt(ctx, selectStmt, indent+2, conf)
					if err != nil {
						return "", err
					}
					bu.WriteString("\n")
					bu.WriteString(internal.GetIndent(conf))

					slt, err := enumconv.SubLinkTypeToString(n.SubLink.SubLinkType)
					if err != nil {
						return "", err
					}
					bu.WriteString(slt)

					bu.WriteString("(\n")
					bu.WriteString(res)
					bu.WriteString("\n")
					bu.WriteString(internal.GetIndent(conf))
					bu.WriteString(")")
				}

			case *pg_query.Node_CoalesceExpr:
				bu.WriteString("\n")
				bu.WriteString(internal.GetIndent(conf))
				bu.WriteString("COALESCE")
				bu.WriteString("(")

				for argI, arg := range n.CoalesceExpr.Args {
					if argI != 0 {
						bu.WriteString(",")
						bu.WriteString(" ")
					}
					switch n := arg.Node.(type) {
					case *pg_query.Node_ColumnRef:
						field, err := nodeformatter.FormatColumnRefFields(ctx, n)
						if err != nil {
							return "", err
						}
						bu.WriteString(field)
					case *pg_query.Node_SubLink:
						if selectStmt, ok := n.SubLink.Subselect.Node.(*pg_query.Node_SelectStmt); ok {
							res, err := FormatSelectStmt(ctx, selectStmt, indent+2, conf)
							if err != nil {
								return "", err
							}
							bu.WriteString("(\n")
							bu.WriteString(res)
							bu.WriteString("\n")
							bu.WriteString(internal.GetIndent(conf))
							bu.WriteString(")")
						}
					case *pg_query.Node_AConst:
						aconst, err := nodeformatter.FormatAConst(ctx, n)
						if err != nil {
							return "", err
						}
						bu.WriteString(aconst)
					case *pg_query.Node_FuncCall:
						funcName, err := nodeformatter.FormatFuncname(ctx, n, conf)
						if err != nil {
							return "", err
						}
						bu.WriteString(funcName)
						bu.WriteString("(")

						arg, err := nodeformatter.FormatFuncCallArgs(ctx, n, indent+1, conf)
						if err != nil {
							return "", err
						}
						bu.WriteString(arg)
						bu.WriteString(")")
					}
				}

				bu.WriteString(")")
			case *pg_query.Node_TypeCast:
				bu.WriteString("\n")
				bu.WriteString(internal.GetIndent(conf))
				for i := 0; i < indent; i++ {
					bu.WriteString(internal.GetIndent(conf))
				}
				tc, err := nodeformatter.FormatTypeCast(ctx, n)
				if err != nil {
					return "", err
				}
				bu.WriteString(tc)
			case *pg_query.Node_CaseExpr:
				bu.WriteString("\n")
				bu.WriteString(internal.GetIndent(conf))
				for i := 0; i < indent; i++ {
					bu.WriteString(internal.GetIndent(conf))
				}
				caseExpr, err := nodeformatter.FormatCaseExpr(ctx, n, indent, conf)
				if err != nil {
					return "", err
				}
				bu.WriteString(caseExpr)
			case *pg_query.Node_AConst:
				aconst, err := nodeformatter.FormatAConst(ctx, n)
				if err != nil {
					return "", err
				}
				bu.WriteString("\n")
				bu.WriteString(internal.GetIndent(conf))
				for i := 0; i < indent; i++ {
					bu.WriteString(internal.GetIndent(conf))
				}
				bu.WriteString(aconst)
			}
			if res.ResTarget.Name != "" {
				bu.WriteString(" AS ")
				bu.WriteString(res.ResTarget.Name)
			}
		}
	}

	// output table name
	for _, node := range stmt.SelectStmt.FromClause {
		res, err := FormatSelectStmtFromClause(ctx, node.Node, indent, conf)
		if err != nil {
			return "", err
		}
		bu.WriteString(res)
	}

	// output where clause
	if stmt.SelectStmt.WhereClause != nil {
		switch n := stmt.SelectStmt.WhereClause.Node.(type) {
		case *pg_query.Node_AExpr:
			res, err := nodeformatter.FormatAExpr(ctx, n, conf)
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
		case *pg_query.Node_BoolExpr:
			res, err := formatBoolExpr(ctx, n, indent, conf)
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
		case *pg_query.Node_NullTest:
			res, err := nodeformatter.FormatNullTest(ctx, n)
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
		case *pg_query.Node_SubLink:
			if selectStmt, ok := n.SubLink.Subselect.Node.(*pg_query.Node_SelectStmt); ok {
				res, err := FormatSelectStmt(ctx, selectStmt, indent+1, conf)
				if err != nil {
					return "", err
				}

				bu.WriteString("\n")
				for i := 0; i < indent; i++ {
					bu.WriteString(internal.GetIndent(conf))
				}
				bu.WriteString("WHERE")
				bu.WriteString(" ")

				slt, err := enumconv.SubLinkTypeToString(n.SubLink.SubLinkType)
				if err != nil {
					return "", err
				}
				bu.WriteString(slt)

				bu.WriteString("(\n")
				bu.WriteString(res)
				bu.WriteString("\n")
				bu.WriteString(")")
			}
		}
	}

	// output group clause
	for gIndex, gClause := range stmt.SelectStmt.GroupClause {
		if gIndex == 0 {
			bu.WriteString("\n")
			bu.WriteString("GROUP BY")
			bu.WriteString(" ")
		} else {
			bu.WriteString(",")
			bu.WriteString(" ")
		}
		switch node := gClause.Node.(type) {
		case *pg_query.Node_ColumnRef:
			field, err := nodeformatter.FormatColumnRefFields(ctx, node)
			if err != nil {
				return "", err
			}
			bu.WriteString(field)
		case *pg_query.Node_TypeCast:
			tc, err := nodeformatter.FormatTypeCast(ctx, node)
			if err != nil {
				return "", err
			}
			bu.WriteString(tc)
		}
	}

	// output sort clause
	if stmt.SelectStmt.SortClause != nil {
		bu.WriteString("\n")
		for i := 0; i < indent; i++ {
			bu.WriteString(internal.GetIndent(conf))
		}
		bu.WriteString("ORDER BY")
		bu.WriteString(" ")
		for sortI, node := range stmt.SelectStmt.SortClause {
			if sortBy, ok := node.Node.(*pg_query.Node_SortBy); ok {
				if sortBy.SortBy.Node != nil {
					switch n := sortBy.SortBy.Node.Node.(type) {
					case *pg_query.Node_ColumnRef:
						if sortI != 0 {
							bu.WriteString(",")
							bu.WriteString("\n")
							bu.WriteString(internal.GetIndent(conf))
						}
						field, err := nodeformatter.FormatColumnRefFields(ctx, n)
						if err != nil {
							return "", err
						}
						bu.WriteString(field)

						sortBy, err := nodeformatter.FormatSortByDir(ctx, sortBy)
						if err != nil {
							return "", err
						}
						bu.WriteString(sortBy)
					case *pg_query.Node_TypeCast:
						if sortI != 0 {
							bu.WriteString(",")
							bu.WriteString("\n")
							bu.WriteString(internal.GetIndent(conf))
						}
						tc, err := nodeformatter.FormatTypeCast(ctx, n)
						if err != nil {
							return "", err
						}
						bu.WriteString(tc)

						sortBy, err := nodeformatter.FormatSortByDir(ctx, sortBy)
						if err != nil {
							return "", err
						}
						bu.WriteString(sortBy)
					case *pg_query.Node_FuncCall:
						for i := 0; i < indent; i++ {
							bu.WriteString(internal.GetIndent(conf))
						}
						funcName, err := nodeformatter.FormatFuncname(ctx, n, conf)
						if err != nil {
							return "", err
						}
						bu.WriteString(funcName)
						bu.WriteString("(")

						arg, err := nodeformatter.FormatFuncCallArgs(ctx, n, indent, conf)
						if err != nil {
							return "", err
						}
						bu.WriteString(arg)

						bu.WriteString(")")
					case *pg_query.Node_CoalesceExpr:
						if sortI != 0 {
							bu.WriteString(",")
							bu.WriteString("\n")
							bu.WriteString(internal.GetIndent(conf))
						}
						bu.WriteString("COALESCE")
						bu.WriteString("(")

						for argI, arg := range n.CoalesceExpr.Args {
							if argI != 0 {
								bu.WriteString(",")
								bu.WriteString(" ")
							}
							switch argNode := arg.Node.(type) {
							case *pg_query.Node_ColumnRef:
								field, err := nodeformatter.FormatColumnRefFields(ctx, argNode)
								if err != nil {
									return "", err
								}
								bu.WriteString(field)
							case *pg_query.Node_AConst:
								aconst, err := nodeformatter.FormatAConst(ctx, argNode)
								if err != nil {
									return "", err
								}
								bu.WriteString(aconst)
							}
						}

						bu.WriteString(")")

						sortBy, err := nodeformatter.FormatSortByDir(ctx, sortBy)
						if err != nil {
							return "", err
						}
						bu.WriteString(sortBy)
					}
				}
			}
		}
	}

	// output limit clause
	if stmt.SelectStmt.LimitCount != nil {
		bu.WriteString("\n")
		for i := 0; i < indent; i++ {
			bu.WriteString(internal.GetIndent(conf))
		}
		bu.WriteString("LIMIT")
		bu.WriteString(" ")

		if aCount, ok := stmt.SelectStmt.LimitCount.Node.(*pg_query.Node_AConst); ok {
			res, err := nodeformatter.FormatAConst(ctx, aCount)
			if err != nil {
				return "", err
			}
			bu.WriteString(res)
		}
	}

	for _, clause := range stmt.SelectStmt.LockingClause {
		if locking, ok := clause.Node.(*pg_query.Node_LockingClause); ok {
			bu.WriteString("\n")
			for i := 0; i < indent; i++ {
				bu.WriteString(internal.GetIndent(conf))
			}
			switch locking.LockingClause.Strength {
			case pg_query.LockClauseStrength_LCS_FORUPDATE:
				bu.WriteString("FOR UPDATE")
			}

			switch locking.LockingClause.WaitPolicy {
			case pg_query.LockWaitPolicy_LockWaitSkip:
				bu.WriteString(" SKIP LOCKED")
			}
		}
	}

	return bu.String(), nil
}

func FormatSelectStmtFromClause(ctx context.Context, node any, indent int, conf *fmtconf.Config) (string, error) {
	var bu strings.Builder

	formatTableName := func(ctx context.Context, n *pg_query.Node_RangeVar) (string, error) {
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
		return tName, nil
	}

	switch n := node.(type) {
	case *pg_query.Node_RangeVar:
		tableName, err := formatTableName(ctx, n)
		if err != nil {
			return "", err
		}
		bu.WriteString("\n")
		for i := 0; i < indent; i++ {
			bu.WriteString(internal.GetIndent(conf))
		}
		bu.WriteString("FROM")
		bu.WriteString(" ")
		bu.WriteString(tableName)
	case *pg_query.Node_RangeFunction:
		bu.WriteString("\n")
		for i := 0; i < indent; i++ {
			bu.WriteString(internal.GetIndent(conf))
		}
		bu.WriteString("FROM")
		bu.WriteString(" ")
		for _, fn := range n.RangeFunction.Functions {
			if list, ok := fn.Node.(*pg_query.Node_List); ok {
				for _, item := range list.List.Items {
					if sqlvalueFunc, ok := item.Node.(*pg_query.Node_SqlvalueFunction); ok {
						if sqlvalueFunc.SqlvalueFunction.Op == pg_query.SQLValueFunctionOp_SVFOP_USER {
							bu.WriteString("user")
						}
					}
				}
			}
		}
		if n.RangeFunction.Alias != nil {
			bu.WriteString(" ")
			bu.WriteString(n.RangeFunction.Alias.Aliasname)
		}
	case *pg_query.Node_RangeSubselect:
		bu.WriteString("\n")
		for i := 0; i < indent; i++ {
			bu.WriteString(internal.GetIndent(conf))
		}
		bu.WriteString("FROM")
		bu.WriteString(" ")

		if selectStmt, ok := n.RangeSubselect.Subquery.Node.(*pg_query.Node_SelectStmt); ok {
			bu.WriteString("(\n")
			res, err := FormatSelectStmt(ctx, selectStmt, indent+1, conf)
			if err != nil {
				return "", err
			}
			bu.WriteString(res)
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
		res, err := FormatSelectStmtFromClause(ctx, n.JoinExpr.Larg.Node, indent, conf)
		if err != nil {
			return "", nil
		}
		bu.WriteString(res)

		switch nRarg := n.JoinExpr.Rarg.Node.(type) {
		case *pg_query.Node_RangeVar:
			if conf.Join.StartIndentType == fmtconf.JOIN_START_INDENT_TYPE_ONE_SPACE {
				bu.WriteString("\n")
				bu.WriteString(internal.GetIndent(conf))
			} else {
				bu.WriteString("\n")
			}

			jt, err := enumconv.JoinTypeToString(n.JoinExpr.Jointype)
			if err != nil {
				return "", err
			}
			bu.WriteString(jt)
			bu.WriteString(" ")

			tableName, err := formatTableName(ctx, nRarg)
			if err != nil {
				return "", err
			}

			bu.WriteString(tableName)
		case *pg_query.Node_RangeSubselect:
			switch nSubquery := nRarg.RangeSubselect.Subquery.Node.(type) {
			case *pg_query.Node_SelectStmt:
				if conf.Join.StartIndentType == fmtconf.JOIN_START_INDENT_TYPE_ONE_SPACE {
					bu.WriteString("\n")
					bu.WriteString(internal.GetIndent(conf))
				} else {
					bu.WriteString("\n")
				}

				jt, err := enumconv.JoinTypeToString(n.JoinExpr.Jointype)
				if err != nil {
					return "", err
				}
				bu.WriteString(jt)
				bu.WriteString(" ")

				bu.WriteString("(")
				bu.WriteString("\n")

				res, err := FormatSelectStmt(ctx, nSubquery, indent+2, conf)
				if err != nil {
					return "", err
				}
				bu.WriteString(res)
			}
			// TODO: Alias
		}

		for _, u := range n.JoinExpr.UsingClause {
			bu.WriteString(" ")
			bu.WriteString("USING")
			bu.WriteString("(")
			switch n := u.Node.(type) {
			case *pg_query.Node_String_:
				bu.WriteString(n.String_.Sval)
			}
			bu.WriteString(")")
		}

		if n.JoinExpr.Quals != nil {
			if conf.Join.LineBreakType == fmtconf.JOIN_LINE_BREAK_ON_CLAUSE {
				bu.WriteString("\n")
				bu.WriteString(internal.GetIndent(conf))
				if conf.Join.StartIndentType == fmtconf.JOIN_START_INDENT_TYPE_ONE_SPACE {
					bu.WriteString(internal.GetIndent(conf))
				}
			} else {
				bu.WriteString(" ")
			}

			bu.WriteString("ON")
			bu.WriteString(" ")

			switch qualsNode := n.JoinExpr.Quals.Node.(type) {
			case *pg_query.Node_AExpr:
				res, err := nodeformatter.FormatAExpr(ctx, qualsNode, conf)
				if err != nil {
					return "", err
				}
				bu.WriteString(res)
			case *pg_query.Node_BoolExpr:
				res, err := formatBoolExpr(ctx, qualsNode, 0, conf)
				if err != nil {
					return "", err
				}
				bu.WriteString(res)
			}
		}
	}

	return bu.String(), nil
}
