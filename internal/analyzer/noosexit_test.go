package analyzer_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/rshafikov/alertme/internal/analyzer"
)

func TestNoOsExit(t *testing.T) {
	testdata := analysistest.TestData()

	a := analyzer.Analyzer
	assert.Equal(t, "noosexit", a.Name)

	t.Run("detects violations", func(t *testing.T) {
		results := analysistest.Run(t, testdata, analyzer.Analyzer, "noosexit/main_with_exit")

		for _, result := range results {
			require.NoError(t, result.Err)
			require.Len(t, result.Diagnostics, 1)
			assert.Equal(t, "direct call to os.Exit in main function is not allowed", result.Diagnostics[0].Message)
		}
	})

	t.Run("no violations", func(t *testing.T) {
		results := analysistest.Run(t, testdata, analyzer.Analyzer, "noosexit/main_valid")
		for _, result := range results {
			require.NoError(t, result.Err)
			require.Len(t, result.Diagnostics, 0)
		}

		results = analysistest.Run(t, testdata, analyzer.Analyzer, "noosexit/main_with_nested_exit")
		for _, result := range results {
			require.NoError(t, result.Err)
			require.Len(t, result.Diagnostics, 0)
		}

		results = analysistest.Run(t, testdata, analyzer.Analyzer, "noosexit/non_main_package")
		for _, result := range results {
			require.NoError(t, result.Err)
			require.Len(t, result.Diagnostics, 0)
		}
	})
}

func TestNoOsExitAnalyzer(t *testing.T) {
	analyzer := analyzer.Analyzer
	assert.Equal(t, "noosexit", analyzer.Name)
	assert.Equal(t, "noosexit checks for direct calls to os.Exit in the main function of the main package", analyzer.Doc)
	assert.Len(t, analyzer.Requires, 1)
}
