package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func NewLivenessCommand() *cobra.Command {
	var (
		timeout time.Duration
	)
	cmd := &cobra.Command{
		Use:   "liveness",
		Short: "Check liveness of the server",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			srv, err := cmd.Flags().GetString("server")
			if err != nil {
				return err
			}
			keepaliveEnabled, err := cmd.Flags().GetBool("keepalive")
			if err != nil {
				return err
			}
			ctx, cancel := context.WithTimeout(cmd.Context(), timeout)
			defer cancel()
			conn, err := newGRPCClient(ctx, srv, keepaliveEnabled)
			if err != nil {
				return err
			}
			client := grpc_health_v1.NewHealthClient(conn)
			if _, err := client.Check(ctx, &grpc_health_v1.HealthCheckRequest{}); err != nil {
				return err
			}
			fmt.Println("ok")
			return nil
		},
	}
	cmd.Flags().DurationVarP(&timeout, "timeout", "t", 1*time.Second, "Request timeout")
	return cmd
}
