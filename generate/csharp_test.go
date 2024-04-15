package generate

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/calculi-corp/workflow-advisor/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCSharp(t *testing.T) {
	gen := &csharp{
		jobName: "cs-test",
	}

	tests := []struct {
		name         string
		src          string
		expectedPath string
		wantError    []string
		copyToTemp   bool
	}{
		{
			name:         "single-proj",
			src:          "testdata/csharp/input/single-sdk-proj",
			expectedPath: "testdata/csharp/expected/single-sdk-proj.yaml",
			copyToTemp:   true,
		},
		{
			name:         "with-solution",
			src:          "testdata/csharp/input/with-solution",
			expectedPath: "testdata/csharp/expected/with-solution.yaml",
		},
		{
			name:         "no-csharp",
			src:          "testdata/csharp/input/no-csharp",
			expectedPath: "testdata/csharp/expected/no-csharp.yaml",
		},
		{
			name:         "multiple-solutions",
			src:          "testdata/csharp/input/multiple-solutions",
			expectedPath: "testdata/csharp/expected/multiple-solutions.yaml",
		},
		{
			name:         "legacy-proj",
			src:          "testdata/csharp/input/legacy-proj",
			expectedPath: "testdata/csharp/expected/legacy-proj.yaml",
		},
		{
			name:         "multiple-versions",
			src:          "testdata/csharp/input/multiple-versions",
			expectedPath: "testdata/csharp/expected/multiple-versions.yaml",
			copyToTemp:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseDir := tt.src
			if tt.copyToTemp {
				temp, err := os.MkdirTemp(os.TempDir(), tt.name)
				require.NoError(t, err)
				defer os.RemoveAll(temp)
				err = copyDir(tt.src, temp)
				require.NoError(t, err)
				baseDir = temp
			}

			wContext := &WorkflowContext{
				SrcDir:   baseDir,
				Workflow: baseWorkflow(),
			}

			err := gen.Generate(context.Background(), wContext)
			require.NoError(t, err)

			actual := wContext.Workflow

			b, err := utils.MarshalWorkflow(actual)
			require.NoError(t, err)

			expected, err := utils.UnmarshalWorkflow(tt.expectedPath)
			require.NoError(t, err)

			if !assert.Equal(t, expected, actual) {
				fmt.Printf("\nDUMPING WORKFLOW:\n%s\n\n", string(b))
				t.FailNow()
			}
		})
	}
}

func copyDir(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	err = os.MkdirAll(dst, info.Mode())
	if err != nil {
		return err
	}

	files, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, file := range files {
		srcFile := filepath.Join(src, file.Name())
		dstFile := filepath.Join(dst, file.Name())
		if file.IsDir() {
			err = copyDir(srcFile, dstFile)
			if err != nil {
				return err
			}
		} else {
			err = copyFile(srcFile, dstFile)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	return nil
}
