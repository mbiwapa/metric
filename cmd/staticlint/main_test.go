package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/analysis"
)

func TestRun_MultipleMainFunctions(t *testing.T) {
	tests := []struct {
		name     string
		src      string
		expected int
	}{
		{
			name: "Single main function with os.Exit",
			src: `
                package main

                import "os"

                func main() {
                    os.Exit(1)
                }
            `,
			expected: 1,
		},
		{
			name: "Single main function without os.Exit",
			src: `
                package main

                func main() {
                    println("Hello, World!")
                }
            `,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "main.go", tt.src, parser.AllErrors)
			require.NoError(t, err)

			pass := &analysis.Pass{
				Fset:  fset,
				Files: []*ast.File{file},
			}

			diagnostics := 0
			pass.Report = func(d analysis.Diagnostic) {
				diagnostics++
			}

			_, err = run(pass)
			require.NoError(t, err)
			require.Equal(t, tt.expected, diagnostics)
		})
	}
}
