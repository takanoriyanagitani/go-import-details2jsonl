package main

import (
	"os"

	ij "github.com/takanoriyanagitani/go-import-details2jsonl"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	var cfg ij.Config = ij.
		ConfigDefault.
		WithWriter(os.Stdout)
	var ana *analysis.Analyzer = cfg.ToAnalyzer()
	singlechecker.Main(ana)
}
