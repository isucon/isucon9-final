package mock

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"net/http/httptest"

	"github.com/chibiegg/isucon9-final/bench/internal/util"
	"github.com/chibiegg/isucon9-final/bench/isutrain"
	"github.com/gorilla/sessions"
	"github.com/jarcoal/httpmock"
)

// Mock は `isutrain` のモック実装です
type Mock struct {
	LoginDelay             time.Duration
	ListStationsDelay      time.Duration
	SearchTrainsDelay      time.Duration
	ListTrainSeatsDelay    time.Duration
	ReserveDelay           time.Duration
	CommitReservationDelay time.Duration
	CancelReservationDelay time.Duration
	ListReservationDelay   time.Duration

	sessionName string
	session     *sessions.CookieStore

	injectFunc func(path string) error

	paymentMock *paymentMock
}

func NewMock(paymentMock *paymentMock) (*Mock, error) {
	randomStr, err := util.SecureRandomStr(20)
	if err != nil {
		return nil, err
	}
	return &Mock{
		injectFunc: func(path string) error {
			return nil
		},
		paymentMock: paymentMock,
		sessionName: "session_isutrain",
		session:     sessions.NewCookieStore([]byte(randomStr)),
	}, nil
}

func (m *Mock) getSession(req *http.Request) (*sessions.Session, error) {
	session, err := m.session.Get(req, m.sessionName)
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (m *Mock) Inject(f func(path string) error) {
	m.injectFunc = f
}

func (m *Mock) Initialize(req *http.Request) ([]byte, int) {
	if err := m.injectFunc(req.URL.Path); err != nil {
		return []byte(http.StatusText(http.StatusInternalServerError)), http.StatusInternalServerError
	}

	b, err := json.Marshal(&isutrain.InitializeResponse{
		AvailableDays: 30,
		Language:      "golang",
	})
	if err != nil {
		return []byte(http.StatusText(http.StatusInternalServerError)), http.StatusInternalServerError
	}

	return b, http.StatusOK
}

// Signup はユーザ登録を行います
func (m *Mock) Signup(req *http.Request) ([]byte, int) {
	user := &isutrain.User{}
	if err := json.NewDecoder(req.Body).Decode(&user); err != nil {
		return []byte(http.StatusText(http.StatusBadRequest)), http.StatusBadRequest
	}

	if len(user.Email) == 0 || len(user.Password) == 0 {
		return []byte(http.StatusText(http.StatusBadRequest)), http.StatusBadRequest
	}

	return []byte(http.StatusText(http.StatusOK)), http.StatusOK
}

// Login はログイン処理結果を返します
func (m *Mock) Login(req *http.Request) (*httptest.ResponseRecorder, int) {
	<-time.After(m.LoginDelay)

	wr := httptest.NewRecorder()

	user := &isutrain.User{}
	if err := json.NewDecoder(req.Body).Decode(&user); err != nil {
		wr.Write([]byte(http.StatusText(http.StatusBadRequest)))
		return wr, http.StatusBadRequest
	}

	if len(user.Email) == 0 || len(user.Password) == 0 {
		wr.Write([]byte(http.StatusText(http.StatusBadRequest)))
		return wr, http.StatusBadRequest
	}

	session, err := m.getSession(req)
	if err != nil {
		wr.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return wr, http.StatusInternalServerError
	}

	csrfToken, err := util.SecureRandomStr(20)
	if err != nil {
		wr.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return wr, http.StatusInternalServerError
	}

	if err := session.Save(req, wr); err != nil {
		wr.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return wr, http.StatusInternalServerError
	}
	http.SetCookie(wr, &http.Cookie{
		Name:  "csrf_token",
		Value: csrfToken,
		Path:  "/",
	})
	http.SetCookie(wr, &http.Cookie{
		Name:  "mock",
		Value: "true",
		Path:  "/",
	})

	wr.Write([]byte(http.StatusText(http.StatusOK)))
	return wr, http.StatusOK
}

func (m *Mock) Logout(req *http.Request) (*httptest.ResponseRecorder, int) {
	// <-time.After()

	wr := httptest.NewRecorder()

	session, err := m.getSession(req)
	if err != nil {
		wr.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return wr, http.StatusInternalServerError
	}
	session.Values[m.sessionName] = ""
	session.Values["csrf_token"] = ""
	session.Values["mock"] = ""

	if err := session.Save(req, wr); err != nil {
		wr.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return wr, http.StatusInternalServerError
	}

	return wr, http.StatusOK
}

func (m *Mock) ListStations(req *http.Request) ([]byte, int) {
	<-time.After(m.ListStationsDelay)
	b, err := json.Marshal(isutrain.ListStationsResponse{
		&isutrain.Station{ID: 1, Name: "東京", Distance: 10.5, IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: false},
	})
	if err != nil {
		return []byte(http.StatusText(http.StatusInternalServerError)), http.StatusInternalServerError
	}

	return b, http.StatusOK
}

// SearchTrains は新幹線検索結果を返します
func (m *Mock) SearchTrains(req *http.Request) ([]byte, int) {
	<-time.After(m.SearchTrainsDelay)
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

	var (
		seatAvailability = map[string]string{
			string(isutrain.SaPremium):       "○",
			string(isutrain.SaPremiumSmoke):  "×",
			string(isutrain.SaReserved):      "△",
			string(isutrain.SaReservedSmoke): "○",
			string(isutrain.SaNonReserved):   "○",
		}
		fareInformation = map[string]int{
			string(isutrain.SaPremium):       24000,
			string(isutrain.SaPremiumSmoke):  24500,
			string(isutrain.SaReserved):      19000,
			string(isutrain.SaReservedSmoke): 19500,
			string(isutrain.SaNonReserved):   15000,
		}
	)
	b, err := json.Marshal(isutrain.SearchTrainsResponse{
		&isutrain.Train{
			Class:            "最速",
			Name:             "96号",
			Start:            "1",
			Last:             "2",
			Departure:        "東京",
			Arrival:          "名古屋",
			DepartedAt:       util.FormatISO8601(time.Now()),
			ArrivedAt:        util.FormatISO8601(time.Now()),
			SeatAvailability: seatAvailability,
			FareInformation:  fareInformation,
		},
		&isutrain.Train{
			Class:            "最速",
			Name:             "96号",
			Start:            "3",
			Last:             "4",
			Departure:        "名古屋",
			Arrival:          "大阪",
			DepartedAt:       util.FormatISO8601(time.Now()),
			ArrivedAt:        util.FormatISO8601(time.Now()),
			SeatAvailability: seatAvailability,
			FareInformation:  fareInformation,
		},
	})
	if err != nil {
		return []byte(http.StatusText(http.StatusInternalServerError)), http.StatusInternalServerError
	}

	return b, http.StatusOK
}

// SearchTrainSeats は列車の席一覧を返します
func (m *Mock) SearchTrainSeats(req *http.Request) ([]byte, int) {
	<-time.After(m.ListTrainSeatsDelay)
	q := req.URL.Query()
	if q == nil {
		return []byte(http.StatusText(http.StatusBadRequest)), http.StatusBadRequest
	}

	// 列車特定情報を受け取る
	jst := time.FixedZone("Asia/Tokyo", 9*60*60)
	var (
		trainClass              = q.Get("train_class")
		trainName               = q.Get("train_name")
		carNumber, carNumberErr = strconv.Atoi(q.Get("car_number"))
		date, dateErr           = time.Parse(time.RFC3339, q.Get("date"))
		fromName                = q.Get("from")
		toName                  = q.Get("to")
	)
	if len(trainClass) == 0 || len(trainName) == 0 {
		return []byte(http.StatusText(http.StatusBadRequest)), http.StatusBadRequest
	}
	if len(fromName) == 0 || len(toName) == 0 {
		return []byte(http.StatusText(http.StatusBadRequest)), http.StatusBadRequest
	}
	if carNumberErr != nil || carNumber == 0 {
		return []byte(http.StatusText(http.StatusBadRequest)), http.StatusBadRequest
	}
	if dateErr != nil {
		return []byte(http.StatusText(http.StatusBadRequest)), http.StatusBadRequest
	}
	date = date.In(jst)

	// 適当な席を返す
	b, err := json.Marshal(&isutrain.SearchTrainSeatsResponse{
		Date:       "2006/01/02",
		TrainClass: "最速",
		TrainName:  "dummy",
		CarNumber:  1,
		Seats: isutrain.TrainSeats{
			&isutrain.TrainSeat{
				Row:           1,
				Column:        "A",
				Class:         "premium",
				IsSmokingSeat: true,
				IsOccupied:    false,
			},
		},
		Cars: isutrain.TrainCars{
			&isutrain.TrainCar{
				CarNumber: 1,
				SeatClass: "premium",
			},
		},
	})
	if err != nil {
		return []byte(http.StatusText(http.StatusInternalServerError)), http.StatusInternalServerError
	}

	return b, http.StatusOK
}

// Reserve は座席予約を実施し、結果を返します
func (m *Mock) Reserve(req *http.Request) ([]byte, int) {
	<-time.After(m.ReserveDelay)

	// 予約情報を受け取って、予約できたかを返す
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return []byte(http.StatusText(http.StatusBadRequest)), http.StatusBadRequest
	}

	// 複数の座席指定で予約するかもしれない
	// なので、予約には複数の座席予約が紐づいている
	var reservationReq *isutrain.ReserveRequest
	if err := json.Unmarshal(b, &reservationReq); err != nil {
		return []byte(http.StatusText(http.StatusBadRequest)), http.StatusBadRequest
	}

	if len(reservationReq.TrainClass) == 0 || len(reservationReq.TrainName) == 0 {
		return []byte(http.StatusText(http.StatusBadRequest)), http.StatusBadRequest
	}

	b, err = json.Marshal(&isutrain.ReserveResponse{
		ReservationID: 1111,
		IsOk:          true,
	})
	if err != nil {
		return []byte(http.StatusText(http.StatusInternalServerError)), http.StatusInternalServerError
	}

	// NOTE: とりあえず、パラメータガン無視でPOSTできるところ先にやる
	return b, http.StatusOK
}

