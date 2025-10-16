// Package main provides a multichecker tool that combines multiple static analysis checks
// for Go code quality and correctness.
//
// The multichecker includes:
//   - All standard analysis passes from golang.org/x/tools/go/analysis/passes
//   - Selected staticcheck analyzers (SA, ST, QF, and S1 prefixes)
//   - A custom analyzer to detect direct calls to os.Exit in main functions
//
// Usage:
//
//	go run cmd/staticlint/main.go [packages...]
//
// The multichecker will analyze the specified packages using all configured analyzers.
// If no packages are specified, it will analyze the current package.
package main

import (
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/atomicalign"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/ctrlflow"
	"golang.org/x/tools/go/analysis/passes/deepequalerrors"
	"golang.org/x/tools/go/analysis/passes/defers"
	"golang.org/x/tools/go/analysis/passes/directive"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/fieldalignment"
	"golang.org/x/tools/go/analysis/passes/findcall"
	"golang.org/x/tools/go/analysis/passes/framepointer"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/ifaceassert"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/nilness"
	"golang.org/x/tools/go/analysis/passes/pkgfact"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/reflectvaluecompare"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/sigchanyzer"
	"golang.org/x/tools/go/analysis/passes/sortslice"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/testinggoroutine"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/timeformat"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"golang.org/x/tools/go/analysis/passes/unusedwrite"

	"honnef.co/go/tools/staticcheck"

	"github.com/rshafikov/alertme/internal/analyzer"
)

