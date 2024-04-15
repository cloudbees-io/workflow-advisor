package utils

import (
	"log"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/calculi-corp/dsl-engine-cli/pkg/dsl"
)

func TestMarshalUnmarshal(t *testing.T) {

	workflowDir, err := os.MkdirTemp("", "workflow")
	require.NoError(t, err)

	defer func() {
		if err := os.RemoveAll(workflowDir); err != nil {
			log.Fatal(err)
		}
	}()

	workflow := &dsl.Workflow{
		APIVersion: CurrentApiVersion,
		Kind:       WorkflowKind,
		Name:       "starter-workflow",
		On: dsl.Triggers{
			Push: &dsl.PushTrigger{
				Branches: []string{"main"},
			},
		},
		Jobs: map[string]dsl.Job{
			"test": dsl.Job{
				Steps: []dsl.Step{
					{
						Name: "step1",
						Uses: "fake/action",
						Run:  "echo hey",
						If:   "true",
					},
					{
						Name: "step2",
						Uses: "fake/action",
						Run:  "echo hey",
						If:   "true",
					},
				},
			},
		},
	}

	workflowPath := path.Join(workflowDir, "workflow.yaml")
	err = MarshalWorkflowToFile(workflowPath, workflow)
	require.NoError(t, err)

	got, err := UnmarshalWorkflow(workflowPath)
	b, _ := os.ReadFile(workflowPath)
	require.NoError(t, err, "raw content:\n"+string(b))

	require.Equal(t, workflow, got)
}
