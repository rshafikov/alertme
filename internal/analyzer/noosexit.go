package analyzer

import (
	"go/ast"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
)

const doc = "noosexit checks for direct calls to os.Exit in the main function of the main package"

var Analyzer = &analysis.Analyzer{
	Name:     "noosexit",
	Doc:      doc,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func inGoBuildCache(filename string) bool {
	p := filepath.Clean(filename)
	if strings.Contains(p, string(os.PathSeparator)+"go-build"+string(os.PathSeparator)) {
		return true
	}
	if gc := strings.TrimSpace(os.Getenv("GOCACHE")); gc != "" {
		gc = filepath.Clean(gc)
		if p == gc || strings.HasPrefix(p, gc+string(os.PathSeparator)) {
			return true
		}
	}
	return false
}

func enclosingFunc(file *ast.File, start, end token.Pos) *ast.FuncDecl {
	var found *ast.FuncDecl
	ast.Inspect(file, func(n ast.Node) bool {
		if n == nil {
			return false
		}
		if fd, ok := n.(*ast.FuncDecl); ok {
			if fd.Pos() <= start && end <= fd.End() {
				found = fd
			}
		}
		return true
	})
	return found
}

func run(pass *analysis.Pass) (any, error) {
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	for _, f := range pass.Files {
		ast.Inspect(f, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			fd := enclosingFunc(f, call.Pos(), call.End())
			if fd == nil || fd.Recv != nil || fd.Name.Name != "main" || fd.Body == nil {
				return true
			}

			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || sel.Sel == nil {
				return true
			}

			if obj := pass.TypesInfo.Uses[sel.Sel]; obj == nil || obj.Pkg() == nil ||
				obj.Pkg().Path() != "os" || obj.Name() != "Exit" {
				return true
			}

			filename := pass.Fset.Position(call.Lparen).Filename
			if inGoBuildCache(filename) {
				return true
			}

			pass.Reportf(call.Lparen, "direct call to os.Exit in main function is not allowed")
			return true
		})
	}
	return nil, nil
}
