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
			res, err := FormatSelectStmt(ctx, internal)
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
								case *pg_query.Node_FuncCall:
									for _, name := range v.FuncCall.Funcname {
										if s, ok := name.Node.(*pg_query.Node_String_); ok {
											if s.String_.Sval == "now" {
												strBuilder.WriteString("\n")
												strBuilder.WriteString("\t")
												strBuilder.WriteString("NOW()")
											}
										}
									}
								}
							}
						}
						strBuilder.WriteString("\n")
						strBuilder.WriteString(")")
					}

					res, err := FormatSelectStmt(ctx, sNode)
					if err != nil {
						return "", err
					}
					strBuilder.WriteString(res)
				}
			}
		}
	}
	strBuilder.WriteString("\n")
	return strings.NewReplacer([]string{
		npMarkPrefix, namedParamPrefix,
	}...).Replace(strBuilder.String()), nil
}

func FormatSelectStmt(ctx context.Context, stmt *pg_query.Node_SelectStmt) (string, error) {
	if len(stmt.SelectStmt.TargetList) == 0 {
		return "", nil
	}

	var bu strings.Builder
	bu.WriteString("SELECT")

	// output column name
	for ti, node := range stmt.SelectStmt.TargetList {
		if res, ok := node.Node.(*pg_query.Node_ResTarget); ok {
			if n, ok := res.ResTarget.Val.Node.(*pg_query.Node_ColumnRef); ok {
				for fi, f := range n.ColumnRef.Fields {
					if s, ok := f.Node.(*pg_query.Node_String_); ok {
						if fi == 0 {
							if ti != 0 {
								bu.WriteString(",")
							}
							bu.WriteString("\n\t")
						} else {
							bu.WriteString(".")
						}
						bu.WriteString(s.String_.Sval)
					}
				}
			}
			if funcCall, ok := res.ResTarget.Val.Node.(*pg_query.Node_FuncCall); ok {
				bu.WriteString("\n\t")
				for _, name := range funcCall.FuncCall.Funcname {
					if s, ok := name.Node.(*pg_query.Node_String_); ok {
						if s.String_.Sval == "count" {
							bu.WriteString("COUNT(*)")
						}
					}
				}
				if funcCall.FuncCall.Over != nil {
					bu.WriteString(" OVER()")
				}
			}
			if res.ResTarget.Name != "" {
				bu.WriteString(" ")
				bu.WriteString(res.ResTarget.Name)
			}
		}
	}

	// output table name
	for _, node := range stmt.SelectStmt.FromClause {
		res, err := FormatSelectStmtFromClause(ctx, node.Node)
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
		bu.WriteString("WHERE")
		bu.WriteString(" ")
		bu.WriteString(res)
	}

	return bu.String(), nil
}

func FormatSelectStmtFromClause(ctx context.Context, node any) (string, error) {
	var bu strings.Builder

	formatTableName := func(ctx context.Context, n *pg_query.Node_RangeVar) (string, error) {
		table := n.RangeVar.Relname
		if n.RangeVar.Alias != nil {
			return fmt.Sprintf("%s %s", table, n.RangeVar.Alias.Aliasname), nil
		}
		return table, nil
	}

	switch n := node.(type) {
	case *pg_query.Node_RangeVar:
		tableName, err := formatTableName(ctx, n)
		if err != nil {
			return "", err
		}
		bu.WriteString("\n")
		bu.WriteString("FROM")
		bu.WriteString(" ")
		bu.WriteString(tableName)
	case *pg_query.Node_JoinExpr:
		res, err := FormatSelectStmtFromClause(ctx, n.JoinExpr.Larg.Node)
		if err != nil {
			return "", nil
		}
		bu.WriteString(res)

		if nRangeVar, ok := n.JoinExpr.Rarg.Node.(*pg_query.Node_RangeVar); ok {
			switch n.JoinExpr.Jointype {
			case pg_query.JoinType_JOIN_INNER:
				bu.WriteString("\nINNER JOIN ")
			}
			tableName, err := formatTableName(ctx, nRangeVar)
			if err != nil {
				return "", err
			}
			bu.WriteString(tableName)
			bu.WriteString(" ON ")
		}

		if nAExpr, ok := n.JoinExpr.Quals.Node.(*pg_query.Node_AExpr); ok {
			res, err := nodeformatter.FormatAExpr(ctx, nAExpr)
			if err != nil {
				return "", err
			}
			bu.WriteString(res)
		}
	}

	return bu.String(), nil
}
