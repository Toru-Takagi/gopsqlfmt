package main

import (
	"Toru-Takagi/sql_formatter_go/formatter"
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"regexp"
	"strings"
)

func main() {
	fset := token.NewFileSet()
	filePath := "testdata/sample_code_01.go"
	astFile, err := parser.ParseFile(fset, filePath, nil, 0)
	if err != nil {
		exit(err)
	}

	ast.Inspect(astFile, func(n ast.Node) bool {
		if n == nil {
			return false
		}
		switch x := n.(type) {
		//インポート、型、変数、定数などの定義。: https://zenn.dev/tenntenn/books/d168faebb1a739/viewer/becd41
		case *ast.GenDecl:
			if x.Tok == token.CONST {
				for _, spec := range x.Specs {
					vspec := spec.(*ast.ValueSpec)
					for _, v := range vspec.Values {
						re := regexp.MustCompile(`^"(.*)"$`)
						trimSQL := re.ReplaceAllString(strings.Trim(v.(*ast.BasicLit).Value, "`"), "$1")
						result, err := formatter.Format(trimSQL)
						if err != nil {
							exit(err)
						}
						v.(*ast.BasicLit).Value = "`" + result + "`"
					}
				}
			}
		}
		return true
	})

	var buf bytes.Buffer
	if err = printer.Fprint(&buf, fset, astFile); err != nil {
		exit(err)
	}
	fmt.Printf("%s", buf.String())
}

func exit(err error) {
	fmt.Printf("Error: %v\n", err)
	os.Exit(1)
}
