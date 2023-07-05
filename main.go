package main

import (
	"github.com/Toru-Takagi/sql_formatter_go/analyzer"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() { singlechecker.Main(analyzer.FormatSQLAnalyzer) }
