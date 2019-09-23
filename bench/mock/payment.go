package mock

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/chibiegg/isucon9-final/bench/payment"
)

// NOTE: https://github.com/grpc-ecosystem/grpc-gateway/blob/master/runtime/errors.go#L17,L54

type paymentMock struct {
	mu           sync.Mutex
	result       *payment.PaymentResult
	reservations map[string]*payment.PaymentInformation
}

func newPaymentMock() *paymentMock {
	return &paymentMock{
		result: &payment.PaymentResult{
			RawData: []*payment.RawData{},
		},
	}
}

func (m *paymentMock) addPaymentInformation() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.result.RawData = append(m.result.RawData, &payment.RawData{
		PaymentInfo: &payment.PaymentInformation{
			CardToken:  "XXXXXXXXXXXX",
			Datetime:   time.Now(),
			Amount:     int64(rand.Intn(200)),
			IsCanceled: rand.Intn(10000)%2 == 0,
		},
	})
}

func (m *paymentMock) Initialize() ([]byte, int) {
	return []byte(http.StatusText(http.StatusOK)), http.StatusOK
}

func (m *paymentMock) RegistCard() ([]byte, int) {
	b, err := json.Marshal(map[string]interface{}{
		"card_token": "11fopawkepfkawpefw",
		"is_ok":      true,
	})
	if err != nil {
		return []byte(http.StatusText(http.StatusInternalServerError)), http.StatusInternalServerError
	}

	return b, http.StatusOK
}

func (m *paymentMock) GetResult() ([]byte, int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.result.IsOK = true
	b, err := json.Marshal(m.result)
	if err != nil {
		return []byte(http.StatusText(http.StatusInternalServerError)), http.StatusInternalServerError
	}

	return b, http.StatusOK
}
