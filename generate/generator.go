package generate

import (
	"context"
	"fmt"

	"github.com/calculi-corp/dsl-engine-cli/pkg/dsl"
	"github.com/calculi-corp/workflow-advisor/pkg/utils"
)

var registeredGenerators = map[string][]Generator{}

type WorkflowContext struct {
	Workflow *dsl.Workflow
	SrcDir   string
}

type Generator interface {
	Generate(ctx context.Context, workflowContext *WorkflowContext) error
}

func Generate(ctx context.Context, workflowPath string, src string, generatorNames []string) error {
	exists, err := utils.Stat(workflowPath)

	if err != nil {
		return err
	}

	if !exists {
		workflow := baseWorkflow()
		err := utils.MarshalWorkflowToFile(workflowPath, workflow)
		if err != nil {
			return err
		}
	}

	workflow, err := utils.UnmarshalWorkflow(workflowPath)

	if err != nil {
		return err
	}

	genPipeline, err := buildPipeline(generatorNames)
	if err != nil {
		return err
	}

	wContext := &WorkflowContext{
		Workflow: workflow,
		SrcDir:   src,
	}

	err = genPipeline.Generate(ctx, wContext)
	if err != nil {
		return err
	}

	err = utils.MarshalWorkflowToFile(workflowPath, wContext.Workflow)
	return err
}

func buildPipeline(generatorNames []string) (Generator, error) {
	generators := []Generator{}

	for _, name := range generatorNames {
		gen, ok := registeredGenerators[name]

		if !ok {
			return nil, fmt.Errorf("can not find generators with name '%s'", name)
		}

		generators = append(generators, gen...)
	}

	return &pipeline{tasks: generators}, nil
}

func baseWorkflow() *dsl.Workflow {
	return &dsl.Workflow{
		APIVersion: utils.CurrentApiVersion,
		Kind:       utils.WorkflowKind,
		Name:       "build",
		On: dsl.Triggers{
			Push: &dsl.PushTrigger{
				Branches: []string{"**"},
			},
		},
	}
}

func registerGenerator(name string, gen Generator) {
	generators, ok := registeredGenerators[name]

	if !ok {
		generators = []Generator{}
	}

	generators = append(generators, gen)
	registeredGenerators[name] = generators
}
