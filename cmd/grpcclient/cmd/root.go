package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "grpcclient",
		Short:         "client implementation for grpcserver",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	cmd.PersistentFlags().StringP("server", "s", "localhost:8080", "server addr")
	cmd.PersistentFlags().BoolP("keepalive", "k", false, "Enable client-side keepalive")

	cmd.AddCommand(NewLivenessCommand())
	cmd.AddCommand(NewHiCommand())
	cmd.AddCommand(NewHelloCommand())
	cmd.AddCommand(NewSeeYouCommand())
	return cmd
}

func Execute() {
	root := NewRootCommand()
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, os.Interrupt)
	defer cancel()
	if err := root.ExecuteContext(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
