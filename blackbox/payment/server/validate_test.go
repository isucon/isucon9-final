package server

import (
	"testing"

	pb "payment/pb"
)

func TestValidator(t *testing.T) {
	s, err := NewNetworkServer()
	if err != nil {
		t.Fatalf("failed to create new server:%s", err)
	}

	t.Run("ValidateCard with correct parameters", func(t *testing.T) {
		card := &pb.CardInformation{
			CardNumber: "12345678",
			Cvv:        "123",
			ExpiryDate: "11/22",
		}
		err := s.ValidateCardInformation(&pb.RegistCardRequest{CardInformation: card})
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("%#v", err)
	})

	t.Run("ValidateCard with invalid card number", func(t *testing.T) {
		card := &pb.CardInformation{
			CardNumber: "1", //less
			Cvv:        "123",
			ExpiryDate: "11/22",
		}
		err := s.ValidateCardInformation(&pb.RegistCardRequest{CardInformation: card})
		if err == nil {
			t.Fatal("should fail")
		}
		t.Logf("%#v", err)

		card.CardNumber = "123456789" //over
		err = s.ValidateCardInformation(&pb.RegistCardRequest{CardInformation: card})
		if err == nil {
			t.Fatal("should fail")
		}
		t.Logf("%#v", err)

		card.CardNumber = "1234567A" //out of number
		err = s.ValidateCardInformation(&pb.RegistCardRequest{CardInformation: card})
		if err == nil {
			t.Fatal("should fail")
		}
		t.Logf("%#v", err)
	})

	t.Run("ValidateCard with invalid cvv", func(t *testing.T) {
		card := &pb.CardInformation{
			CardNumber: "12345678",
			Cvv:        "1", //less
			ExpiryDate: "11/22",
		}
		err := s.ValidateCardInformation(&pb.RegistCardRequest{CardInformation: card})
		if err == nil {
			t.Fatal("should fail")
		}
		t.Logf("%#v", err)

		card.Cvv = "1234" //over
		err = s.ValidateCardInformation(&pb.RegistCardRequest{CardInformation: card})
		if err == nil {
			t.Fatal("should fail")
		}
		t.Logf("%#v", err)

		card.Cvv = "12A" //out of number
		err = s.ValidateCardInformation(&pb.RegistCardRequest{CardInformation: card})
		if err == nil {
			t.Fatal("should fail")
		}
		t.Logf("%#v", err)
	})

	t.Run("ValidateCard with invalid ExpiryDate", func(t *testing.T) {
		card := &pb.CardInformation{
			CardNumber: "12345678",
			Cvv:        "123",
			ExpiryDate: "01/15", //past
		}
		err := s.ValidateCardInformation(&pb.RegistCardRequest{CardInformation: card})
		if err == nil {
			t.Fatal("should fail")
		}
		t.Logf("%#v", err)

		card.ExpiryDate = "13/22" //out of month
		err = s.ValidateCardInformation(&pb.RegistCardRequest{CardInformation: card})
		if err == nil {
			t.Fatal("should fail")
		}
		t.Logf("%#v", err)

		card.ExpiryDate = "1/22" //less digit
		err = s.ValidateCardInformation(&pb.RegistCardRequest{CardInformation: card})
		if err == nil {
			t.Fatal("should fail")
		}
		t.Logf("%#v", err)
	})

}
