package main

import (
	"github.com/Toru-Takagi/gopsqlfmt/analyzer"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() { singlechecker.Main(analyzer.FormatSQLAnalyzer) }
