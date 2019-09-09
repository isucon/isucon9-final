package payment

import "context"

type Payment struct {
	BaseURL string
}

func NewPayment(baseURL string) *Payment {
	return &Payment{
		BaseURL: baseURL,
	}
}

func (p *Payment) Result(ctx context.Context) error {
	return nil
}
