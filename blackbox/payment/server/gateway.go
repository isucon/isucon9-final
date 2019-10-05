package server

import (
	"context"
	"net/http"
	_ "net/http/pprof"

	"payment/config"
	pb "payment/pb"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
)

func newGateway(c config.Config, ctx context.Context, opts ...runtime.ServeMuxOption) (http.Handler, error) {
	opts = []runtime.ServeMuxOption{
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{OrigName: true, EmitDefaults: true}),
	}
	mux := runtime.NewServeMux(opts...)
	dialOpts := []grpc.DialOption{grpc.WithInsecure()}
	conn, err := grpc.Dial(c.GrpcPort, dialOpts...)
	if err != nil {
		return nil, err
	}
	err = pb.RegisterPaymentServiceHandler(ctx, mux, conn)
	if err != nil {
		return nil, err
	}

	return mux, nil
}

func StartGRPCGateway(c config.Config, opts ...runtime.ServeMuxOption) error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	gw, err := newGateway(c, ctx, opts...)
	if err != nil {
		return err
	}

	return http.ListenAndServe(c.HttpPort, gw)
}
