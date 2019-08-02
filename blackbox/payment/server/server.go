package server

import (
	_ "net/http/pprof"
	"context"

	"github.com/chibiegg/isucon9-final/blackbox/payment/config"
	pb "github.com/chibiegg/isucon9-final/blackbox/payment/pb"
)

type Server struct{}

func NewNetworkServer(c *config.Config) (*Server, error) {
	ns := &Server{}
	return ns, nil
}

//決済を行う
func (s *Server) ExecutePayment(ctx context.Context, req *pb.ExecutePaymentRequest) (*pb.ExecutePaymentResponse, error) {
	return nil,nil
}

//決済をキャンセルする
func (s *Server) CancelPayment(ctx context.Context, req *pb.CancelPaymentRequest) (*pb.CancelPaymentResponse, error) {
	return nil,nil
}

//決済情報を取得する
func (s *Server) GetPaymentInformation(ctx context.Context, req *pb.GetPaymentInformationRequest) (*pb.GetPaymentInformationResponse, error) {
	return nil,nil
}