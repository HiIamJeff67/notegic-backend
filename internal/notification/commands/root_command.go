package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	slogs "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/logs"
)

var rootCommand = &cobra.Command{
	Use:   "notification",
	Short: "Run the Notegic Notification runtime.",
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
	if len(os.Args) > 1 {
		slogs.NotegicLogger = slogs.NewCommandLineInterfaceLogger()
	}
	if err := rootCommand.Execute(); err != nil {
		panic(err)
	}
}

func init() {
	rootCommand.AddCommand(migrateDatabaseCommand)
}
