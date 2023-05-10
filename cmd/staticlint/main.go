package main

import (
	"bytes"
	"go/ast"

	"go/printer"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"

	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/loggers"
)

// run is an analyzing function
func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if file.Name.Name == "main.go" {
			// функцией ast.Inspect проходим по всем узлам AST
			ast.Inspect(file, func(node ast.Node) bool {
				switch x := node.(type) {
				case *ast.CallExpr: // выражение
					buf := bytes.Buffer{}
					err := printer.Fprint(&buf, nil, x)
					if err != nil {
						loggers.ErrorLogger.Println("error while printing func name to a buffer")
						return false
					}
					funcName, err := buf.ReadString('\n')
					if err != nil {
						loggers.ErrorLogger.Println("error while reading func name from a buffer")
						return false
					}
					if funcName == "os.Exit()" {
						pass.Reportf(x.Pos(), "os.Exit() call")
					}
				}

				return true
			})
		}
	}
	return nil, nil
}

func main() {
	var mychecks []*analysis.Analyzer
	for _, v := range staticcheck.Analyzers {
		mychecks = append(mychecks, v.Analyzer)
	}
	for _, v := range quickfix.Analyzers {
		mychecks = append(mychecks, v.Analyzer)
	}
	for _, v := range simple.Analyzers {
		mychecks = append(mychecks, v.Analyzer)
	}
	for _, v := range stylecheck.Analyzers {
		mychecks = append(mychecks, v.Analyzer)
	}
	mychecks = append(mychecks, shadow.Analyzer)
	mychecks = append(mychecks, printf.Analyzer)
	osExitAnalyzer := &analysis.Analyzer{
		Name: "osExitAnalyzer",
		Doc:  "checks if file uses os.Exit() function",
		Run:  run,
	}
	mychecks = append(mychecks, osExitAnalyzer)
	multichecker.Main(
		mychecks...,
	)
}
