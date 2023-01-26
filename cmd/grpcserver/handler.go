package main

import (
	"context"
	"errors"
	"io"
	"strings"
	"time"

	"github.com/pddg/grpc-contour-timeout/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

type Handler struct {
	proto.UnimplementedGreeterServer
}

func NewHandler() *Handler {
	return &Handler{}
}

func sleep(ctx context.Context, waitFor time.Duration) error {
	select {
	case <-ctx.Done():
		err := ctx.Err()
		var errCode codes.Code
		if errors.Is(err, context.Canceled) {
			errCode = codes.Canceled
		} else if errors.Is(err, context.DeadlineExceeded) {
			errCode = codes.DeadlineExceeded
		} else {
			errCode = codes.Internal
		}
		return status.Error(errCode, "terminated")
	case <-time.After(waitFor):
	}
	return nil
}

func (h *Handler) Hi(ctx context.Context, in *proto.HiRequest) (*proto.Response, error) {
	waitDuration := time.Duration(in.DelaySec) * time.Second
	if err := sleep(ctx, waitDuration); err != nil {
		return nil, err
	}
	return &proto.Response{
		Message: in.Message,
	}, nil
}

func (h *Handler) Hello(in *proto.HelloRequest, srv proto.Greeter_HelloServer) error {
	waitDuration := time.Duration(in.DelaySec) * time.Second
	intervalDuration := time.Duration(in.IntervalSec) * time.Second
	ctx := srv.Context()
	if err := sleep(ctx, waitDuration); err != nil {
		return err
	}
	for {
		if err := srv.Send(&proto.Response{
			Message: in.Message,
		}); err != nil {
			return status.Error(codes.Internal, err.Error())
		}
		if err := sleep(ctx, intervalDuration); err != nil {
			return err
		}
	}
}

func (h *Handler) SeeYou(srv proto.Greeter_SeeYouServer) error {
	ctx := srv.Context()
	builder := strings.Builder{}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		in, err := srv.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}
		builder.WriteString(in.Message + " ")
	}
	if err := srv.SendAndClose(&proto.Response{
		Message: builder.String(),
	}); err != nil {
		return err
	}
	return nil
}

type HealthHandler struct {
	grpc_health_v1.UnimplementedHealthServer
}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) Check(_ context.Context, _ *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	return &grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_SERVING,
	}, nil
}

func (h *HealthHandler) Watch(_ *grpc_health_v1.HealthCheckRequest, _ grpc_health_v1.Health_WatchServer) error {
	panic("not implemented") // TODO: Implement
}
