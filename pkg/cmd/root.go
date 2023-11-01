package cmd

import (
	"github.com/spf13/cobra"
	"stvz.io/coral/pkg/build"
)

type Root struct{}

func NewRoot() *Root {
	return &Root{}
}

func (r *Root) Execute() error {
	if err := r.Command().Execute(); err != nil {
		return err
	}

	return nil
}

func (r *Root) Command() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "coral",
		Short: "Build controller and image sync tool for kubernetes",
		Long: `Coral is a build controller and image sync tool for kubernetes.  It
provides components for watching source repositories for changes and building containers
when changes and conditions are detected.  It also provides a tool for syncrhonizing the
new images to nodes in a cluster based off of node labels bypassing the need for external
registries.`,
		Version: build.Version,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	rootCmd.PersistentFlags().StringP("kubeconfig", "", "", "path to kubeconfig file")
	rootCmd.AddCommand(NewController().Command())
	return rootCmd
}