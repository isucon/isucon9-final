package server

import (
	"context"
	"net"
	"testing"

	pb "payment/pb"

	"google.golang.org/grpc"
)

func TestServer(t *testing.T) {
	//setup grpc server
	lis, err := net.Listen("tcp", ":5001")
	if err != nil {
		t.Fatalf("listen error: %v", err)
	}
	g := grpc.NewServer()

	s, err := NewNetworkServer()
	if err != nil {
		t.Fatalf("failed to create new server:%s", err)
	}
	pb.RegisterPaymentServiceServer(g, s)
	go g.Serve(lis)

	conn, err := grpc.Dial(lis.Addr().String(), grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}
	c := pb.NewPaymentServiceClient(conn)

	var token string
	t.Run("RegistCard", func(t *testing.T) {
		ctx := context.Background()
		card := &pb.CardInformation{
			CardNumber: "12345678",
			Cvv:        "123",
			ExpiryDate: "11/22",
		}
		r, err := c.RegistCard(ctx, &pb.RegistCardRequest{CardInformation: card})
		if err != nil {
			t.Fatal(err)
		}
		token = r.CardToken
		t.Logf("%#v", r)
	})

	var payid string
	t.Run("ExecutePayment", func(t *testing.T) {
		ctx := context.Background()
		pay := &pb.PaymentInformation{
			CardToken: token,
			Amount:    9800,
		}
		r, err := c.ExecutePayment(ctx, &pb.ExecutePaymentRequest{PaymentInformation: pay})
		if err != nil {
			t.Fatal(err)
		}
		payid = r.PaymentId
		t.Logf("%#v", r)
	})

	t.Run("CancelPayment", func(t *testing.T) {
		ctx := context.Background()
		r, err := c.CancelPayment(ctx, &pb.CancelPaymentRequest{PaymentId: payid})
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("%#v", r)
	})

	t.Run("GetPaymentInformation", func(t *testing.T) {
		ctx := context.Background()
		r, err := c.GetPaymentInformation(ctx, &pb.GetPaymentInformationRequest{PaymentId: payid})
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("%#v", r)
	})

	payidlist := make([]string, 0)
	t.Run("ExecutePayment for BulkCancel", func(t *testing.T) {
		ctx := context.Background()
		pay := &pb.PaymentInformation{
			CardToken: token,
			Amount:    9800,
		}
		for i := 0; i < 3; i++ {
			r, err := c.ExecutePayment(ctx, &pb.ExecutePaymentRequest{PaymentInformation: pay})
			if err != nil {
				t.Fatal(err)
			}
			payidlist = append(payidlist, r.PaymentId)
			t.Logf("%#v", r)
		}
	})

	t.Run("BulkCancelPayment", func(t *testing.T) {
		ctx := context.Background()
		r, err := c.BulkCancelPayment(ctx, &pb.BulkCancelPaymentRequest{PaymentId: payidlist})
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("%#v", r)
	})

	t.Run("RegistCard with invalid parameters", func(t *testing.T) {
		ctx := context.Background()
		card := &pb.CardInformation{
			CardNumber: "1234567", //invalid
			Cvv:        "123",
			ExpiryDate: "11/22",
		}
		r, err := c.RegistCard(ctx, &pb.RegistCardRequest{CardInformation: card})
		if err == nil {
			t.Fatal("should failed")
		}
		t.Logf("%#v", r)

		card = &pb.CardInformation{
			CardNumber: "12345678",
			Cvv:        "12", //invalid
			ExpiryDate: "11/22",
		}
		r, err = c.RegistCard(ctx, &pb.RegistCardRequest{CardInformation: card})
		if err == nil {
			t.Fatal("should failed")
		}
		t.Logf("%#v", r)

		card = &pb.CardInformation{
			CardNumber: "12345678",
			Cvv:        "123",
			ExpiryDate: "01/18", //invalid
		}
		r, err = c.RegistCard(ctx, &pb.RegistCardRequest{CardInformation: card})
		if err == nil {
			t.Fatal("should failed")
		}
		t.Logf("%#v", r)

		card = &pb.CardInformation{
			CardNumber: "1234567A", //invalid
			Cvv:        "12A",      //invalid
			ExpiryDate: "01/18",    //invalid
		}
		r, err = c.RegistCard(ctx, &pb.RegistCardRequest{CardInformation: card})
		if err == nil {
			t.Fatal("should failed")
		}
		t.Logf("%#v", r)
	})

	t.Run("ExecutePayment with invalid parameters", func(t *testing.T) {
		ctx := context.Background()
		pay := &pb.PaymentInformation{
			CardToken: "hoge", //invalid
			Amount:    9800,
		}
		r, err := c.ExecutePayment(ctx, &pb.ExecutePaymentRequest{PaymentInformation: pay})
		if err == nil {
			t.Fatal("should failed")
		}
		t.Logf("%#v", r)
	})

	t.Run("CancelPayment with invalid parameters", func(t *testing.T) {
		ctx := context.Background()
		payid = "a" // insert invalid paymentid
		r, err := c.CancelPayment(ctx, &pb.CancelPaymentRequest{PaymentId: payid})
		if err == nil {
			t.Fatalf("Should fail. %s\n",err)
		}
		t.Logf("%#v", r)
	})

	t.Run("BulkCancelPayment with invalid paparameters", func(t *testing.T) {
		ctx := context.Background()
		payidlist[1] = "a" // insert invalid paymentid
		r, err := c.BulkCancelPayment(ctx, &pb.BulkCancelPaymentRequest{PaymentId: payidlist})
		if r.Deleted != 2 {
			t.Fatalf("Failed. Expected:2 but %d\n", r.Deleted)
		}
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("%#v", r)
	})

	t.Run("GetResult", func(t *testing.T) {
		ctx := context.Background()
		r, err := c.GetResult(ctx, &pb.GetResultRequest{})
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("%#v", r)
	})

	t.Run("Initialize", func(t *testing.T) {
		ctx := context.Background()
		r, err := c.Initialize(ctx, &pb.InitializeRequest{})
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("%#v", r)
	})

}
