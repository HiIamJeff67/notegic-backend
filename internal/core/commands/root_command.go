package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	logs "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/logs"
)

var rootCommand = &cobra.Command{
	Use:   "core",
	Short: "Run the Notegic Core service or its maintenance commands.",
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

func init() {
	rootCommand.AddCommand(
		viewAllDatabaseEnumsCommand,
		migrateDatabaseCommand,
		seedDatabaseCommand,
		writeGraphQLEnumMappingValuesToConfig,
	)
}

func Execute() {
	if len(os.Args) > 1 {
		logs.NotegicLogger = logs.NewCommandLineInterfaceLogger()
	}
	if err := rootCommand.Execute(); err != nil {
		panic(err)
	}
}
