package osexit_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/AGENT3128/shortener-url/cmd/staticlint/analyzers/osexit"
)

func TestOsExitAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, osexit.NewAnalyzer(), "./...")
}
