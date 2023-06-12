package formatter

import (
	nodeformatter "Toru-Takagi/sql_formatter_go/formatter/node_formatter"
	"context"
	"fmt"
	"strings"

	pg_query "github.com/pganalyze/pg_query_go/v4"
)

func Format(sql string) (string, error) {
	ctx := context.Background()

	result, err := pg_query.Parse(sql)
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
	return strBuilder.String(), nil
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
				for _, f := range n.ColumnRef.Fields {
					if s, ok := f.Node.(*pg_query.Node_String_); ok {
						if ti != 0 {
							bu.WriteString(",")
						}
						bu.WriteString("\n\t")
						bu.WriteString(s.String_.Sval)
					}
				}
			}
		}
	}

	// output table name
	for _, node := range stmt.SelectStmt.FromClause {
		if res, ok := node.Node.(*pg_query.Node_RangeVar); ok {
			bu.WriteString("\n")
			bu.WriteString("FROM")
			bu.WriteString(" ")
			bu.WriteString(res.RangeVar.Relname)
		}
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
