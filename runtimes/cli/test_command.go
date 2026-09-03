package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"

	"github.com/spf13/cobra"
)

var testModules = []string{
	"contracts",
	"shared",
	"runtimes/cli",
	"runtimes/core",
	"runtimes/clientgateway",
	"runtimes/apigateway",
	"runtimes/durablejob",
	"runtimes/email",
	"runtimes/notification",
	"runtimes/realtimegateway",
	"test",
}

func init() {
	rootCommand.AddCommand(
		newTestAllCommand(),
		newTestModuleCommand(),
		newTestRaceCommand(),
		newTestArchitectureCommand(),
	)
}

func newTestAllCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "test-all",
		Short: "Run all Go module tests.",
		RunE: func(command *cobra.Command, _ []string) error {
			root, err := findRepositoryRoot()
			if err != nil {
				return err
			}

			for _, module := range testModules {
				if err := runGo(command.Context(), root, module, "test", "./..."); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

func newTestModuleCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "test-module <module>",
		Short: "Run tests for one workspace module.",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			if !slices.Contains(testModules, args[0]) {
				return fmt.Errorf("unsupported module %q", args[0])
			}

			root, err := findRepositoryRoot()
			if err != nil {
				return err
			}

			return runGo(command.Context(), root, args[0], "test", "./...")
		},
	}
}

func newTestRaceCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "test-race <module>",
		Short: "Run the race detector for one workspace module.",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			if !slices.Contains(testModules, args[0]) {
				return fmt.Errorf("unsupported module %q", args[0])
			}

			root, err := findRepositoryRoot()
			if err != nil {
				return err
			}

			return runGo(command.Context(), root, args[0], "test", "-race", "./...")
		},
	}
}

func newTestArchitectureCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "test-architecture",
		Short: "Run root architecture tests.",
		RunE: func(command *cobra.Command, _ []string) error {
			root, err := findRepositoryRoot()
			if err != nil {
				return err
			}

			return runGo(command.Context(), root, "test", "test", "./architecture")
		},
	}
}

func findRepositoryRoot() (string, error) {
	if configuredRoot := os.Getenv("NOTEGIC_REPOSITORY_ROOT"); configuredRoot != "" {
		return filepath.Abs(configuredRoot)
	}

	directory, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(directory, "go.work")); err == nil {
			return directory, nil
		}

		parent := filepath.Dir(directory)
		if parent == directory {
			return "", fmt.Errorf("could not locate repository go.work")
		}
		directory = parent
	}
}

func runGo(ctx context.Context, root, module string, args ...string) error {
	command := exec.CommandContext(ctx, "go", args...)
	command.Dir = filepath.Join(root, module)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	if err := command.Run(); err != nil {
		return fmt.Errorf("go %s in %s: %w", args[0], module, err)
	}

	return nil
}
