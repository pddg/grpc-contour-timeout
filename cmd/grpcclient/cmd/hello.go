package cmd

import (
	"log"

	"github.com/pddg/grpc-contour-timeout/proto"
	"github.com/spf13/cobra"
)

func NewHelloCommand() *cobra.Command {
	var (
		delaySec    int64
		intervalSec int64
	)
	cmd := &cobra.Command{
		Use:   "hello MESSAGE",
		Args:  cobra.ExactArgs(1),
		Short: "Call Hello function",
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
			resp, err := client.Hello(ctx, &proto.HelloRequest{
				DelaySec:    delaySec,
				IntervalSec: intervalSec,
				Message:     args[0],
			})
			if err != nil {
				return err
			}
			for {
				msg, err := resp.Recv()
				if err != nil {
					return err
				}
				log.Println(msg.Message)
			}
		},
	}
	cmd.Flags().Int64VarP(&delaySec, "delay", "d", 1, "Initial delay seconds")
	cmd.Flags().Int64VarP(&intervalSec, "interval", "i", 1, "Interval seconds")
	return cmd
}
