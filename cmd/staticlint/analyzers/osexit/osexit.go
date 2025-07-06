// Package osexit defines an analyzer that reports direct usage of os.Exit in main function of main package.
//
// The osexit analyzer checks for direct calls to os.Exit in the main function of the main package.
// Using os.Exit directly in main function makes testing difficult and doesn't allow
// proper cleanup of resources. Instead, main should return normally and let the
// runtime handle the exit.
//
// Example of flagged code:
//
//	package main
//
//	import "os"
//
//	func main() {
//	    if err := doSomething(); err != nil {
//	        os.Exit(1) // flagged: direct call to os.Exit in main function
//	    }
//	}
package osexit

import (
	"errors"
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// NewAnalyzer creates a new analyzer instance.
// Analyzer checks for os.Exit() calls in main package and func main().
func NewAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "osexit",
		Doc:  "Checks for os.Exit() calls in main package and func main()",
		Requires: []*analysis.Analyzer{
			inspect.Analyzer,
		},
		Run: run,
	}
}

func run(pass *analysis.Pass) (any, error) {
	if !isMainPackage(pass) {
		return nil, nil
	}

	inspector, err := getInspector(pass)
	if err != nil {
		return nil, err
	}

	checkMainFunctions(pass, inspector)
	return nil, nil
}

func isMainPackage(pass *analysis.Pass) bool {
	return pass.Pkg.Name() == "main"
}

func getInspector(pass *analysis.Pass) (*inspector.Inspector, error) {
	result, found := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !found {
		return nil, errors.New("inspect.Analyzer not found")
	}
	return result, nil
}

func checkMainFunctions(pass *analysis.Pass, insp *inspector.Inspector) {
	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
	}

	insp.Preorder(nodeFilter, func(n ast.Node) {
		funcDecl, castOK := n.(*ast.FuncDecl)
		if !castOK || !isMainFunction(funcDecl) {
			return
		}

		checkFunctionBodyForOSExit(pass, funcDecl.Body)
	})
}

func isMainFunction(funcDecl *ast.FuncDecl) bool {
	return funcDecl.Name.Name == "main"
}

func checkFunctionBodyForOSExit(pass *analysis.Pass, body *ast.BlockStmt) {
	ast.Inspect(body, func(node ast.Node) bool {
		callExpr, isCall := node.(*ast.CallExpr)
		if !isCall {
			return true
		}

		if isOSExitCall(pass, callExpr) {
			pass.Reportf(
				callExpr.Pos(),
				"direct call to os.Exit in main function of main package is prohibited",
			)
		}

		return true
	})
}

func isOSExitCall(pass *analysis.Pass, call *ast.CallExpr) bool {
	selectorExpr, isSelector := call.Fun.(*ast.SelectorExpr)
	if !isSelector || selectorExpr.Sel.Name != "Exit" {
		return false
	}

	identifier, isIdent := selectorExpr.X.(*ast.Ident)
	if !isIdent {
		return false
	}

	return isOSPackage(pass, identifier)
}

func isOSPackage(pass *analysis.Pass, ident *ast.Ident) bool {
	obj := pass.TypesInfo.Uses[ident]
	if obj == nil {
		return false
	}

	pkgName, isPkgName := obj.(*types.PkgName)
	return isPkgName && pkgName.Imported().Path() == "os"
}
