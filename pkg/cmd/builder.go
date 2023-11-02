package cmd

import "github.com/spf13/cobra"

const (
	BuilderUsage     = "build [ARG...]"
	BuilderShortDesc = "Start the image builder service"
	BuilderLongDesc  = `Starts the image builder service which will trigger builds as build events
are received.`
)

type Builder struct{}

func NewBuilder() *Builder {
	return &Builder{}
}

func (b *Builder) RunE(cmd *cobra.Command, args []string) error {
	return nil
}

func (b *Builder) Command() *cobra.Command {
	builderCmd := &cobra.Command{
		Use:   BuilderUsage,
		Short: BuilderShortDesc,
		Long:  BuilderLongDesc,
		RunE:  b.RunE,
	}

	return builderCmd
}