// CommitReservation は予約を確定します
func (m *Mock) CommitReservation(req *http.Request) ([]byte, int) {
	<-time.After(m.CommitReservationDelay)
	// 予約IDを受け取って、確定するだけ

	var reservation *isutrain.CommitReservationRequest
	if err := json.NewDecoder(req.Body).Decode(&reservation); err != nil {
		return []byte(http.StatusText(http.StatusBadRequest)), http.StatusBadRequest
	}

	// FIXME: ちゃんとした決済情報を追加する
	m.paymentMock.addPaymentInformation()

	return []byte(http.StatusText(http.StatusOK)), http.StatusOK
}

// CancelReservation は予約をキャンセルします
func (m *Mock) CancelReservation(req *http.Request) ([]byte, int) {
	<-time.After(m.CancelReservationDelay)
	// 予約IDを受け取って

	_, err := httpmock.GetSubmatchAsUint(req, 1)
	if err != nil {
		return []byte(http.StatusText(http.StatusBadRequest)), http.StatusBadRequest
	}

	return []byte(http.StatusText(http.StatusNoContent)), http.StatusNoContent
}

// ListReservations はアカウントにひもづく予約履歴を返します
func (m *Mock) ListReservations(req *http.Request) ([]byte, int) {
	<-time.After(m.ListReservationDelay)
	b, err := json.Marshal([]*isutrain.ReservationResponse{
		&isutrain.ReservationResponse{ReservationID: 1111, Amount: 20250},
	})
	if err != nil {
		return []byte(http.StatusText(http.StatusBadRequest)), http.StatusBadRequest
	}

	return b, http.StatusOK
}

func (m *Mock) ShowReservation(req *http.Request) ([]byte, int) {

	reservationID, err := httpmock.GetSubmatchAsInt(req, 1)
	if err != nil {
		return []byte(http.StatusText(http.StatusBadRequest)), http.StatusBadRequest
	}

	b, err := json.Marshal(&isutrain.ShowReservationResponse{
		ReservationID: int(reservationID),
		Amount:        20250,
	})

	return b, http.StatusOK
}
