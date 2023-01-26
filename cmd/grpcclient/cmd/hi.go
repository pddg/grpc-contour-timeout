package cmd

import (
	"fmt"

	"github.com/pddg/grpc-contour-timeout/proto"
	"github.com/spf13/cobra"
)

func NewHiCommand() *cobra.Command {
	var (
		delaySec int64
	)
	cmd := &cobra.Command{
		Use:   "hi MESSAGE",
		Args:  cobra.ExactArgs(1),
		Short: "Call Hi function",
		RunE: func(cmd *cobra.Command, args []string) error {
			if args[0] == "" {
				return cmd.Help()
			}
			ctx := cmd.Context()
			srv, err := cmd.Flags().GetString("server")
			if err != nil {
				return err
			}
			keepaliveEnabled, err := cmd.Flags().GetBool("keepalive")
			if err != nil {
				return err
			}
			conn, err := newGRPCClient(ctx, srv, keepaliveEnabled)
			if err != nil {
				return err
			}
			client := proto.NewGreeterClient(conn)
			resp, err := client.Hi(ctx, &proto.HiRequest{
				DelaySec: delaySec,
				Message:  args[0],
			})
			if err != nil {
				return err
			}
			fmt.Println(resp.Message)
			return nil
		},
	}
	cmd.Flags().Int64VarP(&delaySec, "delay", "d", 1, "Initial delay seconds")
	return cmd
}
