package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Toru-Takagi/sql_formatter_go/formatter"
)

func sql_formatter_go_main() {
	path := "."
	if len(os.Args) > 1 {
		path = os.Args[1]
	}

	switch dir, err := os.Stat(path); {
	case err != nil:
		exit(err)
	case dir.IsDir():
		err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				if !info.IsDir() && !strings.HasPrefix(info.Name(), ".") && strings.HasSuffix(info.Name(), ".go") {
					formatFile(path)
				}
			}
			return nil
		})
		if err != nil {
			exit(err)
		}

	default:
		info, err := os.Stat(path)
		if err != nil {
			exit(err)
		}
		if !info.IsDir() && !strings.HasPrefix(info.Name(), ".") && strings.HasSuffix(info.Name(), ".go") {
			formatFile(path)
		}
	}
}

func formatFile(path string) error {
	isFormatted := false
	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, path, nil, parser.ParseComments) // コード上のコメントが消えてほしくないので、parser.ParseCommentsを指定
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
						if basicList, ok := v.(*ast.BasicLit); ok {
							re := regexp.MustCompile(`^"(.*)"$`)
							trimSQL := re.ReplaceAllString(strings.Trim(basicList.Value, "`"), "$1")
							trimSQL = strings.TrimSpace(trimSQL)
							upperSQL := strings.ToUpper(trimSQL)
							if strings.HasPrefix(upperSQL, "SELECT") || strings.HasPrefix(upperSQL, "INSERT") || strings.HasPrefix(upperSQL, "UPDATE") || strings.HasPrefix(upperSQL, "DELETE") {
								result, err := formatter.Format(trimSQL, nil)
								if err != nil {
									exit(err)
								}
								basicList.Value = "`" + result + "`"
								isFormatted = true
							}
						}
					}
				}
			}
		}
		return true
	})

	if isFormatted {
		var buf bytes.Buffer
		if err = printer.Fprint(&buf, fset, astFile); err != nil {
			exit(err)
		}
		if err := ioutil.WriteFile(path, []byte(buf.String()), 0); err != nil {
			exit(err)
		}
	}

	return nil
}

func exit(err error) {
	fmt.Printf("Error: %v\n", err)
	os.Exit(1)
}
