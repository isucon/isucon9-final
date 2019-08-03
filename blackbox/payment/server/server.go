package server

import (
	"context"
	_ "net/http/pprof"
	"time"

	"github.com/chibiegg/isucon9-final/blackbox/payment/config"
	pb "github.com/chibiegg/isucon9-final/blackbox/payment/pb"
	"github.com/golang/protobuf/ptypes"
	"github.com/rs/xid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	PayInfoMap  map[string]*pb.PaymentInformation
	CardInfoMap map[string]*pb.CardInformation
}

func NewNetworkServer(c *config.Config) (*Server, error) {
	ns := &Server{
		PayInfoMap:  make(map[string]*pb.PaymentInformation, 1000000),
		CardInfoMap: make(map[string]*pb.CardInformation, 1000000),
	}
	return ns, nil
}

//クレジットカードのトークン発行(非保持化対応)
func (s *Server) RegistCard(ctx context.Context, req *pb.RegistCardRequest) (*pb.RegistCardResponse, error) {
	done := make(chan *pb.RegistCardResponse, 1)
	ec := make(chan error, 1)
	go func() {
		if req.CardInformation == nil {
			ec <- status.Errorf(codes.InvalidArgument, "Invalid POST data")
			return
		}
		guid := xid.New()

		s.CardInfoMap[guid.String()] = &pb.CardInformation{
			CardNumber: req.CardInformation.CardNumber,
			Cvv:        req.CardInformation.Cvv,
			ExpiryDate: req.CardInformation.ExpiryDate,
		}

		done <- &pb.RegistCardResponse{CardToken: guid.String(), IsOk: true}
	}()
	select {
	case r := <-done:
		return r, nil
	case err := <-ec:
		return &pb.RegistCardResponse{IsOk: false}, err
	}
}

//決済を行う
func (s *Server) ExecutePayment(ctx context.Context, req *pb.ExecutePaymentRequest) (*pb.ExecutePaymentResponse, error) {
	done := make(chan *pb.ExecutePaymentResponse, 1)
	ec := make(chan error, 1)
	go func() {
		if req.PaymentInformation == nil {
			ec <- status.Errorf(codes.InvalidArgument, "Invalid POST data")
			return
		}
		if _, ok := s.CardInfoMap[req.PaymentInformation.CardToken]; ok {
			date, err := ptypes.TimestampProto(time.Now())
			if err != nil {
				ec <- err
				return
			}
			guid := xid.New()

			s.PayInfoMap[guid.String()] = &pb.PaymentInformation{
				CardToken:  req.PaymentInformation.CardToken,
				Datetime:   date,
				Amount:     req.PaymentInformation.Amount,
				IsCanceled: false,
			}

			done <- &pb.ExecutePaymentResponse{PaymentId: guid.String(), IsOk: true}
		}
		ec <- status.Errorf(codes.NotFound, "Card_Token Not Found")
	}()
	select {
	case r := <-done:
		return r, nil
	case err := <-ec:
		return &pb.ExecutePaymentResponse{IsOk: false}, err
	}
}

//決済をキャンセルする
func (s *Server) CancelPayment(ctx context.Context, req *pb.CancelPaymentRequest) (*pb.CancelPaymentResponse, error) {
	done := make(chan struct{}, 1)
	ec := make(chan error, 1)
	go func(){
		if val, ok := s.PayInfoMap[req.PaymentId]; ok {
			val.IsCanceled = true
			
			done <- struct{}{}
		}
		ec <- status.Errorf(codes.NotFound, "PaymentID Not Found")
	}()
	select {
	case <- done:
		return &pb.CancelPaymentResponse{IsOk: true}, nil
	case err := <- ec:
		return &pb.CancelPaymentResponse{IsOk: false}, err 
	}
}

//決済情報を取得する
func (s *Server) GetPaymentInformation(ctx context.Context, req *pb.GetPaymentInformationRequest) (*pb.GetPaymentInformationResponse, error) {
	done := make(chan *pb.GetPaymentInformationResponse, 1)
	ec := make(chan error, 1)
	go func(){
		if val, ok := s.PayInfoMap[req.PaymentId]; ok {
			done <- &pb.GetPaymentInformationResponse{PaymentInformation: val, IsOk: true}
		}
		ec <- status.Errorf(codes.NotFound, "PaymentID Not Found")
	}()
	select {
	case r := <-done:
		return r, nil
	case err := <- ec:
		return &pb.GetPaymentInformationResponse{PaymentInformation: nil, IsOk: false}, err
	}
}
