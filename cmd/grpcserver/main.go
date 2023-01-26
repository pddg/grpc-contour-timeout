package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"

	"github.com/pddg/grpc-contour-timeout/proto"
)

var (
	listenPort               int
	serverKeepaliveInterval  time.Duration
	enforceKeepaliveInterval time.Duration
	enableServerKeepalive    bool
	enforceKeepalive         bool
)

func innerMain() error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, os.Interrupt)
	defer cancel()

	opts := make([]grpc.ServerOption, 0)
	if enableServerKeepalive {
		opts = append(opts, grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    serverKeepaliveInterval,
			Timeout: 10 * time.Second,
		}))
	}
	if enforceKeepalive {
		opts = append(opts, grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             enforceKeepaliveInterval,
			PermitWithoutStream: true,
		}))
	}

	handler := NewHandler()
	srv := grpc.NewServer(opts...)
	proto.RegisterGreeterServer(srv, handler)
	grpc_health_v1.RegisterHealthServer(srv, NewHealthHandler())

	listner, err := net.Listen("tcp", fmt.Sprintf(":%v", listenPort))
	if err != nil {
		return err
	}

	go func() {
		<-ctx.Done()
		stopChan := make(chan struct{})
		go func() {
			srv.GracefulStop()
			close(stopChan)
		}()
		select {
		case <-stopChan:
			return
		case <-time.After(5 * time.Second):
			fmt.Fprintln(os.Stderr, "Error: failed to stop server gracefully")
			srv.Stop()
			return
		}
	}()
	fmt.Fprintf(os.Stderr, "Start to listen on %v", listner.Addr())
	return srv.Serve(listner)
}

func main() {
	flag.IntVar(&listenPort, "port", 8080, "Port number")
	flag.BoolVar(&enableServerKeepalive, "server-keepalive", false, "Enable server side keepalive")
	flag.BoolVar(&enforceKeepalive, "enforce-keepalive", false, "Permit client side high frequency probe")
	flag.DurationVar(&enforceKeepaliveInterval, "enforce-keepalive-interval", 10*time.Second, "Keepalive interval to enforce")
	flag.DurationVar(&serverKeepaliveInterval, "serevr-keepalive-interval", 10*time.Second, "Keepalive interval for server side")
	flag.Parse()
	if err := innerMain(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
