/*
Package main implements a multichecker that combines multiple static analysis tools
for comprehensive code quality checking.

# Usage

To run the staticlint tool on your Go code:

	go run cmd/staticlint/main.go ./...

Or build and run:

	go build -o staticlint cmd/staticlint/main.go
	./staticlint ./...

# Included Analyzers

This multichecker includes the following categories of analyzers:

# Standard Go Analysis Passes

All standard analyzers from golang.org/x/tools/go/analysis/passes:
  - appends: checks for missing values after append
  - asmdecl: reports mismatches between assembly files and Go declarations
  - assign: detects useless assignments
  - atomic: checks for common mistakes using the sync/atomic package
  - atomicalign: checks for non-64-bit-aligned arguments to sync/atomic functions
  - bools: detects common mistakes involving boolean operators
  - buildssa: constructs the SSA representation of Go programs
  - buildtag: checks build constraints
  - cgocall: detects some violations of the cgo pointer passing rules
  - composite: checks for unkeyed composite literals
  - copylock: checks for locks erroneously passed by value
  - ctrlflow: provides a syntactic control-flow graph for the body of a function
  - deepequalerrors: checks for the use of reflect.DeepEqual with error values
  - defers: checks for common mistakes in defer statements
  - directive: checks known Go toolchain directives
  - errorsas: checks that the second argument to errors.As is a pointer to a type implementing error
  - fieldalignment: detects structs that would use less memory if their fields were sorted
  - gofix: applies fixes suggested by the go fix command
  - hostport: checks for common mistakes using net.JoinHostPort
  - httpmux: reports using Go 1.22's enhanced ServeMux patterns in older Go versions
  - httpresponse: checks for mistakes using HTTP responses
  - ifaceassert: flags impossible interface-interface type assertions
  - inspect: provides an AST inspector for the syntax trees of a package
  - loopclosure: checks for references to loop variables from within nested functions
  - lostcancel: checks for failure to call a context cancellation function
  - nilfunc: checks for useless comparisons between functions and nil
  - nilness: inspects the control-flow graph of an SSA function and reports errors such as nil pointer dereferences
  - pkgfact: demonstrates how to write and read analysis facts
  - printf: checks consistency of Printf format strings and arguments
  - reflectvaluecompare: checks for comparing reflect.Value values with == or reflect.DeepEqual
  - shadow: checks for shadowed variables
  - shift: checks for shifts that equal or exceed the width of the integer
  - sigchanyzer: detects misuse of unbuffered signal as argument to signal.Notify
  - slog: checks for invalid structured logging calls
  - sortslice: checks for calls to sort.Slice that do not use a slice type as first argument
  - stdmethods: checks for misspellings of well-known method signatures
  - stdversion: reports uses of standard library symbols that are too new for the Go version in use
  - stringintconv: flags type conversions from integers to strings
  - structtag: checks struct field tags
  - testinggoroutine: detects calls to Fatal from a test goroutine
  - tests: checks for common mistakes in tests
  - timeformat: checks for the use of time.Format or time.Parse calls with a bad format
  - unmarshal: checks for passing non-pointer or non-interface values to unmarshal
  - unreachable: checks for unreachable code
  - unsafeptr: checks for invalid conversions of uintptr to unsafe.Pointer
  - unusedresult: checks for unused results of calls to certain functions
  - unusedwrite: checks for unused writes to the elements of a struct or array object
  - usesgenerics: checks for usage of generic features added in Go 1.18
  - waitgroup: detects simple misuses of sync.WaitGroup

# Staticcheck SA Class Analyzers

All SA (Static Analysis) class analyzers from https://staticcheck.dev/docs/checks/ that detect
bugs and performance issues:
  - SA1xxx: Various incorrectness checks
  - SA2xxx: Concurrency issues
  - SA3xxx: Testing issues
  - SA4xxx: Code that isn't really doing anything
  - SA5xxx: Correctness issues
  - SA6xxx: Performance issues
  - SA9xxx: Questionable constructs

# Additional Staticcheck Analyzers

Selected analyzers from https://staticcheck.dev/docs/checks/ that detect
style issues:
  - ST1003 (stylecheck): Poorly named identifier
  - QF1003 (quickfix): Convert if/else-if chain to tagged switch
  - S1000 (simple): Use plain channel send or receive instead of single-case select

# Third-party Analyzers

  - exhaustive: Checks exhaustiveness of enum switch statements
  - bodyclose: Static analysis tool which checks whether res.Body is correctly closed.

# Custom Analyzers

  - osexit: Prohibits direct calls to os.Exit in main function of main package

# Custom Analyzer: osexit

The osexit analyzer prevents direct calls to os.Exit in the main function
of the main package. This practice makes testing difficult and prevents
proper resource cleanup. Instead of calling os.Exit directly in main,
consider extracting the main logic into a separate function that returns
an error, and only call os.Exit at the very end of main if needed.

Bad:

	func main() {
	    if err := doSomething(); err != nil {
	        log.Fatal(err)
	        os.Exit(1) // This will be flagged
	    }
	}

# Exit Codes

The tool exits with status 0 if no issues are found, and non-zero status
if any issues are detected or if there are errors during analysis.
*/
package main

