package server

import (
	_ "net/http/pprof"
	"context"
	"time"
	"sync"
	"log"

	"payment/config"
	pb "payment/pb"
	"github.com/golang/protobuf/ptypes"
	"github.com/rs/xid"
	"github.com/nu7hatch/gouuid"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/codes"
)

type Server struct {
	PayInfoMap  map[string]pb.PaymentInformation
	CardInfoMap map[string]pb.CardInformation
	mu sync.RWMutex
}

func NewNetworkServer(c *config.Config) (*Server, error) {
	ns := &Server{
		PayInfoMap:  make(map[string]pb.PaymentInformation, 1000000),
		CardInfoMap: make(map[string]pb.CardInformation, 1000000),
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
		id, err := uuid.NewV4()
		if err != nil {
			ec <- status.Errorf(codes.Internal, "Internal Error, Generate UUID")
			return
		}

		s.mu.Lock()
		s.CardInfoMap[id.String()] = pb.CardInformation{
			CardNumber: req.CardInformation.CardNumber,
			Cvv:        req.CardInformation.Cvv,
			ExpiryDate: req.CardInformation.ExpiryDate,
		}
		s.mu.Unlock()

		done <- &pb.RegistCardResponse{CardToken: id.String(), IsOk: true}
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

		s.mu.RLock()
		_, ok := s.CardInfoMap[req.PaymentInformation.CardToken]
		s.mu.RUnlock()
		if ok {
			date, err := ptypes.TimestampProto(time.Now())
			if err != nil {
				ec <- err
				return
			}
			guid := xid.New()

			s.mu.Lock()
			s.PayInfoMap[guid.String()] = pb.PaymentInformation{
				CardToken:  req.PaymentInformation.CardToken,
				Datetime:   date,
				Amount:     req.PaymentInformation.Amount,
				IsCanceled: false,
			}
			s.mu.Unlock()

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
		s.mu.RLock()
		id, ok := s.PayInfoMap[req.PaymentId]
		s.mu.RUnlock()
		if ok {
			id.IsCanceled = true
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

//バルクで決済をキャンセルする
func (s *Server) BulkCancelPayment(ctx context.Context, req *pb.BulkCancelPaymentRequest) (*pb.BulkCancelPaymentResponse, error) {
	return nil,nil
}

//決済情報を取得する
func (s *Server) GetPaymentInformation(ctx context.Context, req *pb.GetPaymentInformationRequest) (*pb.GetPaymentInformationResponse, error) {
	done := make(chan *pb.GetPaymentInformationResponse, 1)
	ec := make(chan error, 1)
	go func(){
		s.mu.RLock()
		id, ok := s.PayInfoMap[req.PaymentId]
		s.mu.RUnlock()
		if ok {
			done <- &pb.GetPaymentInformationResponse{PaymentInformation: &id, IsOk: true}
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

//メモリ初期化
func (s *Server) Initialize(ctx context.Context, req *pb.InitializeRequest) (*pb.InitializeResponse, error) {
	done := make(chan struct{}, 1)
	ec := make(chan error, 1)
	go func(){
		s.mu.Lock()
		s.PayInfoMap = nil
		s.CardInfoMap = nil
		s.PayInfoMap = make(map[string]pb.PaymentInformation, 1000000)
		s.CardInfoMap = make(map[string]pb.CardInformation, 1000000)
		s.mu.Unlock()
		done <- struct{}{}
	}()
	select {
	case <-done:
		return &pb.InitializeResponse{IsOk: true}, nil
	case err := <- ec:
		return &pb.InitializeResponse{IsOk: false}, err
	}
}

//ベンチマーカー用結果取得API
func (s *Server) GetResult(ctx context.Context, req *pb.GetResultRequest) (*pb.GetResultResponse, error) {
	done := make(chan *pb.GetResultResponse, 1)
	ec := make(chan error, 1)
	go func(){
		log.Printf("Card count: %d\n",len(s.CardInfoMap))
		log.Printf("Payment count: %d\n",len(s.PayInfoMap))

		payinfo := &pb.PaymentInformation{}
		cardinfo := &pb.CardInformation{}
		raw := []*pb.RawData{}
		s.mu.RLock()
		for k, v := range s.PayInfoMap {
			payinfo.CardToken = k
			payinfo.Datetime =  v.Datetime
			payinfo.Amount = v.Amount
			payinfo.IsCanceled = v.IsCanceled

			cardinfo.CardNumber = s.CardInfoMap[k].CardNumber
			cardinfo.Cvv = s.CardInfoMap[k].Cvv
			cardinfo.ExpiryDate = s.CardInfoMap[k].ExpiryDate

			rawdata := &pb.RawData {
				PaymentInformation: payinfo,
				CardInformation: cardinfo,
			}
			raw = append(raw, rawdata)
		}
		s.mu.RUnlock()
		done <- &pb.GetResultResponse{RawData: raw, IsOk: true}
	}()
	select {
	case r := <-done:
		return r, nil
	case err := <- ec:
		return &pb.GetResultResponse{IsOk: false}, err
	}
}