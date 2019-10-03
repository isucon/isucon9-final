package server

import (
	"context"
	"net"
	"strconv"
	"testing"

	pb "payment/pb"

	"google.golang.org/grpc"
)

/*
	テスト内容
	・カード登録(1枚)
	・決済(1回/カードは上記のもの)
	・キャンセル(1回/決済IDは上記のもの)
	・キャンセルの確認(1回)
	・決済(3回/カードは上記のもの/バルクキャンセル用)
	・バルクキャンセル(3回)
	・キャンセルの確認(3回)
	・誤った内容のカード登録(4種類)
	・誤った内容の決済(1種類)
	・誤った内容のキャンセル(1種類)
	・誤った内容のバルクキャンセル(1種類)
	・ベンチマーカー用生データ取得(決済4回分のデータが出てくる)
*/
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
		if !r.PaymentInformation.IsCanceled {
			t.Fatal("should failed") // キャンセルされていないとここで落ちる
		}
		t.Logf("Canceled: %#v", r.PaymentInformation.IsCanceled)
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
		if r.Deleted != 3 {
			t.Fatalf("Failed. Expected:3 but %d\n", r.Deleted)
		}
		t.Logf("%#v", r)
	})

	t.Run("Check BulkCancelPayment", func(t *testing.T) {
		ctx := context.Background()
		for i := 0; i < 3; i++ {
			r, err := c.GetPaymentInformation(ctx, &pb.GetPaymentInformationRequest{PaymentId: payidlist[i]})
			if err != nil {
				t.Fatal(err)
			}
			if !r.PaymentInformation.IsCanceled {
				t.Fatal("should failed") // キャンセルされていないとここで落ちる
			}
			t.Logf("Canceled: %#v", r.PaymentInformation.IsCanceled)
		}
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
			t.Fatalf("Should fail. %s\n", err)
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

		if len(r.RawData) != 4 {
			t.Fatalf("Failed. Expected:4 but %d\n", len(r.RawData))
		}
	})

	t.Run("Initialize", func(t *testing.T) {
		ctx := context.Background()
		r, err := c.Initialize(ctx, &pb.InitializeRequest{})
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("%#v", r)
	})

	cardlist := make([]pb.CardInformation, 3)
	tokenlist := make([]string, 3)
	t.Log(len(cardlist))
	t.Run("[Ex]RegistCard for GetResult", func(t *testing.T) {
		ctx := context.Background()
		for i, _ := range cardlist {
			card := pb.CardInformation{
				CardNumber: strconv.Itoa(i + 11111111),
				Cvv:        strconv.Itoa(i + 111),
				ExpiryDate: "11/22",
			}
			cardlist[i] = card

			r, err := c.RegistCard(ctx, &pb.RegistCardRequest{CardInformation: &card})
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("%#v", r)
			tokenlist[i] = r.CardToken
		}
		t.Logf("%#v", cardlist)
	})

	t.Logf("%#v", tokenlist)

	payinfolist := make([]pb.PaymentInformation, 3)
	payidlist2 := make([]string, 3)
	t.Run("[Ex]ExecutePayment for GetResult", func(t *testing.T) {
		ctx := context.Background()
		for i, _ := range payinfolist {
			pay := pb.PaymentInformation{
				CardToken: tokenlist[i],
				Amount:    int32(i + 1000),
			}
			payinfolist[i] = pay

			r, err := c.ExecutePayment(ctx, &pb.ExecutePaymentRequest{PaymentInformation: &pay})
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("%#v", r)
			payidlist2[i] = r.PaymentId
		}
		t.Logf("%#v", payidlist2)
	})

	t.Run("[Ex]GetResult", func(t *testing.T) {
		ctx := context.Background()
		r, err := c.GetResult(ctx, &pb.GetResultRequest{})
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("%#v", r)

		if len(r.RawData) != 3 {
			t.Fatalf("Failed. Expected:3 but %d\n", len(r.RawData))
		}

		life := 0
		for i, v := range r.RawData {
			t.Logf("----------No %d----------\n", i)
			for life < 3 {
				if v.PaymentInformation.CardToken == payinfolist[life].CardToken {
					break
				}
				if life == 2 {
					t.Fatalf("Failed. Wrong card_token. Expected:%v, Got:%v\n", v.PaymentInformation.CardToken, payinfolist[life].CardToken)
				}
				life++
			}
			t.Logf("[Ex] CardToken Check OK. Expected:%v, Got:%v", v.PaymentInformation.CardToken, payinfolist[life].CardToken)

			life = 0
			for life < 3 {
				if v.PaymentInformation.Amount == payinfolist[life].Amount {
					break
				}
				if life == 2 {
					t.Fatalf("Failed. Wrong amount. Expected:%v, Got:%v\n", v.PaymentInformation.Amount, payinfolist[life].Amount)
				}
				life++
			}
			t.Logf("[Ex] Amount Check OK. Expected:%v, Got:%v", v.PaymentInformation.Amount, payinfolist[life].Amount)

			life = 0
			for life < 3 {
				if v.CardInformation.CardNumber == cardlist[life].CardNumber {
					break
				}
				if life == 2 {
					t.Fatalf("Failed. Wrong card_number. Expected:%v, Got:%v\n", v.CardInformation.CardNumber, cardlist[life].CardNumber)
				}
				life++
			}
			t.Logf("[Ex] CardNumber OK. Expected:%v, Got:%v", v.CardInformation.CardNumber, cardlist[life].CardNumber)

			life = 0
			for life < 3 {
				if v.CardInformation.Cvv == cardlist[life].Cvv {
					break
				}
				if life == 2 {
					t.Fatalf("Failed. Wrong cvv. Expected:%v, Got:%v\n", v.CardInformation.Cvv, cardlist[life].Cvv)
				}
				life++
			}
			t.Logf("[Ex] Cvv OK. Expected:%v, Got:%v", v.CardInformation.Cvv, cardlist[life].Cvv)

			life = 0
			for life < 3 {
				if v.CardInformation.ExpiryDate == cardlist[life].ExpiryDate {
					break
				}
				if life == 2 {
					t.Fatalf("Failed. Wrong expiry_date. Expected:%v, Got:%v\n", v.CardInformation.ExpiryDate, cardlist[life].ExpiryDate)
				}
				life++
			}
			t.Logf("[Ex] ExpiryDate OK. Expected:%v, Got:%v", v.CardInformation.ExpiryDate, cardlist[life].ExpiryDate)
		}
	})
}
