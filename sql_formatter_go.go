package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"regexp"
	"strings"

	pg_query "github.com/pganalyze/pg_query_go/v4"
)

func main() {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "testdata/sample_code_01.go", nil, 0)
	if err != nil {
		exit(err)
	}

	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		//インポート、型、変数、定数などの定義。: https://zenn.dev/tenntenn/books/d168faebb1a739/viewer/becd41
		case *ast.GenDecl:
			if x.Tok == token.CONST {
				for _, spec := range x.Specs {
					vspec := spec.(*ast.ValueSpec)
					for _, v := range vspec.Values {
						re := regexp.MustCompile(`^"(.*)"$`)
						trimSQL := re.ReplaceAllString(strings.Trim(v.(*ast.BasicLit).Value, "`"), "$1")
						result, err := pg_query.Parse(trimSQL)
						if err != nil {
							exit(err)
						}
						for _, raw := range result.Stmts {
							switch internal := raw.Stmt.Node.(type) {
							case *pg_query.Node_SelectStmt:
								fmt.Println("Select Statement")
								for _, node := range internal.SelectStmt.TargetList {
									if res, ok := node.Node.(*pg_query.Node_ResTarget); ok {
										if n, ok := res.ResTarget.Val.Node.(*pg_query.Node_ColumnRef); ok {
											for _, f := range n.ColumnRef.Fields {
												if s, ok := f.Node.(*pg_query.Node_String_); ok {
													fmt.Printf("output column name: %+v\n", s.String_.Sval)
												}
											}
										}
									}
								}
								for _, node := range internal.SelectStmt.FromClause {
									if res, ok := node.Node.(*pg_query.Node_RangeVar); ok {
										fmt.Printf("table name: %+v\n", res.RangeVar.Relname)
									}
								}
							}
						}
					}
				}
			}
		}
		return true
	})
}

func exit(err error) {
	fmt.Printf("Error: %v\n", err)
	os.Exit(1)
}
