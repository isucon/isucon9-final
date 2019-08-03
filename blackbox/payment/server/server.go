package server

import (
	_ "net/http/pprof"
	"context"
	"time"

	"github.com/chibiegg/isucon9-final/blackbox/payment/config"
	pb "github.com/chibiegg/isucon9-final/blackbox/payment/pb"
	"github.com/golang/protobuf/ptypes"
	"github.com/rs/xid"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/codes"
)

type Server struct{
	PayInfoMap	map[string]*pb.PaymentInformation
	CardInfoMap map[string]*pb.CardInformation
}

func NewNetworkServer(c *config.Config) (*Server, error) {
	m := make(map[string]*pb.PaymentInformation,1000000)
	c := make(map[string]*pb.CardInformation,1000000)
	ns := &Server{
		PayInfoMap: m,
		CardInfomap: c,
	}
	return ns, nil
}

//クレジットカードのトークン発行(非保持化対応)
func (s *Server) RegistCard(ctx context.Context, req *pb.RegistCardRequest) (*pb.RegistCardResponse, error) {
	guid := xid.New()
	
	cinfo := &pb.CardInformation{
		CardNumber: req.CardInformation.CardNumber,
		Cvv: req.CardInformation.Cvv,
		ExpiryDate: req.CardInformation.ExpiryDate,
	}

	s.CardInfoMap[guid.String()] = cinfo

	return &pb.RegistCardResponse{
		Token: guid.String(),
		IsOk: true,
	},nil
}

//決済を行う
func (s *Server) ExecutePayment(ctx context.Context, req *pb.ExecutePaymentRequest) (*pb.ExecutePaymentResponse, error) {
	date, err := ptypes.TimestampProto(time.Now())
	if err != nil {
		return nil,nil
	}
	guid := xid.New()
	
	info := &pb.PaymentInformation{
		CardNumber: req.PaymentInformation.CardNumber,
		Datetime: date,
		Cvv: req.PaymentInformation.Cvv,
		Amount: req.PaymentInformation.Amount,
		IsCanceled: false,
	}

	s.PayInfoMap[guid.String()] = info

	return &pb.ExecutePaymentResponse{
		PaymentId: guid.String(),
		IsOk: true,
	},nil
}

//決済をキャンセルする
func (s *Server) CancelPayment(ctx context.Context, req *pb.CancelPaymentRequest) (*pb.CancelPaymentResponse, error) {
	if val, ok := s.PayInfoMap[req.PaymentId]; ok {
		val.IsCanceled = true
		return &pb.CancelPaymentResponse{
			IsOk: true,
		},nil
	}
	
	return &pb.CancelPaymentResponse{
		IsOk: false,
	},status.Errorf(codes.NotFound,"PaymenID Not Found")
}

//決済情報を取得する
func (s *Server) GetPaymentInformation(ctx context.Context, req *pb.GetPaymentInformationRequest) (*pb.GetPaymentInformationResponse, error) {
	if val, ok := s.PayInfoMap[req.PaymentId]; ok {
		return &pb.GetPaymentInformationResponse{
			PaymentInformation: val,
			IsOk: true,
		},nil
	}

	return &pb.GetPaymentInformationResponse{
		PaymentInformation: nil,
		IsOk: false,
	},status.Errorf(codes.NotFound,"PaymenID Not Found")
}
