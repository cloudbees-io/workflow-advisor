package generate

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/calculi-corp/dsl-engine-cli/pkg/dsl"
	"github.com/calculi-corp/workflow-advisor/pkg/utils"
)

type golang struct {
	jobName string
}

func init() {
	registerGenerator("go", &golang{
		jobName: "go-build",
	})
}

func (g *golang) Generate(ctx context.Context, workflowContext *WorkflowContext) error {
	isGolang, err := g.containsSource(workflowContext.SrcDir)
	if err != nil {
		return err
	}
	if !isGolang {
		return nil
	}

	workflow := workflowContext.Workflow
	return g.addJobIfNotExists(workflow)
}

func (g *golang) containsSource(srcDir string) (bool, error) {
	for _, f := range []string{"go.mod", "go.sum"} {
		exists, err := utils.Stat(filepath.Join(srcDir, f))
		if err != nil {
			return false, err
		}
		if !exists {
			return false, nil
		}
	}
	return true, nil
}

func (g *golang) addJobIfNotExists(wf *dsl.Workflow) error {
	if wf.Jobs == nil {
		wf.Jobs = make(map[string]dsl.Job)
	}

	if _, ok := wf.Jobs[g.jobName]; ok {
		return fmt.Errorf("error adding job: job %s already exists", g.jobName)
	}

	wf.Jobs[g.jobName] = dsl.Job{
		Steps: []dsl.Step{
			{
				Name: "checkout",
				Uses: "cloudbees-io/checkout@v1",
			},
			{
				Name: "test",
				Uses: "docker://golang:1.22-alpine3.19",
				Run:  "go test -cover ./...",
			},
			{
				Name: "build",
				Uses: "docker://golang:1.22-alpine3.19",
				Run:  "go build ./...",
			},
			{
				Name: "scan",
				Uses: "cloudbees-io/sonarqube-bundled-sast-scan-code@v2",
				With: map[string]string{
					"language": "LANGUAGE_GO",
				},
			},
		},
	}

	return nil
}
