package cmd

import (
	"log"
	"time"

	"github.com/pddg/grpc-contour-timeout/proto"
	"github.com/spf13/cobra"
)

func NewSeeYouCommand() *cobra.Command {
	var (
		delaySec    int64
		intervalSec int64
		maxMessages int
	)
	cmd := &cobra.Command{
		Use:   "seeyou MESSAGE",
		Args:  cobra.ExactArgs(1),
		Short: "Call SeeYou function",
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
			server, err := client.SeeYou(ctx)
			if err != nil {
				return err
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Duration(delaySec) * time.Second):
			}
			for i := 0; i < maxMessages; i++ {
				if err := server.Send(&proto.SeeYouRequest{
					Message: args[0],
				}); err != nil {
					return err
				}
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(time.Duration(intervalSec) * time.Second):
				}
			}
			resp, err := server.CloseAndRecv()
			if err != nil {
				return err
			}
			log.Println(resp.Message)
			return nil
		},
	}
	cmd.Flags().Int64VarP(&delaySec, "delay", "d", 1, "Initial delay seconds")
	cmd.Flags().Int64VarP(&intervalSec, "interval", "i", 1, "Interval seconds")
	cmd.Flags().IntVarP(&maxMessages, "max-messages", "m", 5, "Number of messages to be sent")
	return cmd
}
