package isutrain

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// Mock は `isutrain` のモック実装です
type Mock struct {
}

// Login はログイン処理結果を返します
func (m *Mock) Login(req *http.Request) ([]byte, int) {
	if err := req.ParseForm(); err != nil {
		return []byte(http.StatusText(http.StatusBadRequest)), http.StatusBadRequest
	}

	form := req.Form
	log.Println(form)

	// var (
	// 	username = form.Get("username")
	// 	password = form.Get("password")
	// )

	return []byte(http.StatusText(http.StatusOK)), http.StatusOK
}

// SearchTrains は新幹線検索結果を返します
func (m *Mock) SearchTrains(req *http.Request) ([]byte, int) {
	q := req.URL.Query()
	if q == nil {
		return []byte(http.StatusText(http.StatusBadRequest)), http.StatusBadRequest
	}

	// TODO: 検索クエリを受け取る
	// いつ(use_at)、どこからどこまで(from, to), 人数(number of people) で結果が帰って来れば良い
	// 日付を投げてきていて、DB称号可能などこからどこまでがあればいい
	// どこからどこまでは駅名を書く(IDはユーザから見たらまだわからない)
	// var (
	// 	useAt = q["use_at"]
	// 	from  = q["from"]
	// 	to    = q["to"]
	// )

	b, err := json.Marshal(&Trains{
		&Train{Class: "のぞみ", Name: "96号", StartStationID: 1, EndStationID: 2},
		&Train{Class: "こだま", Name: "96号", StartStationID: 3, EndStationID: 4},
	})
	if err != nil {
		return []byte(http.StatusText(http.StatusInternalServerError)), http.StatusInternalServerError
	}

	return b, http.StatusOK
}

// ListSeats は列車の席一覧を返します
func (m *Mock) ListSeats(req *http.Request) ([]byte, int) {
	q := req.URL.Query()
	if q == nil {
		return []byte(http.StatusText(http.StatusBadRequest)), http.StatusBadRequest
	}

	// 列車特定情報を受け取る

	// 適当な席を返す
	b, err := json.Marshal(&Seats{
		&Seat{},
	})
	if err != nil {
		return []byte(http.StatusText(http.StatusInternalServerError)), http.StatusInternalServerError
	}

	return b, http.StatusOK
}

// Reserve は座席予約を実施し、結果を返します
func (m *Mock) Reserve(req *http.Request) ([]byte, int) {
	// 
}

func (m *Mock) CommitReservation(req *http.Request) ([]byte, int) {
	// 
}

// ListReservations はアカウントにひもづく予約履歴を返します
func (m *Mock) ListReservations(req *http.Request) ([]byte, int) {
	b, err := json.Marshal(Reservations{
		&Reservation{ID: 1111, PaymentMethod: string(CreditCard), Status: string(Pending), ReserveAt: time.Now()},
	})
	if err != nil {
		return []byte(http.StatusText(http.StatusBadRequest)), http.StatusBadRequest)
	}

	return b, http.StatusOK
}

// httpmockのエンドポイントを登録する
func RegisterMockEndpoints() {

}
