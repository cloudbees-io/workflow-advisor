package main

import (
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_GenerateJavaScript_npm(t *testing.T) {
	workflowDir, srcDir := initTest(t, []string{"package.json"})
	args := []string{
		"--workflow", path.Join(workflowDir, wfFilename),
		"--src", srcDir,
		"--generator", "js",
	}

	cmd := GenerateCommand()
	cmd.SetArgs(args)
	err := cmd.Execute()
	require.NoError(t, err)

	assertWorkflow(t, workflowDir, "js-npm.yaml")
}

func Test_GenerateJavaScript_yarn(t *testing.T) {
	workflowDir, srcDir := initTest(t, []string{"package.json", "yarn.lock"})
	args := []string{
		"--workflow", path.Join(workflowDir, wfFilename),
		"--src", srcDir,
		"--generator", "js",
	}

	cmd := GenerateCommand()
	cmd.SetArgs(args)
	err := cmd.Execute()
	require.NoError(t, err)

	assertWorkflow(t, workflowDir, "js-yarn.yaml")
}

func Test_GenerateJavaScript_NotDetected(t *testing.T) {
	workflowDir, srcDir := initTest(t, []string{})
	args := []string{
		"--workflow", path.Join(workflowDir, wfFilename),
		"--src", srcDir,
		"--generator", "js",
	}

	cmd := GenerateCommand()
	cmd.SetArgs(args)
	err := cmd.Execute()
	require.NoError(t, err)

	assertWorkflow(t, workflowDir, "undetected_smoke.yaml")
}
