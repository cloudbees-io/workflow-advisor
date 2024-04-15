package main

import (
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_GenerateGolang(t *testing.T) {
	workflowDir, srcDir := initTest(t, []string{"go.mod", "go.sum"})
	args := []string{
		"--workflow", path.Join(workflowDir, wfFilename),
		"--src", srcDir,
		"--generator", "go",
	}

	cmd := GenerateCommand()
	cmd.SetArgs(args)
	err := cmd.Execute()
	require.NoError(t, err)

	assertWorkflow(t, workflowDir, "golang_smoke.yaml")
}

func Test_GenerateGolang_NotDetected(t *testing.T) {
	// Golang detector requires go.mod and go.sum, should not generate golang job
	workflowDir, srcDir := initTest(t, []string{"go.mod"})
	args := []string{
		"--workflow", path.Join(workflowDir, wfFilename),
		"--src", srcDir,
		"--generator", "go",
	}

	cmd := GenerateCommand()
	cmd.SetArgs(args)
	err := cmd.Execute()
	require.NoError(t, err)

	assertWorkflow(t, workflowDir, "undetected_smoke.yaml")
}
