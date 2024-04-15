package main

import (
	"log"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/calculi-corp/workflow-advisor/pkg/utils"
)

const wfFilename = "workflow.yaml"
const testdata = "testdata"

func Test_GenerateMultiLanguage(t *testing.T) {
	workflowDir, srcDir := initTest(t, []string{"package.json"})

	err := os.WriteFile(path.Join(srcDir, "test.csproj"), []byte(`<Project Sdk="Microsoft.NET.Sdk.Web">
	<PropertyGroup>
	  <TargetFramework>net8.0</TargetFramework>
	</PropertyGroup>
  </Project>`), 0640)
	require.NoError(t, err)

	args := []string{
		"--workflow", path.Join(workflowDir, wfFilename),
		"--src", srcDir,
		"--generator", "csharp",
		"--generator", "js",
	}

	cmd := GenerateCommand()
	cmd.SetArgs(args)
	err = cmd.Execute()
	require.NoError(t, err)

	assertWorkflow(t, workflowDir, "smoke.yaml")
}

func Test_JavaGenerator(t *testing.T) {
	tests := []struct {
		name           string
		srcFiles       []string
		expectedOutput string
	}{
		{
			name:           "maven wrapper",
			srcFiles:       []string{"mvnw", "pom.xml"},
			expectedOutput: "java_mvmw.yaml",
		},
		{
			name:           "pom.xml",
			srcFiles:       []string{"pom.xml"},
			expectedOutput: "java_maven.yaml",
		},
		{
			name:           "gradle wrapper",
			srcFiles:       []string{"gradlew"},
			expectedOutput: "java_gradlew.yaml",
		},
		{
			name:           "build.gradle",
			srcFiles:       []string{"build.gradle"},
			expectedOutput: "java_gradle.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workflowDir, srcDir := initTest(t, tt.srcFiles)
			args := []string{
				"--workflow", path.Join(workflowDir, wfFilename),
				"--src", srcDir,
				"--generator", "java",
			}

			cmd := GenerateCommand()
			cmd.SetArgs(args)
			err := cmd.Execute()
			require.NoError(t, err)

			assertWorkflow(t, workflowDir, tt.expectedOutput)
		})
	}
}

func initTest(t *testing.T, testFiles []string) (workflowDir string, srcDir string) {
	t.Helper()

	workflowDir, err := os.MkdirTemp("", "workflow")
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := os.RemoveAll(workflowDir); err != nil {
			log.Fatal(err)
		}
	})

	srcDir, err = os.MkdirTemp("", "src")
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := os.RemoveAll(srcDir); err != nil {
			log.Fatal(err)
		}
	})

	for _, name := range testFiles {
		err = os.WriteFile(path.Join(srcDir, name), []byte("content"), 0640)
		require.NoError(t, err)
	}

	return workflowDir, srcDir
}

func assertWorkflow(t *testing.T, workflowDir string, expectedFilename string) {
	t.Helper()

	actualFile := path.Join(workflowDir, wfFilename)
	actual, err := utils.UnmarshalWorkflow(actualFile)
	require.NoError(t, err)

	goldenFile := path.Join(testdata, expectedFilename)
	expected, err := utils.UnmarshalWorkflow(goldenFile)
	require.NoError(t, err)
	require.Equal(t, expected, actual)

	b, err := os.ReadFile(actualFile)
	require.NoError(t, err)
	actualWfStr := string(b)

	b, err = os.ReadFile(goldenFile)
	require.NoError(t, err)
	goldenWfStr := string(b)

	require.Equal(t, goldenWfStr, actualWfStr, "formatting")
}
