package generate

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type fakeGen struct {
	count int
}

func (g *fakeGen) Generate(ctx context.Context, workflowContext *WorkflowContext) error {
	g.count += 1
	return nil
}

func TestPipeline(t *testing.T) {
	gen := &fakeGen{
		count: 0,
	}

	pip := pipeline{
		tasks: []Generator{gen, gen},
	}

	err := pip.Generate(context.Background(), &WorkflowContext{})
	require.NoError(t, err)

	require.Equal(t, 2, gen.count)
}
