package analyzer

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"io/ioutil"
	"strings"

	"github.com/Toru-Takagi/gopsqlfmt/formatter"
	"golang.org/x/tools/go/analysis"
)

var FormatSQLAnalyzer = &analysis.Analyzer{
	Name: "format_sql",
	Doc:  "Only target those of sql string",
	Run:  formatSQLRun,
}

func formatSQLRun(pass *analysis.Pass) (interface{}, error) {
	for _, astFile := range pass.Files {
		var formatErr error
		isFormatted := false

		fname := pass.Fset.Position(astFile.Package).Filename

		if strings.HasSuffix(fname, "_gen.go") {
			// Skip generated files
			return nil, nil
		}

		ast.Inspect(astFile, func(n ast.Node) bool {
			if n == nil {
				return false
			}

			switch x := n.(type) {
			case *ast.GenDecl:
				if x.Tok == token.CONST {
					for _, spec := range x.Specs {
						vspec := spec.(*ast.ValueSpec)
						for _, v := range vspec.Values {
							if basicList, ok := v.(*ast.BasicLit); ok {
								trimSQL := strings.TrimSpace(strings.NewReplacer([]string{
									"`", "",
									`"`, "",
								}...).Replace(basicList.Value))
								upperSQL := strings.ToUpper(trimSQL)
								if strings.HasPrefix(upperSQL, "SELECT") || strings.HasPrefix(upperSQL, "INSERT") || strings.HasPrefix(upperSQL, "UPDATE") || strings.HasPrefix(upperSQL, "DELETE") {
									result, err := formatter.Format(trimSQL, nil)
									if err != nil {
										formatErr = err
										return false
									}
									basicList.Value = fmt.Sprintf("`%s`", result)
									isFormatted = true
								}
							}
						}
					}
				}
			}
			return true
		})

		if formatErr != nil {
			return nil, formatErr
		}

		if isFormatted {
			var buf bytes.Buffer
			if err := format.Node(&buf, pass.Fset, astFile); err != nil {
				return nil, err
			}

			// Overwrite the original file with the new code
			filename := pass.Fset.File(astFile.Pos()).Name()
			if err := ioutil.WriteFile(filename, buf.Bytes(), 0644); err != nil {
				return nil, err
			}
		}

	}
	return nil, nil
}