// main configures and runs the multichecker with all selected analyzers.
//
// The analyzers are grouped into several categories:
//  1. Standard Go analysis passes from golang.org/x/tools/go/analysis/passes
//  2. Staticcheck analyzers (filtered by prefix)
//  3. Custom analyzers
//
// The multichecker is then executed with multichecker.Main(checks...).
func main() {
	var checks []*analysis.Analyzer

	// Standard Go analysis passes
	// These are the standard analyzers provided by golang.org/x/tools/go/analysis/passes
	// Each analyzer checks for specific issues in Go code:
	// - asmdecl: reports mismatches between assembly files and Go function declarations
	// - assign: detects useless assignments
	// - atomic: checks for common mistakes using the sync/atomic package
	// - atomicalign: checks for non-64-bit-aligned arguments to sync/atomic functions
	// - bools: detects common boolean expressions mistakes
	// - buildssa: builds SSA-form IR for later passes
	// - buildtag: checks //go:build directives
	// - cgocall: detects some violations of the cgo pointer passing rules
	// - composite: suggests use of struct literals to enforce field names
	// - copylock: detects locks erroneously passed by value
	// - ctrlflow: provides a control-flow graph
	// - deepequalerrors: checks for the use of reflect.DeepEqual with error values
	// - defers: reports common mistakes in defer statements
	// - directive: checks known Go toolchain directives
	// - errorsas: reports improper use of errors.As
	// - fieldalignment: finds structs that would use less memory if their fields were sorted
	// - findcall: serves as a template for writing a checker
	// - framepointer: reports assembly code that clobbers the frame pointer before saving it
	// - httpresponse: checks for mistakes using HTTP responses
	// - ifaceassert: detects impossible interface-interface type assertions
	// - loopclosure: checks for references to loop variables from within nested functions
	// - lostcancel: checks for contexts that are canceled but not propagated
	// - nilfunc: checks for useless comparisons against nil
	// - nilness: checks for redundant or impossible nil comparisons
	// - pkgfact: gathers information about package facts
	// - printf: checks consistency of Printf format strings and arguments
	// - reflectvaluecompare: reports comparisons between reflect.Value values
	// - shadow: reports variables that may have been shadowed
	// - shift: reports shifts that exceed the width of an integer
	// - sigchanyzer: detects misuse of unbuffered os.Signal channels
	// - sortslice: reports calls to sort.Slice that do not use a less function that is slice-dependent
	// - stdmethods: checks for malformed standard method signatures
	// - stringintconv: flags type conversions from int to string
	// - structtag: checks struct field tags for well-formedness
	// - testinggoroutine: reports calls to (*testing.T).Fatal from goroutines started by a test
	// - tests: checks for common mistaken usages of tests and examples
	// - timeformat: checks for calls of (time.Time).Format or time.Parse with 2006-01-02 format
	// - unmarshal: reports suspicious arguments to unmarshal functions
	// - unreachable: checks for unreachable code
	// - unsafeptr: reports likely incorrect uses of unsafe.Pointer
	// - unusedresult: checks for unused results of calls to functions in an allowlist
	// - unusedwrite: checks for unused writes to the elements of a struct or array object
	checks = append(checks,
		asmdecl.Analyzer,
		assign.Analyzer,
		atomic.Analyzer,
		atomicalign.Analyzer,
		bools.Analyzer,
		buildssa.Analyzer,
		buildtag.Analyzer,
		cgocall.Analyzer,
		composite.Analyzer,
		copylock.Analyzer,
		ctrlflow.Analyzer,
		deepequalerrors.Analyzer,
		defers.Analyzer,
		directive.Analyzer,
		errorsas.Analyzer,
		fieldalignment.Analyzer,
		findcall.Analyzer,
		framepointer.Analyzer,
		httpresponse.Analyzer,
		ifaceassert.Analyzer,
		loopclosure.Analyzer,
		lostcancel.Analyzer,
		nilfunc.Analyzer,
		nilness.Analyzer,
		pkgfact.Analyzer,
		printf.Analyzer,
		reflectvaluecompare.Analyzer,
		shadow.Analyzer,
		shift.Analyzer,
		sigchanyzer.Analyzer,
		sortslice.Analyzer,
		stdmethods.Analyzer,
		stringintconv.Analyzer,
		structtag.Analyzer,
		testinggoroutine.Analyzer,
		tests.Analyzer,
		timeformat.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer,
		unsafeptr.Analyzer,
		unusedresult.Analyzer,
		unusedwrite.Analyzer,
	)

	// Staticcheck analyzers with SA prefix (Static Analysis)
	// These analyzers detect various code issues and potential bugs:
	// - SA1xxx: Various analysis checks
	// - SA2xxx: Time-related checks
	// - SA3xxx: Testing-related checks
	// - SA4xxx: Performance-related checks
	// - SA5xxx: Error handling checks
	// - SA6xxx: Crypto-related checks
	// - SA9xxx: Stdlib usage checks
	for _, v := range staticcheck.Analyzers {
		if len(v.Analyzer.Name) > 2 && v.Analyzer.Name[0:2] == "SA" {
			checks = append(checks, v.Analyzer)
		}
	}

	// Staticcheck analyzers with ST prefix (Style)
	// These analyzers detect style issues and code improvements:
	// - ST1xxx: Various style checks
	for _, v := range staticcheck.Analyzers {
		if len(v.Analyzer.Name) > 2 && v.Analyzer.Name[0:2] == "ST" {
			checks = append(checks, v.Analyzer)
			break
		}
	}

	// Additional staticcheck analyzers with QF, S1 prefixes
	// QF analyzers focus on code simplification
	// S1 analyzers detect various issues and improvements
	for _, v := range staticcheck.Analyzers {
		if len(v.Analyzer.Name) > 2 && v.Analyzer.Name[0:2] == "QF" {
			checks = append(checks, v.Analyzer)
		}
		if len(v.Analyzer.Name) > 2 && v.Analyzer.Name[0:2] == "S1" {
			checks = append(checks, v.Analyzer)
		}
		if len(checks) > 120 {
			break
		}
	}

	// Custom analyzers
	// These are project-specific analyzers developed for this project:
	// - noosexit: checks for direct calls to os.Exit in the main function of the main package
	checks = append(checks, analyzer.Analyzer)

	// Execute the multichecker with all configured analyzers
	// The multichecker will run all analyzers against the specified packages
	// and report any issues found.
	multichecker.Main(checks...)
}
