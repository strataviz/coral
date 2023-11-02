package cmd

import "github.com/spf13/cobra"

const (
	WatchUsage     = "watch [ARG...]"
	WatchShortDesc = "Start the repository watcher"
	WatchLongDesc  = `Starts the repository watcher which will generate build events when a repository
changes based on a set of defined conditions.`
)

type Watch struct{}

func NewWatch() *Watch {
	return &Watch{}
}

func (w *Watch) RunE(cmd *cobra.Command, args []string) error {
	return nil
}

func (w *Watch) Command() *cobra.Command {
	watchCmd := &cobra.Command{
		Use:   WatchUsage,
		Short: WatchShortDesc,
		Long:  WatchLongDesc,
		RunE:  w.RunE,
	}

	return watchCmd
}
