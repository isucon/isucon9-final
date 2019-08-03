package isutrain

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/chibiegg/isucon9-final/bench/util"
	"github.com/jarcoal/httpmock"
)

// Mock は `isutrain` のモック実装です
type Mock struct{}

// Register はユーザ登録を行います
func (m *Mock) Register(req *http.Request) ([]byte, int) {
	if err := req.ParseForm(); err != nil {
		return []byte(http.StatusText(http.StatusBadRequest)), http.StatusBadRequest
	}

	var (
		username = req.Form.Get("username")
		password = req.Form.Get("password")
		// TODO: 他にも登録情報を追加
	)
	if len(username) == 0 || len(password) == 0 {
		return []byte(http.StatusText(http.StatusBadRequest)), http.StatusBadRequest
	}

	return []byte(http.StatusText(http.StatusAccepted)), http.StatusAccepted
}

// Login はログイン処理結果を返します
func (m *Mock) Login(req *http.Request) ([]byte, int) {
	if err := req.ParseForm(); err != nil {
		return []byte(http.StatusText(http.StatusBadRequest)), http.StatusBadRequest
	}

	var (
		username = req.Form.Get("username")
		password = req.Form.Get("password")
	)
	if len(username) == 0 || len(password) == 0 {
		return []byte(http.StatusText(http.StatusBadRequest)), http.StatusBadRequest
	}

	return []byte(http.StatusText(http.StatusAccepted)), http.StatusAccepted
}

// SearchTrains は新幹線検索結果を返します
func (m *Mock) SearchTrains(req *http.Request) ([]byte, int) {
	query := req.URL.Query()
	if query == nil {
		return []byte(http.StatusText(http.StatusBadRequest)), http.StatusBadRequest
	}

	// TODO: 検索クエリを受け取る
	// いつ(use_at)、どこからどこまで(from, to), 人数(number of people) で結果が帰って来れば良い
	// 日付を投げてきていて、DB称号可能などこからどこまでがあればいい
	// どこからどこまでは駅名を書く(IDはユーザから見たらまだわからない)
	useAt, err := util.ParseISO8601(query.Get("use_at"))
	if err != nil {
		return []byte(http.StatusText(http.StatusBadRequest)), http.StatusBadRequest
	}
	if useAt.IsZero() {
		return []byte(http.StatusText(http.StatusBadRequest)), http.StatusBadRequest
	}

	var (
		from = query.Get("from")
		to   = query.Get("to")
	)
	if len(from) == 0 || len(to) == 0 {
		return []byte(http.StatusText(http.StatusBadRequest)), http.StatusBadRequest
	}

	b, err := json.Marshal(&Trains{
		&Train{Class: "のぞみ", Name: "96号", StartStationID: 1, EndStationID: 2},
		&Train{Class: "こだま", Name: "96号", StartStationID: 3, EndStationID: 4},
	})
	if err != nil {
		return []byte(http.StatusText(http.StatusInternalServerError)), http.StatusInternalServerError
	}

	return b, http.StatusOK
}

// ListTrainSeats は列車の席一覧を返します
func (m *Mock) ListTrainSeats(req *http.Request) ([]byte, int) {
	q := req.URL.Query()
	if q == nil {
		return []byte(http.StatusText(http.StatusBadRequest)), http.StatusBadRequest
	}

	// 列車特定情報を受け取る
	var (
		trainClass = q["train_class"]
		trainName  = q["train_name"]
	)
	if len(trainClass) == 0 || len(trainName) == 0 {
		return []byte(http.StatusText(http.StatusBadRequest)), http.StatusBadRequest
	}

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
	// 予約情報を受け取って、予約できたかを返す

	// FIXME: ユーザID
	// 複数の座席指定で予約するかもしれない
	// なので、予約には複数の座席予約が紐づいている

	// NOTE: とりあえず、パラメータガン無視でPOSTできるところ先にやる
	return []byte(http.StatusText(http.StatusAccepted)), http.StatusAccepted
}

