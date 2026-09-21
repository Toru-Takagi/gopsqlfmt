package analyzer

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestFormatSQLAnalyzerWithImportedStandardLibrary(t *testing.T) {
	dir, cleanup, err := analysistest.WriteFiles(map[string]string{
		"repro/repro.go": `package repro

import "errors"

const query = "SELECT 1"

var _ = errors.New
`,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	analysistest.Run(t, dir, FormatSQLAnalyzer, "repro")
}