import (
	"github.com/nishanths/exhaustive"
	"github.com/timakin/bodyclose/passes/bodyclose"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/appends"
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
	"golang.org/x/tools/go/analysis/passes/gofix"
	"golang.org/x/tools/go/analysis/passes/hostport"
	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"

	"github.com/AGENT3128/shortener-url/cmd/staticlint/analyzers/osexit"

	"golang.org/x/tools/go/analysis/passes/httpmux"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/ifaceassert"
	"golang.org/x/tools/go/analysis/passes/inspect"
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
	"golang.org/x/tools/go/analysis/passes/slog"
	"golang.org/x/tools/go/analysis/passes/sortslice"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stdversion"
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
	"golang.org/x/tools/go/analysis/passes/usesgenerics"
	"golang.org/x/tools/go/analysis/passes/waitgroup"
)

func main() {
	var analyzers []*analysis.Analyzer
	analyzers = append(analyzers,
		appends.Analyzer,
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
		gofix.Analyzer,
		hostport.Analyzer,
		httpmux.Analyzer,
		httpresponse.Analyzer,
		ifaceassert.Analyzer,
		inspect.Analyzer,
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
		slog.Analyzer,
		sortslice.Analyzer,
		stdmethods.Analyzer,
		stdversion.Analyzer,
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
		usesgenerics.Analyzer,
		waitgroup.Analyzer,
	)

	// add all 'SA' class analyzers from staticcheck.io
	for _, analyzer := range staticcheck.Analyzers {
		analyzers = append(analyzers, analyzer.Analyzer)
	}

	// add 'S' class analyzers from staticcheck.io
	for _, analyzer := range simple.Analyzers {
		// S1000	Use plain channel send or receive instead of single-case select
		if analyzer.Analyzer.Name == "S1000" {
			analyzers = append(analyzers, analyzer.Analyzer)
		}
	}

	// add 'ST' class analyzers from staticcheck.io
	for _, analyzer := range stylecheck.Analyzers {
		// ST1003	Poorly chosen identifier
		if analyzer.Analyzer.Name == "ST1003" {
			analyzers = append(analyzers, analyzer.Analyzer)
		}
	}

	// add 'QF' class analyzers from staticcheck.io
	for _, analyzer := range quickfix.Analyzers {
		// QF1003	Convert if/else-if chain to tagged switch
		if analyzer.Analyzer.Name == "QF1003" {
			analyzers = append(analyzers, analyzer.Analyzer)
		}
	}
	// Add third-party analyzers
	analyzers = append(analyzers,
		exhaustive.Analyzer,
		bodyclose.Analyzer,
	)

	// Add custom analyzer - osexit: checks for os.Exit() calls in main package and func main()
	analyzers = append(analyzers, osexit.NewAnalyzer())

	multichecker.Main(analyzers...)
}
