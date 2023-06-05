package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
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
						fmt.Println(strings.Trim(v.(*ast.BasicLit).Value, "`"))
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
