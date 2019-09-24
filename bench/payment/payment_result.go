package payment

import "time"

type PaymentInformation struct {
	CardToken     string    `json:"card_token"`
	ReservationID int       `json:"reservation_id"`
	Datetime      time.Time `json:"datetime"`
	Amount        int64     `json:"amount"`
	IsCanceled    bool      `json:"is_canceled"`
}

type CardInformation struct {
	CardNumber string `json:"card_number"`
	Cvv        string `json:"cvv"`
	ExpiryDate string `json:"expiry_date"`
}

type RawData struct {
	PaymentInfo *PaymentInformation `json:"payment_information"`
	CardInfo    *CardInformation    `json:"card_information"`
}

type PaymentResult struct {
	RawData []*RawData `json:"raw_data"`
	IsOK    bool       `json:"is_ok"`
}

type RegistCardResponse struct {
	CardToken string `json:"card_token"`
	IsOK      bool   `json:"is_ok"`
}
