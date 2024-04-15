package main

import (
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_GeneratePython(t *testing.T) {
	type srcContents struct {
		path string
		data []byte
	}
	tests := []struct {
		name        string
		testFiles   []string
		expected    string
		srcContents []srcContents
	}{
		{
			name:      "no requirements.txt no tests",
			testFiles: []string{"hello.py"},
			expected:  "python_no_requirements_no_tests.yaml",
		},
		{
			name:      "requirements.txt no tests",
			testFiles: []string{"requirements.txt", "hello.py"},
			expected:  "python_requirements_no_tests.yaml",
		},
		{
			name:      "requirements.txt has pytest",
			testFiles: []string{"hello.py"},
			srcContents: []srcContents{
				{
					path: "requirements.txt",
					data: []byte("pytest==8.1.1"),
				},
			},
			expected: "python_requirements_pytest.yaml",
		},
		{
			name: "no requirements has pytest tests",
			srcContents: []srcContents{
				{
					path: "tests/test_hello.py",
					data: []byte("import pytest"),
				},
			},
			expected: "python_no_requirements_has_pytest_tests.yaml",
		},
		{
			name:      "has setup.py",
			testFiles: []string{"setup.py"},
			expected:  "python_setup.yaml",
		},
		{
			name:      "python undetected",
			testFiles: []string{},
			expected:  "undetected_smoke.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workflowDir, srcDir := initTest(t, tt.testFiles)
			for _, fc := range tt.srcContents {
				fp := filepath.Join(srcDir, filepath.Dir(fc.path))
				err := os.MkdirAll(fp, 0700)
				require.NoError(t, err, "error creating directory path")

				name := filepath.Base(fc.path)
				err = os.WriteFile(filepath.Join(fp, name), fc.data, 0640)
				require.NoError(t, err, "error writing file contents")
			}

			args := []string{
				"--workflow", path.Join(workflowDir, wfFilename),
				"--src", srcDir,
				"--generator", "python",
			}

			cmd := GenerateCommand()
			cmd.SetArgs(args)
			err := cmd.Execute()
			require.NoError(t, err)

			assertWorkflow(t, workflowDir, tt.expected)
		})

	}
}
