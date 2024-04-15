package generate

import (
	"context"
)

type pipeline struct {
	tasks []Generator
}

func (p *pipeline) Generate(ctx context.Context, workflowContext *WorkflowContext) error {
	for _, task := range p.tasks {
		err := task.Generate(ctx, workflowContext)
		if err != nil {
			return err
		}
	}

	return nil
}
