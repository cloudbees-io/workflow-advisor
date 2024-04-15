package generate

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/calculi-corp/dsl-engine-cli/pkg/dsl"
	"github.com/calculi-corp/workflow-advisor/pkg/utils"
)

const nodeImage = "docker://node:21-alpine3.19"

type javascript struct {
	jobName string
}

func init() {
	registerGenerator("js", &javascript{
		jobName: "js-build",
	})
}

func (g *javascript) Generate(ctx context.Context, workflowContext *WorkflowContext) error {
	srcDir := workflowContext.SrcDir
	exists, err := utils.Stat(filepath.Join(srcDir, "package.json"))
	if err != nil || !exists {
		return err
	}

	depsStep := dsl.Step{
		Name: "get dependencies",
		Uses: nodeImage,
		Run:  "npm install",
	}
	buildStep := dsl.Step{
		Name: "build",
		Uses: nodeImage,
		Run:  "npm run build",
	}
	testStep := dsl.Step{
		Name: "test",
		Uses: nodeImage,
		Run:  "npm run test",
	}

	yarnLockExists, err := utils.Stat(filepath.Join(srcDir, "yarn.lock"))
	if err != nil {
		return err
	}
	if yarnLockExists {
		depsStep.Run = "yarn install"
		buildStep.Run = "yarn run build"
		testStep.Run = "yarn run test"
	}

	return g.addJob(workflowContext.Workflow, depsStep, buildStep, testStep)
}

func (g *javascript) addJob(wf *dsl.Workflow, steps ...dsl.Step) error {
	if wf.Jobs == nil {
		wf.Jobs = make(map[string]dsl.Job)
	}

	if _, ok := wf.Jobs[g.jobName]; ok {
		return fmt.Errorf("error adding job: job %s already exists", g.jobName)
	}

	wf.Jobs[g.jobName] = dsl.Job{
		Steps: append(append([]dsl.Step{
			{
				Name: "checkout",
				Uses: "cloudbees-io/checkout@v1",
			},
		}, steps...),
			dsl.Step{
				Name: "scan",
				Uses: "cloudbees-io/sonarqube-bundled-sast-scan-code@v2",
				With: map[string]string{
					"language": "LANGUAGE_JS",
				},
			},
		),
	}

	return nil
}
