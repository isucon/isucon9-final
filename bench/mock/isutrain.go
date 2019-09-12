package mock

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/chibiegg/isucon9-final/bench/internal/util"
	"github.com/chibiegg/isucon9-final/bench/isutrain"
	"github.com/jarcoal/httpmock"
)

// Mock は `isutrain` のモック実装です
type Mock struct {
	delay time.Duration
}

func (m *Mock) SetDelay(d time.Duration) {
	m.delay = d
}

func (m *Mock) Delay() time.Duration {
	return m.delay
}

func (m *Mock) Initialize(req *http.Request) ([]byte, int) {
	<-time.After(m.delay)
	log.Println("[mock] Initialize")
	return []byte(http.StatusText(http.StatusAccepted)), http.StatusAccepted
}

// Register はユーザ登録を行います
func (m *Mock) Register(req *http.Request) ([]byte, int) {
	<-time.After(m.delay)
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
	log.Printf("[mock] Register: username=%s, password=%s\n", username, password)

	return []byte(http.StatusText(http.StatusAccepted)), http.StatusAccepted
}

// Login はログイン処理結果を返します
func (m *Mock) Login(req *http.Request) ([]byte, int) {
	<-time.After(m.delay)
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
	log.Printf("[mock] Login: username=%s, password=%s\n", username, password)

	return []byte(http.StatusText(http.StatusAccepted)), http.StatusAccepted
}

// SearchTrains は新幹線検索結果を返します
func (m *Mock) SearchTrains(req *http.Request) ([]byte, int) {
	<-time.After(m.delay)
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
	log.Printf("[mock] SearchTrains: use_at=%s, from=%s, to=%s\n", query.Get("use_at"), query.Get("from"), query.Get("to"))

	b, err := json.Marshal(&isutrain.Trains{
		&isutrain.Train{Class: "のぞみ", Name: "96号", Start: 1, Last: 2},
		&isutrain.Train{Class: "こだま", Name: "96号", Start: 3, Last: 4},
	})
	if err != nil {
		return []byte(http.StatusText(http.StatusInternalServerError)), http.StatusInternalServerError
	}

	return b, http.StatusOK
}

// ListTrainSeats は列車の席一覧を返します
func (m *Mock) ListTrainSeats(req *http.Request) ([]byte, int) {
	<-time.After(m.delay)
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
	log.Printf("[mock] ListTrainSeats: trainClass=%s, trainName=%s\n", trainClass, trainName)

	// 適当な席を返す
	b, err := json.Marshal(&isutrain.TrainSeats{
		&isutrain.TrainSeat{},
	})
	if err != nil {
		return []byte(http.StatusText(http.StatusInternalServerError)), http.StatusInternalServerError
	}

	return b, http.StatusOK
}

// Reserve は座席予約を実施し、結果を返します
func (m *Mock) Reserve(req *http.Request) ([]byte, int) {
	<-time.After(m.delay)
	// 予約情報を受け取って、予約できたかを返す

	// FIXME: ユーザID
	// 複数の座席指定で予約するかもしれない
	// なので、予約には複数の座席予約が紐づいている

	log.Println("[mock] Reserve")

	// NOTE: とりあえず、パラメータガン無視でPOSTできるところ先にやる
	return []byte(http.StatusText(http.StatusAccepted)), http.StatusAccepted
}

// CommitReservation は予約を確定します
func (m *Mock) CommitReservation(req *http.Request) ([]byte, int) {
	<-time.After(m.delay)
	// 予約IDを受け取って、確定するだけ
	if err := req.ParseForm(); err != nil {
		return []byte(http.StatusText(http.StatusBadRequest)), http.StatusBadRequest
	}
	log.Println("[mock] CommitReservation")

	_, err := httpmock.GetSubmatchAsUint(req, 1)
	if err != nil {
		return []byte(http.StatusText(http.StatusBadRequest)), http.StatusBadRequest
	}

	return []byte(http.StatusText(http.StatusAccepted)), http.StatusAccepted
}

// CancelReservation は予約をキャンセルします
func (m *Mock) CancelReservation(req *http.Request) ([]byte, int) {
	<-time.After(m.delay)
	// 予約IDを受け取って
	if err := req.ParseForm(); err != nil {
		return []byte(http.StatusText(http.StatusBadRequest)), http.StatusBadRequest
	}
	log.Println("[mock] CancelReservation")

	_, err := httpmock.GetSubmatchAsUint(req, 1)
	if err != nil {
		return []byte(http.StatusText(http.StatusBadRequest)), http.StatusBadRequest
	}

	return []byte(http.StatusText(http.StatusNoContent)), http.StatusNoContent
}

// ListReservations はアカウントにひもづく予約履歴を返します
func (m *Mock) ListReservations(req *http.Request) ([]byte, int) {
	<-time.After(m.delay)
	log.Println("[mock] ListReservations")
	b, err := json.Marshal(isutrain.SeatReservations{
		&isutrain.SeatReservation{ID: 1111, PaymentMethod: string(isutrain.CreditCard), Status: string(isutrain.Pending), ReserveAt: time.Now()},
	})
	if err != nil {
		return []byte(http.StatusText(http.StatusBadRequest)), http.StatusBadRequest
	}

	return b, http.StatusOK
}
