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
	for _, raw := range result.Stmts {
		switch internal := raw.Stmt.Node.(type) {
		case *pg_query.Node_SelectStmt:
			fmt.Println("Select Statement")
			strBuilder.WriteString("\n")
			strBuilder.WriteString("SELECT")
			for ti, node := range internal.SelectStmt.TargetList {
				if res, ok := node.Node.(*pg_query.Node_ResTarget); ok {
					if n, ok := res.ResTarget.Val.Node.(*pg_query.Node_ColumnRef); ok {
						for _, f := range n.ColumnRef.Fields {
							if s, ok := f.Node.(*pg_query.Node_String_); ok {
								fmt.Printf("output column name: %+v\n", s.String_.Sval)
								if ti != 0 {
									strBuilder.WriteString(",")
								}
								strBuilder.WriteString("\n\t")
								strBuilder.WriteString(s.String_.Sval)
							}
						}
					}
				}
			}
			for _, node := range internal.SelectStmt.FromClause {
				if res, ok := node.Node.(*pg_query.Node_RangeVar); ok {
					strBuilder.WriteString("\n")
					strBuilder.WriteString("FROM")
					strBuilder.WriteString(" ")
					strBuilder.WriteString(res.RangeVar.Relname)
				}
			}
			if internal.SelectStmt.WhereClause != nil {
				var (
					res string
					err error
				)
				if n, ok := internal.SelectStmt.WhereClause.Node.(*pg_query.Node_AExpr); ok {
					res, err = nodeformatter.FormatAExpr(ctx, n)
				}
				if nBoolExpr, ok := internal.SelectStmt.WhereClause.Node.(*pg_query.Node_BoolExpr); ok {
					res, err = formatBoolExpr(ctx, nBoolExpr, 0)
				}
				if err != nil {
					return "", err
				}
				strBuilder.WriteString("\n")
				strBuilder.WriteString("WHERE")
				strBuilder.WriteString(" ")
				strBuilder.WriteString(res)
			}
		case *pg_query.Node_InsertStmt:
			strBuilder.WriteString("\n")
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
			strBuilder.WriteString(") VALUES (")

			// output parameter
			if internal.InsertStmt.SelectStmt != nil {
				if sNode, ok := internal.InsertStmt.SelectStmt.Node.(*pg_query.Node_SelectStmt); ok {
					for _, value := range sNode.SelectStmt.ValuesLists {
						if list, ok := value.Node.(*pg_query.Node_List); ok {
							for itemI, item := range list.List.Items {
								if pRef, ok := item.Node.(*pg_query.Node_ParamRef); ok {
									if itemI != 0 {
										strBuilder.WriteString(",")
									}
									strBuilder.WriteString("\n")
									strBuilder.WriteString("\t")
									strBuilder.WriteString("$")
									strBuilder.WriteString(fmt.Sprint(pRef.ParamRef.Number))
								}
							}
						}
					}
				}
			}

			strBuilder.WriteString("\n")
			strBuilder.WriteString(")")
		}
	}
	strBuilder.WriteString("\n")
	return strBuilder.String(), nil
}
