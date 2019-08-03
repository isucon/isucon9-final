package payment

type RegistCardRequest struct {
	CardNumber string `json:"card_number"`
	Cvv        string `json:"cvv"`
	ExpiryDate string `json:"expiry_date"`
}
