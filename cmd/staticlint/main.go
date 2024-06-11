// Package staticlint provides a custom static analysis tool that checks for
// os.Exit calls in the main function of the main package, along with a set of
// predefined analyzers from the golang.org/x/tools/go/analysis/passes package
// and the honnef.co/go/tools/staticcheck package.
package main

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/atomicalign"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/deepequalerrors"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/testinggoroutine"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"honnef.co/go/tools/staticcheck"
)

// CustomAnalyzer checks for os.Exit calls in the main function of the main package.
var CustomAnalyzer = &analysis.Analyzer{
	Name: "noosexit",                                                         // Name of the analyzer
	Doc:  "check for os.Exit calls in the main function of the main package", // Documentation for the analyzer
	Run:  run,                                                                // Function to run the analysis
}

// run is the function that performs the analysis for the CustomAnalyzer.
// It inspects the AST of the main function in the main package to check for
// calls to os.Exit and reports them.
//
// Parameters:
// - pass: The analysis.Pass object which provides the context for the analysis.
//
// Returns:
// - interface{}: The result of the analysis, which is nil in this case.
// - error: An error if any occurs during the analysis.
func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if file.Name.Name == "main" {
			for _, decl := range file.Decls {
				if fn, isFn := decl.(*ast.FuncDecl); isFn && fn.Name.Name == "main" {
					ast.Inspect(fn.Body, func(n ast.Node) bool {
						if call, isCall := n.(*ast.CallExpr); isCall {
							if fun, isFun := call.Fun.(*ast.SelectorExpr); isFun {
								if pkg, isPkg := fun.X.(*ast.Ident); isPkg && pkg.Name == "os" && fun.Sel.Name == "Exit" {
									pass.Reportf(call.Pos(), "os.Exit call in main function")
								}
							}
						}
						return true
					})
				}
			}
		}
	}
	return nil, nil
}

// main is the entry point for the static analysis tool.
// It initializes a list of analyzers, including predefined analyzers from
// the golang.org/x/tools/go/analysis/passes package, analyzers from the
// honnef.co/go/tools/staticcheck package, and a custom analyzer that checks
// for os.Exit calls in the main function of the main package.
//
// The function then runs the multichecker with the specified analyzers.
func main() {
	mychecks := []*analysis.Analyzer{
		asmdecl.Analyzer,
		assign.Analyzer,
		atomic.Analyzer,
		atomicalign.Analyzer,
		bools.Analyzer,
		buildssa.Analyzer,
		cgocall.Analyzer,
		composite.Analyzer,
		copylock.Analyzer,
		deepequalerrors.Analyzer,
		errorsas.Analyzer,
		httpresponse.Analyzer,
		loopclosure.Analyzer,
		lostcancel.Analyzer,
		nilfunc.Analyzer,
		printf.Analyzer,
		shift.Analyzer,
		stdmethods.Analyzer,
		structtag.Analyzer,
		testinggoroutine.Analyzer,
		tests.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer,
		unsafeptr.Analyzer,
		unusedresult.Analyzer,
		inspect.Analyzer,
		shadow.Analyzer,
		CustomAnalyzer,
	}

	// Add all SA analyzers from staticcheck
	for _, v := range staticcheck.Analyzers {
		if v.Analyzer.Name[:2] == "SA" {
			mychecks = append(mychecks, v.Analyzer)
		}
		// Add ST1013 analyzer from staticcheck
		if v.Analyzer.Name == "ST1013" {
			mychecks = append(mychecks, v.Analyzer)
		}
	}

	multichecker.Main(
		mychecks...,
	)
}
