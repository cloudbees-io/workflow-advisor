package main

import (
	"context"
	"os"

	"github.com/calculi-corp/workflow-advisor/generate"
	"github.com/spf13/cobra"
)

func GenerateCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   os.Args[0],
		Short: "Detect technologies and generate workflow",
		Long:  `Detect technologies and generate workflow`,
		RunE: func(cmd *cobra.Command, args []string) error {
			generators, _ := cmd.Flags().GetStringSlice("generator")
			workflow, _ := cmd.Flags().GetString("workflow")
			src, _ := cmd.Flags().GetString("src")

			return generate.Generate(context.Background(), workflow, src, generators)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.SetFlagErrorFunc(handleError)

	cmd.Flags().StringSliceP("generator", "g", []string{}, "Generators to run")
	cmd.MarkFlagRequired("generator")

	cmd.Flags().StringP("workflow", "w", "", "Workflow file to use, created if not exists")
	cmd.MarkFlagRequired("workflow")

	cmd.Flags().String("src", "", "Workflow file to use, created if not exists")
	cmd.MarkFlagRequired("workflow")

	return &cmd
}

func handleError(cmd *cobra.Command, err error) error {
	cmd.Help()
	return err
}
