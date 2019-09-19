package server

import (
	"context"
	"log"
	_ "net/http/pprof"
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/nu7hatch/gouuid"
	"github.com/rs/xid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"payment/config"
	pb "payment/pb"
)

var rawDataPool = sync.Pool{
	New: func() interface{} {
		return &pb.RawData{
			PaymentInformation: &pb.PaymentInformation{},
			CardInformation:    &pb.CardInformation{},
		}
	},
}

func getRawData() *pb.RawData {
	return rawDataPool.Get().(*pb.RawData)
}

func putRawData(rawData *pb.RawData) {
	rawDataPool.Put(rawData)
}

func init() {
	for i := 0; i < 1000000; i++ {
		putRawData(getRawData())
	}
}

type Server struct {
	PayInfoMap  map[string]pb.PaymentInformation
	CardInfoMap map[string]pb.CardInformation
	mu          sync.RWMutex
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
		err := s.ValidateCardInformation(req)
		if err != nil {
			ec <- status.Errorf(codes.InvalidArgument, err.Error())
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
	go func() {
		s.mu.RLock()
		paydata, ok := s.PayInfoMap[req.PaymentId]
		s.mu.RUnlock()
		if ok {
			paydata.IsCanceled = true
			done <- struct{}{}
		}

		ec <- status.Errorf(codes.NotFound, "PaymentID Not Found")
	}()
	select {
	case <-done:
		return &pb.CancelPaymentResponse{IsOk: true}, nil
	case err := <-ec:
		return &pb.CancelPaymentResponse{IsOk: false}, err
	}
}

//バルクで決済をキャンセルする
func (s *Server) BulkCancelPayment(ctx context.Context, req *pb.BulkCancelPaymentRequest) (*pb.BulkCancelPaymentResponse, error) {
	done := make(chan int32, 1)
	ec := make(chan int32, 1)
	go func() {
		s.mu.Lock()
		if len(req.PaymentId) < 1 {
			ec <- 0
			return
		}

		var i int32
		for _, v := range req.PaymentId {
			paydata, ok := s.PayInfoMap[v]
			if ok {
				paydata.IsCanceled = true
			} else {
				i--
			}
			i++
		}
		done <- i
	}()
	select {
	case num := <-done:
		s.mu.Unlock()
		time.Sleep(time.Second * 1)
		return &pb.BulkCancelPaymentResponse{Deleted: num}, nil
	case num := <-ec:
		s.mu.Unlock()
		time.Sleep(time.Second * 1)
		return &pb.BulkCancelPaymentResponse{Deleted: num}, nil
	}
	return nil, nil
}

//決済情報を取得する
func (s *Server) GetPaymentInformation(ctx context.Context, req *pb.GetPaymentInformationRequest) (*pb.GetPaymentInformationResponse, error) {
	done := make(chan *pb.GetPaymentInformationResponse, 1)
	ec := make(chan error, 1)
	go func() {
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
	case err := <-ec:
		return &pb.GetPaymentInformationResponse{PaymentInformation: nil, IsOk: false}, err
	}
}

//メモリ初期化
func (s *Server) Initialize(ctx context.Context, req *pb.InitializeRequest) (*pb.InitializeResponse, error) {
	done := make(chan struct{}, 1)
	ec := make(chan error, 1)
	go func() {
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
	case err := <-ec:
		return &pb.InitializeResponse{IsOk: false}, err
	}
}

//ベンチマーカー用結果取得API
func (s *Server) GetResult(ctx context.Context, req *pb.GetResultRequest) (*pb.GetResultResponse, error) {
	done := make(chan *pb.GetResultResponse, 1)
	ec := make(chan error, 1)
	go func() {
		log.Printf("Card count: %d\n", len(s.CardInfoMap))
		log.Printf("Payment count: %d\n", len(s.PayInfoMap))

		raw := []*pb.RawData{}
		s.mu.RLock()
		for k, v := range s.PayInfoMap {
			rawData := getRawData()
			defer putRawData(rawData)
			rawData.PaymentInformation.CardToken = k
			rawData.PaymentInformation.Datetime = v.Datetime
			rawData.PaymentInformation.Amount = v.Amount
			rawData.PaymentInformation.IsCanceled = v.IsCanceled

			rawData.CardInformation.CardNumber = s.CardInfoMap[k].CardNumber
			rawData.CardInformation.Cvv = s.CardInfoMap[k].Cvv
			rawData.CardInformation.ExpiryDate = s.CardInfoMap[k].ExpiryDate

			raw = append(raw, rawData)
		}
		s.mu.RUnlock()
		done <- &pb.GetResultResponse{RawData: raw, IsOk: true}
	}()
	select {
	case r := <-done:
		return r, nil
	case err := <-ec:
		return &pb.GetResultResponse{IsOk: false}, err
	}
}