// CommitReservation は予約を確定します
func (m *Mock) CommitReservation(req *http.Request) ([]byte, int) {
	// 予約IDを受け取って、確定するだけ
	if err := req.ParseForm(); err != nil {
		return []byte(http.StatusText(http.StatusBadRequest)), http.StatusBadRequest
	}

	_, err := strconv.ParseInt(req.Form.Get("reservation_id"), 10, 64)
	if err != nil {
		return []byte(http.StatusText(http.StatusBadRequest)), http.StatusBadRequest
	}

	return []byte(http.StatusText(http.StatusAccepted)), http.StatusAccepted
}

// CancelReservation は予約をキャンセルします
func (m *Mock) CancelReservation(req *http.Request) ([]byte, int) {
	// 予約IDを受け取って
	if err := req.ParseForm(); err != nil {
		return []byte(http.StatusText(http.StatusBadRequest)), http.StatusBadRequest
	}

	var (
		reservationID = req.Form.Get("reservation_id")
	)
	if len(reservationID) == 0 {
		return []byte(http.StatusText(http.StatusBadRequest)), http.StatusBadRequest
	}

	return []byte(http.StatusText(http.StatusNoContent)), http.StatusNoContent
}

// ListReservations はアカウントにひもづく予約履歴を返します
func (m *Mock) ListReservations(req *http.Request) ([]byte, int) {
	b, err := json.Marshal(SeatReservations{
		&SeatReservation{ID: 1111, PaymentMethod: string(CreditCard), Status: string(Pending), ReserveAt: time.Now()},
	})
	if err != nil {
		return []byte(http.StatusText(http.StatusBadRequest)), http.StatusBadRequest
	}

	return b, http.StatusOK
}

// RegisterMockEndpoints はhttpmockのエンドポイントを登録する
// NOTE: httpmock.Activate, httpmock.Deactivateは別途実施する必要があります
func RegisterMockEndpoints() (*Mock, []string) {
	var (
		m = &Mock{}
		// FIXME: endpoints と Responderの２箇所書き換えるのダメダメなので、いい感じにする
		// 構造体ベースにして、いてレートしてRegisterResponderする
		endpoints = []string{
			"/register",
			"/login",
			"/train/search",
			"/train/seats",
			"/reserve",
			"/reservation",
			"/reservation/:id/commit",
			"/reservation/:id/cancel",
		}
	)

	// GET
	httpmock.RegisterResponder("GET", "/train/search", func(req *http.Request) (*http.Response, error) {
		body, status := m.SearchTrains(req)
		return httpmock.NewBytesResponse(status, body), nil
	})
	httpmock.RegisterResponder("GET", "/train/seats", func(req *http.Request) (*http.Response, error) {
		body, status := m.ListTrainSeats(req)
		return httpmock.NewBytesResponse(status, body), nil
	})
	httpmock.RegisterResponder("GET", "/reservation", func(req *http.Request) (*http.Response, error) {
		body, status := m.ListReservations(req)
		return httpmock.NewBytesResponse(status, body), nil
	})

	// POST
	httpmock.RegisterResponder("POST", "/register", func(req *http.Request) (*http.Response, error) {
		body, status := m.Register(req)
		return httpmock.NewBytesResponse(status, body), nil
	})
	httpmock.RegisterResponder("POST", "/login", func(req *http.Request) (*http.Response, error) {
		body, status := m.Login(req)
		return httpmock.NewBytesResponse(status, body), nil
	})
	httpmock.RegisterResponder("POST", "/reserve", func(req *http.Request) (*http.Response, error) {
		body, status := m.Reserve(req)
		return httpmock.NewBytesResponse(status, body), nil
	})
	httpmock.RegisterResponder("POST", `=~^/reservation/\d+/commit\z`, func(req *http.Request) (*http.Response, error) {
		body, status := m.CommitReservation(req)
		return httpmock.NewBytesResponse(status, body), nil
	})

	// DELETE
	httpmock.RegisterResponder("DELETE", `=~^/reservation/\d+/cancel\z`, func(req *http.Request) (*http.Response, error) {
		body, status := m.CancelReservation(req)
		return httpmock.NewBytesResponse(status, body), nil
	})

	return m, endpoints
}
