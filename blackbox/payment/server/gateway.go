package server

import (
	"context"
	"net/http"
	_ "net/http/pprof"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"payment/config"
	pb "payment/pb"
	"google.golang.org/grpc"
)

func newGateway(c config.Config, ctx context.Context, opts ...runtime.ServeMuxOption) (http.Handler, error) {
	opts = []runtime.ServeMuxOption{
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{OrigName: true, EmitDefaults: true}),
	}
	mux := runtime.NewServeMux(opts...)
	allow := allowCORS(mux)
	dialOpts := []grpc.DialOption{grpc.WithInsecure()}
	conn, err := grpc.Dial(c.GrpcPort, dialOpts...)
	if err != nil {
		return nil, err
	}
	err = pb.RegisterPaymentServiceHandler(ctx, mux, conn)
	if err != nil {
		return nil, err
	}

	return allow, nil
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

func allowCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			if r.Method == "OPTIONS" && r.Header.Get("Access-Control-Request-Method") != "" {
				headers := []string{"Content-Type", "Accept", "Authorization"}
				w.Header().Set("Access-Control-Allow-Headers", strings.Join(headers, ","))
				methods := []string{"GET", "HEAD", "POST", "PUT", "DELETE", "OPTIONS"}
				w.Header().Set("Access-Control-Allow-Methods", strings.Join(methods, ","))
				return
			}
		}
		h.ServeHTTP(w, r)
	})
}