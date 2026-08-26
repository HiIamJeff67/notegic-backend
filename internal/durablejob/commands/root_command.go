package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

var rootCommand = &cobra.Command{
	Use:   "durablejob",
	Short: "Run the Notegic DurableJob service.",
	Run: func(_ *cobra.Command, _ []string) {
		ctx, stop := signal.NotifyContext(
			context.Background(),
			os.Interrupt,
			syscall.SIGTERM,
		)
		defer stop()

		<-ctx.Done()
	},
}

func Execute() {
	if err := rootCommand.Execute(); err != nil {
		panic(err)
	}
}

func init() {
	rootCommand.AddCommand(migrateDatabaseCommand)
}
