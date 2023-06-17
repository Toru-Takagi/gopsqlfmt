package main

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"github.com/Toru-Takagi/sql_formatter_go/formatter"
)

func sql_formatter_go_main() {
	if len(os.Args) < 2 {
		fmt.Println("Please provide a path to a file or directory")
		return
	}

	path := os.Args[1]

	switch dir, err := os.Stat(path); {
	case err != nil:
		exit(err)
	case dir.IsDir():
		exit(errors.New("not implemented"))
	default:
		info, err := os.Stat(path)
		if err != nil {
			exit(err)
		}
		if !info.IsDir() && !strings.HasPrefix(info.Name(), ".") && strings.HasSuffix(info.Name(), ".go") {
			fset := token.NewFileSet()
			astFile, err := parser.ParseFile(fset, path, nil, 0)
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
			if err := ioutil.WriteFile(path, []byte(buf.String()), 0); err != nil {
				exit(err)
			}
		}
	}
}

func exit(err error) {
	fmt.Printf("Error: %v\n", err)
	os.Exit(1)
}
