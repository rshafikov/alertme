package analyzer

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const doc = "noosexit checks for direct calls to os.Exit in the main function of the main package"

var Analyzer = &analysis.Analyzer{
	Name:     "noosexit",
	Doc:      doc,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

// func run(pass *analysis.Pass) (interface{}, error) {
// 	ins := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

// 	if pass.Pkg.Path() == "command-line-arguments" {
// 		return nil, nil
// 	}

// 	nodeFilter := []ast.Node{(*ast.CallExpr)(nil)}

// 	ins.WithStack(nodeFilter, func(n ast.Node, push bool, stack []ast.Node) bool {

// 		if !push {
// 			return true
// 		}

// 		call := n.(*ast.CallExpr)

// 		if f := pass.Fset.File(call.Pos()); f != nil {
// 			name := f.Name()
// 			if !strings.HasSuffix(name, ".go") || strings.Contains(name, "/go-build/") {
// 				return true
// 			}
// 			if strings.HasSuffix(name, "_test.go") {
// 				return true
// 			}
// 		}

// 		if pass.Pkg.Name() != "main" {
// 			return true
// 		}
// 		var fn *ast.FuncDecl
// 		for i := len(stack) - 1; i >= 0; i-- {
// 			if d, ok := stack[i].(*ast.FuncDecl); ok {
// 				fn = d
// 				break
// 			}
// 		}
// 		if fn == nil || fn.Name.Name != "main" {
// 			return true
// 		}

// 		sel, ok := call.Fun.(*ast.SelectorExpr)
// 		if !ok || sel.Sel == nil {
// 			return true
// 		}
// 		id, ok := sel.X.(*ast.Ident)
// 		if !ok {
// 			return true
// 		}
// 		if pass.TypesInfo == nil {
// 			return true
// 		}
// 		if obj, ok := pass.TypesInfo.Uses[id]; ok {
// 			if pkgName, ok := obj.(*types.PkgName); ok {
// 				if pkgName.Imported().Path() == "os" && sel.Sel.Name == "Exit" {
// 					pass.Reportf(call.Pos(), "direct call to os.Exit in main function is not allowed")
// 				}
// 			}
// 		}

// 		return true
// 	})

// 	return nil, nil
// }

func run(pass *analysis.Pass) (interface{}, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	nodeFilter := []ast.Node{(*ast.FuncDecl)(nil), (*ast.CallExpr)(nil)}

	inMainFunc := false

	insp.WithStack(nodeFilter, func(n ast.Node, push bool, _ []ast.Node) bool {
		if !push {
			// leaving node
			if _, ok := n.(*ast.FuncDecl); ok {
				inMainFunc = false
			}
			return true
		}

		switch x := n.(type) {
		case *ast.FuncDecl:
			// we only care inside main.main of package main
			inMainFunc = pass.Pkg.Name() == "main" && x.Name.Name == "main"
		case *ast.CallExpr:
			if !inMainFunc {
				return true
			}
			if sel, ok := x.Fun.(*ast.SelectorExpr); ok {
				if pkg, ok := sel.X.(*ast.Ident); ok && pkg.Name == "os" && sel.Sel.Name == "Exit" {
					pass.Reportf(x.Pos(), "direct call to os.Exit in main function is not allowed")
				}
			}
		}
		return true
	})

	return nil, nil
}
