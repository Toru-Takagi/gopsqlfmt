package formatter

import (
	"context"
	"fmt"
	"strings"

	nodeformatter "github.com/Toru-Takagi/sql_formatter_go/formatter/node_formatter"

	pg_query "github.com/pganalyze/pg_query_go/v4"
)

const (
	namedParamPrefix = ":"
	npMarkPrefix     = "ttpre_"
)

func Format(sql string) (string, error) {
	ctx := context.Background()

	// support named parameter
	replacedSQL := strings.NewReplacer([]string{
		namedParamPrefix, npMarkPrefix,
	}...).Replace(sql)

	result, err := pg_query.Parse(replacedSQL)
	if err != nil {
		return "", err
	}
	var strBuilder strings.Builder
	strBuilder.WriteString("\n")
	for _, raw := range result.Stmts {
		switch internal := raw.Stmt.Node.(type) {
		case *pg_query.Node_SelectStmt:
			res, err := FormatSelectStmt(ctx, internal, 0)
			if err != nil {
				return "", err
			}
			strBuilder.WriteString(res)

		case *pg_query.Node_InsertStmt:
			strBuilder.WriteString("INSERT INTO")

			// output table name
			if internal.InsertStmt.Relation != nil {
				strBuilder.WriteString(" ")
				strBuilder.WriteString(internal.InsertStmt.Relation.Relname)
				strBuilder.WriteString("(")
			}

			// output column name
			for i, col := range internal.InsertStmt.Cols {
				if target, ok := col.Node.(*pg_query.Node_ResTarget); ok {
					if i != 0 {
						strBuilder.WriteString(",")
					}
					strBuilder.WriteString("\n")
					strBuilder.WriteString("\t")
					strBuilder.WriteString(target.ResTarget.Name)
				}
			}

			strBuilder.WriteString("\n")
			strBuilder.WriteString(") ")

			// output parameter
			if internal.InsertStmt.SelectStmt != nil {
				if sNode, ok := internal.InsertStmt.SelectStmt.Node.(*pg_query.Node_SelectStmt); ok {
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
									strBuilder.WriteString("\t")
									strBuilder.WriteString("$")
									strBuilder.WriteString(fmt.Sprint(v.ParamRef.Number))
								case *pg_query.Node_ColumnRef:
									strBuilder.WriteString("\n")
									strBuilder.WriteString("\t")
									for _, f := range v.ColumnRef.Fields {
										if s, ok := f.Node.(*pg_query.Node_String_); ok {
											strBuilder.WriteString(s.String_.Sval)
										}
									}
								case *pg_query.Node_FuncCall:
									for _, name := range v.FuncCall.Funcname {
										if s, ok := name.Node.(*pg_query.Node_String_); ok {
											if s.String_.Sval == "now" {
												strBuilder.WriteString("\n")
												strBuilder.WriteString("\t")
												strBuilder.WriteString("NOW(")
											}
											if s.String_.Sval == "gen_random_uuid" {
												strBuilder.WriteString("\n")
												strBuilder.WriteString("\t")
												strBuilder.WriteString("GEN_RANDOM_UUID(")
											}
											if s.String_.Sval == "current_setting" {
												strBuilder.WriteString("\n")
												strBuilder.WriteString("\t")
												strBuilder.WriteString("CURRENT_SETTING")
												strBuilder.WriteString("(")
											}
										}
									}
									for argI, arg := range v.FuncCall.Args {
										if argI != 0 {
											strBuilder.WriteString(",")
											strBuilder.WriteString(" ")
										}
										if a, ok := arg.Node.(*pg_query.Node_AConst); ok {
											res, err := nodeformatter.FormatAConst(ctx, a)
											if err != nil {
												return "", err
											}
											strBuilder.WriteString(res)
										}
										if paramRef, ok := arg.Node.(*pg_query.Node_ParamRef); ok {
											strBuilder.WriteString("$")
											strBuilder.WriteString(fmt.Sprint(paramRef.ParamRef.Number))
										}
										if cRef, ok := arg.Node.(*pg_query.Node_ColumnRef); ok {
											for fi, f := range cRef.ColumnRef.Fields {
												if s, ok := f.Node.(*pg_query.Node_String_); ok {
													if fi != 0 {
														strBuilder.WriteString(".")
													}
													strBuilder.WriteString(s.String_.Sval)
												}
											}
										}
									}
									strBuilder.WriteString(")")
								}
							}
						}
						strBuilder.WriteString("\n")
						strBuilder.WriteString(")")
					}

					res, err := FormatSelectStmt(ctx, sNode, 0)
					if err != nil {
						return "", err
					}
					strBuilder.WriteString(res)
				}
			}

			// output on conflict
			if internal.InsertStmt.OnConflictClause != nil {
				strBuilder.WriteString("\n")
				strBuilder.WriteString("ON CONFLICT")
				if internal.InsertStmt.OnConflictClause.Infer != nil {
					if len(internal.InsertStmt.OnConflictClause.Infer.IndexElems) > 0 {
						strBuilder.WriteString("(")
					}
					for _, elm := range internal.InsertStmt.OnConflictClause.Infer.IndexElems {
						if idxElm, ok := elm.Node.(*pg_query.Node_IndexElem); ok {
							strBuilder.WriteString(idxElm.IndexElem.Name)
						}
					}
					if len(internal.InsertStmt.OnConflictClause.Infer.IndexElems) > 0 {
						strBuilder.WriteString(")")
					}

					if internal.InsertStmt.OnConflictClause.Infer.Conname != "" {
						strBuilder.WriteString(" ")
						strBuilder.WriteString("ON CONSTRAINT")
						strBuilder.WriteString(" ")
						strBuilder.WriteString(internal.InsertStmt.OnConflictClause.Infer.Conname)
					}
				}

				switch internal.InsertStmt.OnConflictClause.Action {
				case pg_query.OnConflictAction_ONCONFLICT_NOTHING:
					strBuilder.WriteString("\n")
					strBuilder.WriteString("DO NOTHING")
				case pg_query.OnConflictAction_ONCONFLICT_UPDATE:
					strBuilder.WriteString("\n")
					strBuilder.WriteString("DO UPDATE SET")
				}
				for targetI, target := range internal.InsertStmt.OnConflictClause.TargetList {
					if res, ok := target.Node.(*pg_query.Node_ResTarget); ok {
						if targetI != 0 {
							strBuilder.WriteString(",")
						}
						strBuilder.WriteString("\n")
						strBuilder.WriteString("\t")
						strBuilder.WriteString(res.ResTarget.Name)
						strBuilder.WriteString(" = ")

						if res.ResTarget.Val != nil {
							switch n := res.ResTarget.Val.Node.(type) {
							case *pg_query.Node_ColumnRef:
								for fi, f := range n.ColumnRef.Fields {
									if s, ok := f.Node.(*pg_query.Node_String_); ok {
										if fi != 0 {
											strBuilder.WriteString(".")
										}
										if s.String_.Sval == "excluded" {
											strBuilder.WriteString("EXCLUDED")
										} else {
											strBuilder.WriteString(s.String_.Sval)
										}
									}
								}
							case *pg_query.Node_FuncCall:
								for _, name := range n.FuncCall.Funcname {
									if s, ok := name.Node.(*pg_query.Node_String_); ok {
										if s.String_.Sval == "now" {
											strBuilder.WriteString("NOW()")
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
	strBuilder.WriteString("\n")
	return strings.NewReplacer([]string{
		npMarkPrefix, namedParamPrefix,
	}...).Replace(strBuilder.String()), nil
}

func FormatSelectStmt(ctx context.Context, stmt *pg_query.Node_SelectStmt, indent int) (string, error) {
	if len(stmt.SelectStmt.TargetList) == 0 {
		return "", nil
	}

	var bu strings.Builder
	for i := 0; i < indent; i++ {
		bu.WriteString("\t")
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
				for fi, f := range n.ColumnRef.Fields {
					if s, ok := f.Node.(*pg_query.Node_String_); ok {
						if fi == 0 {
							bu.WriteString("\n\t")
							for i := 0; i < indent; i++ {
								bu.WriteString("\t")
							}
						} else {
							bu.WriteString(".")
						}
						bu.WriteString(s.String_.Sval)
					}
				}
			case *pg_query.Node_FuncCall:
				bu.WriteString("\n\t")
				for i := 0; i < indent; i++ {
					bu.WriteString("\t")
				}
				for _, name := range n.FuncCall.Funcname {
					if s, ok := name.Node.(*pg_query.Node_String_); ok {
						if s.String_.Sval == "count" {
							bu.WriteString("COUNT")
							bu.WriteString("(*")
						}
						if s.String_.Sval == "current_setting" {
							bu.WriteString("CURRENT_SETTING")
							bu.WriteString("(")
						}
						if s.String_.Sval == "set_config" {
							bu.WriteString("SET_CONFIG")
							bu.WriteString("(")
						}
						if s.String_.Sval == "array_agg" {
							bu.WriteString("ARRAY_AGG")
							bu.WriteString("(")
						}
					}
				}
				for argI, arg := range n.FuncCall.Args {
					if argI != 0 {
						bu.WriteString(",")
						bu.WriteString(" ")
					}
					if a, ok := arg.Node.(*pg_query.Node_AConst); ok {
						res, err := nodeformatter.FormatAConst(ctx, a)
						if err != nil {
							return "", err
						}
						bu.WriteString(res)
					}
					if paramRef, ok := arg.Node.(*pg_query.Node_ParamRef); ok {
						bu.WriteString("$")
						bu.WriteString(fmt.Sprint(paramRef.ParamRef.Number))
					}
					if cRef, ok := arg.Node.(*pg_query.Node_ColumnRef); ok {
						for fi, f := range cRef.ColumnRef.Fields {
							if s, ok := f.Node.(*pg_query.Node_String_); ok {
								if fi != 0 {
									bu.WriteString(".")
								}
								bu.WriteString(s.String_.Sval)
							}
						}
					}
				}
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
									bu.WriteString("\t")
								}
								for fi, f := range n.ColumnRef.Fields {
									if s, ok := f.Node.(*pg_query.Node_String_); ok {
										if fi != 0 {
											bu.WriteString(".")
										}
										bu.WriteString(s.String_.Sval)
									}
								}
								switch sortBy.SortBy.SortbyDir {
								case pg_query.SortByDir_SORTBY_ASC:
									bu.WriteString(" ASC")
								case pg_query.SortByDir_SORTBY_DESC:
									bu.WriteString(" DESC")
								}
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
					res, err := FormatSelectStmt(ctx, selectStmt, indent+2)
					if err != nil {
						return "", err
					}
					bu.WriteString("\n\t")
					bu.WriteString("(\n")
					bu.WriteString(res)
					bu.WriteString("\n\t")
					bu.WriteString(")")
				}
			}
			if res.ResTarget.Name != "" {
				bu.WriteString(" AS ")
				bu.WriteString(res.ResTarget.Name)
			}
		}
	}

	// output table name
	for _, node := range stmt.SelectStmt.FromClause {
		res, err := FormatSelectStmtFromClause(ctx, node.Node, indent)
		if err != nil {
			return "", err
		}
		bu.WriteString(res)
	}

	// output where clause
	if stmt.SelectStmt.WhereClause != nil {
		var (
			res string
			err error
		)
		if n, ok := stmt.SelectStmt.WhereClause.Node.(*pg_query.Node_AExpr); ok {
			res, err = nodeformatter.FormatAExpr(ctx, n)
		}
		if nBoolExpr, ok := stmt.SelectStmt.WhereClause.Node.(*pg_query.Node_BoolExpr); ok {
			res, err = formatBoolExpr(ctx, nBoolExpr, 0)
		}
		if err != nil {
			return "", err
		}
		bu.WriteString("\n")
		for i := 0; i < indent; i++ {
			bu.WriteString("\t")
		}
		bu.WriteString("WHERE")
		bu.WriteString(" ")
		bu.WriteString(res)
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
		if cRef, ok := gClause.Node.(*pg_query.Node_ColumnRef); ok {
			for fi, f := range cRef.ColumnRef.Fields {
				if s, ok := f.Node.(*pg_query.Node_String_); ok {
					if fi != 0 {
						bu.WriteString(".")
					}
					bu.WriteString(s.String_.Sval)
				}
			}
		}
	}

	// output sort clause
	if stmt.SelectStmt.SortClause != nil {
		bu.WriteString("\n")
		for i := 0; i < indent; i++ {
			bu.WriteString("\t")
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
							bu.WriteString("\t")
						}
						for fi, f := range n.ColumnRef.Fields {
							if s, ok := f.Node.(*pg_query.Node_String_); ok {
								if fi != 0 {
									bu.WriteString(".")
								}
								bu.WriteString(s.String_.Sval)
							}
						}
						switch sortBy.SortBy.SortbyDir {
						case pg_query.SortByDir_SORTBY_ASC:
							bu.WriteString(" ASC")
						case pg_query.SortByDir_SORTBY_DESC:
							bu.WriteString(" DESC")
						}
					case *pg_query.Node_FuncCall:
						for i := 0; i < indent; i++ {
							bu.WriteString("\t")
						}
						for _, name := range n.FuncCall.Funcname {
							if s, ok := name.Node.(*pg_query.Node_String_); ok {
								if s.String_.Sval == "min" {
									bu.WriteString("MIN")
									bu.WriteString("(")
								}
							}
						}
						for argI, arg := range n.FuncCall.Args {
							if argI != 0 {
								bu.WriteString(",")
								bu.WriteString(" ")
							}
							if a, ok := arg.Node.(*pg_query.Node_AConst); ok {
								res, err := nodeformatter.FormatAConst(ctx, a)
								if err != nil {
									return "", err
								}
								bu.WriteString(res)
							}
							if paramRef, ok := arg.Node.(*pg_query.Node_ParamRef); ok {
								bu.WriteString("$")
								bu.WriteString(fmt.Sprint(paramRef.ParamRef.Number))
							}
							if cRef, ok := arg.Node.(*pg_query.Node_ColumnRef); ok {
								for fi, f := range cRef.ColumnRef.Fields {
									if s, ok := f.Node.(*pg_query.Node_String_); ok {
										if fi != 0 {
											bu.WriteString(".")
										}
										bu.WriteString(s.String_.Sval)
									}
								}
							}
						}
						bu.WriteString(")")
					}
				}
			}
		}
	}

	// output limit clause
	if stmt.SelectStmt.LimitCount != nil {
		bu.WriteString("\n")
		for i := 0; i < indent; i++ {
			bu.WriteString("\t")
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
				bu.WriteString("\t")
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

func FormatSelectStmtFromClause(ctx context.Context, node any, indent int) (string, error) {
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
			bu.WriteString("\t")
		}
		bu.WriteString("FROM")
		bu.WriteString(" ")
		bu.WriteString(tableName)
	case *pg_query.Node_RangeFunction:
		bu.WriteString("\n")
		for i := 0; i < indent; i++ {
			bu.WriteString("\t")
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
	case *pg_query.Node_JoinExpr:
		res, err := FormatSelectStmtFromClause(ctx, n.JoinExpr.Larg.Node, indent)
		if err != nil {
			return "", nil
		}
		bu.WriteString(res)

		if nRangeVar, ok := n.JoinExpr.Rarg.Node.(*pg_query.Node_RangeVar); ok {
			switch n.JoinExpr.Jointype {
			case pg_query.JoinType_JOIN_INNER:
				bu.WriteString("\nINNER JOIN ")
			case pg_query.JoinType_JOIN_LEFT:
				bu.WriteString("\nLEFT JOIN ")
			}
			tableName, err := formatTableName(ctx, nRangeVar)
			if err != nil {
				return "", err
			}
			bu.WriteString(tableName)
			bu.WriteString(" ON ")
		}

		switch qualsNode := n.JoinExpr.Quals.Node.(type) {
		case *pg_query.Node_AExpr:
			res, err := nodeformatter.FormatAExpr(ctx, qualsNode)
			if err != nil {
				return "", err
			}
			bu.WriteString(res)
		case *pg_query.Node_BoolExpr:
			res, err := formatBoolExpr(ctx, qualsNode, 0)
			if err != nil {
				return "", err
			}
			bu.WriteString(res)
		}
	}

	return bu.String(), nil
}
