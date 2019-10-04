package isutrain

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"time"

	"github.com/chibiegg/isucon9-final/bench/internal/bencherror"
	"github.com/chibiegg/isucon9-final/bench/internal/config"
	"github.com/chibiegg/isucon9-final/bench/internal/endpoint"
	"github.com/chibiegg/isucon9-final/bench/internal/util"
	"go.uber.org/zap"
)

type Client struct {
	sess    *Session
	baseURL *url.URL

	loginUser *User
}

func NewClient() (*Client, error) {
	sess, err := NewSession()
	if err != nil {
		return nil, bencherror.SystemErrs.AddError(bencherror.NewCriticalError(err, "Isutrainクライアントが作成できません. 運営に確認をお願いいたします"))
	}

	u, err := url.Parse(config.TargetBaseURL)
	if err != nil {
		return nil, bencherror.SystemErrs.AddError(bencherror.NewCriticalError(err, "Isutrainクライアントが作成できません. 運営に確認をお願いいたします"))
	}

	return &Client{
		sess:    sess,
		baseURL: u,
	}, nil
}

func NewClientForInitialize() (*Client, error) {
	sess, err := newSessionForInitialize()
	if err != nil {
		return nil, bencherror.SystemErrs.AddError(bencherror.NewCriticalError(err, "Isutrainクライアントが作成できません. 運営に確認をお願いいたします"))
	}

	u, err := url.Parse(config.TargetBaseURL)
	if err != nil {
		return nil, bencherror.SystemErrs.AddError(bencherror.NewCriticalError(err, "Isutrainクライアントが作成できません. 運営に確認をお願いいたします"))
	}

	return &Client{
		sess:    sess,
		baseURL: u,
	}, nil
}

// ReplaceMockTransport は、clientの利用するhttp.RoundTripperを、DefaultTransportに差し替えます
// NOTE: httpmockはhttp.DefaultTransportを利用するため、モックテストの時この関数を利用する
func (c *Client) ReplaceMockTransport() {
	c.sess.httpClient.Transport = http.DefaultTransport
}

func (c *Client) Initialize(ctx context.Context) {
	var (
		successCode  = http.StatusOK
		u            = *c.baseURL
		endpointPath = endpoint.GetPath(endpoint.Initialize)
	)
	u.Path = filepath.Join(u.Path, endpointPath)

	ctx, cancel := context.WithTimeout(ctx, config.InitializeTimeout)
	defer cancel()

	req, err := c.sess.newRequest(ctx, http.MethodPost, u.String(), nil)
	if err != nil {
		bencherror.InitializeErrs.AddError(bencherror.NewCriticalError(err, "POST %s: リクエストに失敗しました", endpointPath))
		return
	}

	resp, err := c.sess.do(req)
	if err != nil {
		bencherror.InitializeErrs.AddError(bencherror.NewWrapError(err, "POST %s: リクエストに失敗しました", endpointPath))
		return
	}
	defer resp.Body.Close()

	var initializeResp *InitializeResponse
	if resp.StatusCode == successCode {
		if err := json.NewDecoder(resp.Body).Decode(&initializeResp); err != nil {
			bencherror.BenchmarkErrs.AddError(bencherror.NewCriticalError(err, "POST %s: レスポンスの形式が不正です", endpointPath))
			return
		}

		if initializeResp.AvailableDays <= 0 {
			bencherror.InitializeErrs.AddError(bencherror.NewCriticalError(err, "POST %s: 予約可能日数は正の整数値でなければなりません: got=%d", endpointPath, initializeResp.AvailableDays))
			return
		}

		config.Language = initializeResp.Language
		if len(initializeResp.Language) == 0 {
			bencherror.InitializeErrs.AddError(bencherror.NewSimpleCriticalError("POST %s: languageが指定されていません"))
			return
		}

		if err := config.SetAvailReserveDays(initializeResp.AvailableDays); err != nil {
			bencherror.InitializeErrs.AddError(bencherror.NewCriticalError(err, "POST %s: 予約可能日数の設定に失敗しました", endpointPath))
			return
		}
	}

	if err := bencherror.NewHTTPStatusCodeError(req, resp, successCode); err != nil {
		bencherror.InitializeErrs.AddError(bencherror.NewCriticalError(err, "POST %s: ステータスコードが不正です: got=%d, want=%d", endpointPath, resp.StatusCode, http.StatusOK))
		return
	}

	endpoint.IncPathCounter(endpoint.Initialize)
}

func (c *Client) Settings(ctx context.Context) (*SettingsResponse, error) {
	var (
		successCode  = http.StatusOK
		endpointPath = endpoint.GetPath(endpoint.Settings)
		u            = *c.baseURL
	)
	u.Path = filepath.Join(u.Path, endpointPath)

	req, err := c.sess.newRequest(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, bencherror.NewApplicationError(err, "GET %s: 設定情報の取得に失敗しました", endpointPath)
	}

	resp, err := c.sess.do(req)
	if err != nil {
		return nil, bencherror.NewWrapError(err, "GET %s: 設定情報の取得に失敗しました", endpointPath)
	}
	defer resp.Body.Close()

	var settings *SettingsResponse
	if resp.StatusCode == successCode {
		if err := json.NewDecoder(resp.Body).Decode(&settings); err != nil {
			return nil, bencherror.NewApplicationError(err, "GET %s: レスポンスのUnmarshalに失敗しました", endpointPath)
		}
	}

	if err := bencherror.NewHTTPStatusCodeError(req, resp, successCode); err != nil {
		return nil, bencherror.NewApplicationError(err, "GET %s: ステータスコードが不正です: got=%d, want=%d", endpointPath, resp.StatusCode, http.StatusOK)
	}

	return settings, nil
}

func (c *Client) Signup(ctx context.Context, email, password string, opt ...ClientOption) error {
	var (
		successCode  = http.StatusOK
		opts         = newClientOptions(successCode, opt...)
		endpointPath = endpoint.GetPath(endpoint.Signup)
		u            = *c.baseURL
	)
	u.Path = filepath.Join(u.Path, endpointPath)

	b, err := json.Marshal(&User{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return bencherror.NewApplicationError(err, "POST %s: リクエストに失敗しました", endpointPath)
	}

	req, err := c.sess.newRequest(ctx, http.MethodPost, u.String(), bytes.NewBuffer(b))
	if err != nil {
		return bencherror.NewApplicationError(err, "POST %s: リクエストに失敗しました", endpointPath)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.sess.do(req)
	if err != nil {
		return bencherror.NewWrapError(err, "POST %s: リクエストに失敗しました", endpointPath)
	}
	defer resp.Body.Close()

	if err := bencherror.NewHTTPStatusCodeError(req, resp, opts.wantStatusCode); err != nil {
		return bencherror.NewApplicationError(err, "POST %s: ステータスコードが不正です: got=%d, want=%d", endpointPath, resp.StatusCode, opts.wantStatusCode)
	}

	endpoint.IncPathCounter(endpoint.Signup)

	return nil
}

func (c *Client) Login(ctx context.Context, email, password string, opt ...ClientOption) error {
	var (
		successCode  = http.StatusOK
		opts         = newClientOptions(successCode, opt...)
		endpointPath = endpoint.GetPath(endpoint.Login)
		u            = *c.baseURL
	)
	u.Path = filepath.Join(u.Path, endpointPath)

	loginUser := &User{
		Email:    email,
		Password: password,
	}

	b, err := json.Marshal(loginUser)
	if err != nil {
		return bencherror.NewApplicationError(err, "POST %s: リクエストに失敗しました", endpointPath)
	}

	req, err := c.sess.newRequest(ctx, http.MethodPost, u.String(), bytes.NewBuffer(b))
	if err != nil {
		return bencherror.NewApplicationError(err, "POST %s: リクエストに失敗しました", endpointPath)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.sess.do(req)
	if err != nil {
		return bencherror.NewWrapError(err, "POST %s: リクエストに失敗しました", endpointPath)
	}
	defer resp.Body.Close()

	if resp.StatusCode == successCode {
		c.loginUser = loginUser
	}

	if err := bencherror.NewHTTPStatusCodeError(req, resp, opts.wantStatusCode); err != nil {
		return bencherror.NewApplicationError(err, "POST %s: ステータスコードが不正です: got=%d, want=%d", endpointPath, resp.StatusCode, http.StatusOK)
	}

	endpoint.IncPathCounter(endpoint.Login)

	return nil
}

func (c *Client) Logout(ctx context.Context, opt ...ClientOption) error {
	var (
		successCode  = http.StatusOK
		opts         = newClientOptions(successCode, opt...)
		endpointPath = endpoint.GetPath(endpoint.Logout)
		u            = *c.baseURL
	)
	u.Path = filepath.Join(u.Path, endpointPath)
	req, err := c.sess.newRequest(ctx, http.MethodPost, u.String(), nil)
	if err != nil {
		return bencherror.NewApplicationError(err, "POST %s: リクエストに失敗しました", endpointPath)
	}

	resp, err := c.sess.do(req)
	if err != nil {
		return bencherror.NewWrapError(err, "POST %s: リクエストに失敗しました", endpointPath)
	}
	defer resp.Body.Close()

	if err := bencherror.NewHTTPStatusCodeError(req, resp, opts.wantStatusCode); err != nil {
		return bencherror.NewApplicationError(err, "POST %s: ステータスコードが不正です: got=%d, want=%d", endpointPath, resp.StatusCode, http.StatusOK)
	}

	endpoint.IncPathCounter(endpoint.Logout)

	return nil
}

// ListStations は駅一覧列挙APIです
func (c *Client) ListStations(ctx context.Context, opt ...ClientOption) (ListStationsResponse, error) {
	var (
		successCode  = http.StatusOK
		opts         = newClientOptions(successCode, opt...)
		endpointPath = endpoint.GetPath(endpoint.ListStations)
		u            = *c.baseURL
	)
	u.Path = filepath.Join(u.Path, endpointPath)

	req, err := c.sess.newRequest(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return ListStationsResponse{}, bencherror.NewApplicationError(err, "GET %s: リクエストに失敗しました", endpointPath)
	}

	resp, err := c.sess.do(req)
	if err != nil {
		return ListStationsResponse{}, bencherror.NewWrapError(err, "GET %s: リクエストに失敗しました", endpointPath)
	}
	defer resp.Body.Close()

	var listStationsResp ListStationsResponse
	if resp.StatusCode == successCode {
		if err := json.NewDecoder(resp.Body).Decode(&listStationsResp); err != nil {
			return ListStationsResponse{}, bencherror.NewApplicationError(err, "GET %s: レスポンスのUnmarshalに失敗しました", endpointPath)
		}
	}

	if err := bencherror.NewHTTPStatusCodeError(req, resp, opts.wantStatusCode); err != nil {
		return ListStationsResponse{}, bencherror.NewApplicationError(err, "GET %s: ステータスコードが不正です: got=%d, want=%d", endpointPath, resp.StatusCode, opts.wantStatusCode)
	}

	endpoint.IncPathCounter(endpoint.ListStations)

	return listStationsResp, nil
}

// SearchTrains は 列車検索APIです
func (c *Client) SearchTrains(ctx context.Context, useAt time.Time, from, to, trainClass string, adult, child int, opt ...ClientOption) (SearchTrainsResponse, error) {
	var (
		successCode  = http.StatusOK
		opts         = newClientOptions(successCode, opt...)
		endpointPath = endpoint.GetPath(endpoint.SearchTrains)
		u            = *c.baseURL
	)
	u.Path = filepath.Join(u.Path, endpointPath)

	req, err := c.sess.newRequest(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return SearchTrainsResponse{}, bencherror.NewApplicationError(err, "GET %s: 列車検索リクエストに失敗しました", endpointPath)
	}

	query := req.URL.Query()
	query.Set("use_at", util.FormatISO8601(useAt))
	query.Set("train_class", trainClass)
	query.Set("from", from)
	query.Set("to", to)
	query.Set("adult", strconv.Itoa(adult))
	query.Set("child", strconv.Itoa(child))
	req.URL.RawQuery = query.Encode()

	resp, err := c.sess.do(req)
	if err != nil {
		return SearchTrainsResponse{}, bencherror.NewWrapError(err, "GET %s: 列車検索リクエストに失敗しました", endpointPath)
	}
	defer resp.Body.Close()

	var searchTrainsResp SearchTrainsResponse
	if resp.StatusCode == successCode {
		if err := json.NewDecoder(resp.Body).Decode(&searchTrainsResp); err != nil {
			return SearchTrainsResponse{}, bencherror.NewApplicationError(err, "GET %s: レスポンスのUnmarshalに失敗しました", endpointPath)
		}
	}

	if opts.autoAssert && resp.StatusCode == successCode {
		if err := assertSearchTrains(ctx, endpointPath, searchTrainsResp); err != nil {
			return nil, err
		}
	}

	if err := bencherror.NewHTTPStatusCodeError(req, resp, opts.wantStatusCode); err != nil {
		return SearchTrainsResponse{}, bencherror.NewApplicationError(err, "GET %s: ステータスコードが不正です: got=%d, want=%d", endpointPath, resp.StatusCode, opts.wantStatusCode)
	}

	endpoint.IncPathCounter(endpoint.SearchTrains)

	return searchTrainsResp, nil
}

func (c *Client) SearchTrainSeats(ctx context.Context, date time.Time, trainClass, trainName string, carNum int, departure, arrival string, opt ...ClientOption) (*SearchTrainSeatsResponse, error) {
	var (
		successCode  = http.StatusOK
		opts         = newClientOptions(successCode, opt...)
		endpointPath = endpoint.GetPath(endpoint.ListTrainSeats)
		u            = *c.baseURL
	)
	u.Path = filepath.Join(u.Path, endpointPath)

	lgr := zap.S()

	req, err := c.sess.newRequest(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		lgr.Warnf("座席列挙 リクエスト作成に失敗: %+v", err)
		return nil, bencherror.NewApplicationError(err, "GET %s: リクエストに失敗しました", endpointPath)
	}

	query := req.URL.Query()
	query.Set("date", util.FormatISO8601(date))
	query.Set("train_class", trainClass)
	query.Set("train_name", trainName)
	query.Set("car_number", strconv.Itoa(carNum))
	query.Set("from", departure)
	query.Set("to", arrival)
	req.URL.RawQuery = query.Encode()

	lgr.Infow("座席列挙",
		"date", util.FormatISO8601(date),
		"train_class", trainClass,
		"train_name", trainName,
		"car_number", strconv.Itoa(carNum),
		"from", departure,
		"to", arrival,
	)

	resp, err := c.sess.do(req)
	if err != nil {
		lgr.Warnf("座席列挙リクエスト失敗: %+v", err)
		return nil, bencherror.NewWrapError(err, "GET %s: リクエストに失敗しました", endpointPath)
	}
	defer resp.Body.Close()

	var searchTrainSeatsResp *SearchTrainSeatsResponse
	if resp.StatusCode == successCode {
		if err := json.NewDecoder(resp.Body).Decode(&searchTrainSeatsResp); err != nil {
			lgr.Warnf("座席列挙Unmarshal失敗: %+v", err)
			return nil, bencherror.NewApplicationError(err, "GET %s: レスポンスのUnmarshalに失敗しました", endpointPath)
		}
	}

	// NotFound、あるいはBadRequestの場合、座席を得ることはできない
	if opts.autoAssert && resp.StatusCode == successCode {
		if err := assertSearchTrainSeats(ctx, endpointPath, searchTrainSeatsResp); err != nil {
			return nil, err
		}
	}

	if err := bencherror.NewHTTPStatusCodeError(req, resp, opts.wantStatusCode); err != nil {
		lgr.Warnf("座席列挙 ステータスコードが不正: %+v", err)
		return nil, bencherror.NewApplicationError(err, "GET %s: ステータスコードが不正です: got=%d, want=%d", endpointPath, resp.StatusCode, opts.wantStatusCode)
	}

	endpoint.IncPathCounter(endpoint.ListTrainSeats)

	return searchTrainSeatsResp, nil
}

func (c *Client) Reserve(
	ctx context.Context,
	trainClass, trainName string,
	seatClass string,
	seats TrainSeats,
	departure, arrival string,
	useAt time.Time,
	carNum int,
	child, adult int,
	opt ...ClientOption,
) (*ReserveResponse, error) {
	var (
		successCode  = http.StatusOK
		opts         = newClientOptions(successCode, opt...)
		endpointPath = endpoint.GetPath(endpoint.Reserve)
		u            = *c.baseURL
	)
	u.Path = filepath.Join(u.Path, endpointPath)

	lgr := zap.S()

	reserveReq := &ReserveRequest{
		Date:          util.FormatISO8601(useAt),
		TrainName:     trainName,
		TrainClass:    trainClass,
		CarNum:        carNum,
		IsSmokingSeat: false, // FIXME: 喫煙席を選べるように
		SeatClass:     seatClass,
		Departure:     departure,
		Arrival:       arrival,
		Child:         child,
		Adult:         adult,
		Column:        "", // FIXME: カラムを選べるように
		Seats:         seats,
	}

	b, err := json.Marshal(reserveReq)
	if err != nil {
		return nil, err
	}

	lgr.Infof("予約クエリ: %s", string(b))

	req, err := c.sess.newRequest(ctx, http.MethodPost, u.String(), bytes.NewBuffer(b))
	if err != nil {
		lgr.Warnf("予約リクエスト失敗: %+v", err)
		return nil, bencherror.NewApplicationError(err, "POST %s: リクエストに失敗しました", endpointPath)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.sess.do(req)
	if err != nil {
		lgr.Warnf("予約リクエスト失敗: %+v", err)
		return nil, bencherror.NewWrapError(err, "POST %s: リクエストに失敗しました", endpointPath)
	}
	defer resp.Body.Close()

	var reserveResp *ReserveResponse
	if resp.StatusCode == successCode {
		if err := json.NewDecoder(resp.Body).Decode(&reserveResp); err != nil {
			lgr.Warnf("予約リクエストのUnmarshal失敗: %+v", err)
			return nil, bencherror.NewApplicationError(err, "POST %s: JSONのUnmarshalに失敗しました", endpointPath)
		}
	}

	if opts.autoAssert && resp.StatusCode == successCode {
		if err := assertCanReserve(ctx, endpointPath, reserveReq, reserveResp); err != nil {
			return nil, err
		}
	}

	if resp.StatusCode == successCode {
		ReservationCache.Add(c.loginUser, reserveReq, reserveResp.ReservationID)
	}
	if opts.autoAssert && resp.StatusCode == successCode {
		if err := assertReserve(ctx, endpointPath, c, reserveReq, reserveResp); err != nil {
			return nil, err
		}
	}

	if err := bencherror.NewHTTPStatusCodeError(req, resp, opts.wantStatusCode); err != nil {
		return nil, bencherror.NewApplicationError(err, "POST %s: ステータスコードが不正です: got=%d, want=%d", endpointPath, resp.StatusCode, opts.wantStatusCode)
	}

	// FIXME: webappが座席を返してこないので、ボーナス付与は一旦保留
	// if len(reserveReq.Seats) == 0 {
	// 	// リクエストに席を指定しない曖昧予約の場合、予約できた座席で隣り合う数が多いほど加点される
	// 	var (
	// 		weight     = float64(endpoint.GetWeight(endpoint.ListTrainSeats))
	// 		multiplier = reserveResp.Seats.GetNeighborSeatsMultiplier()
	// 	)
	// 	endpoint.AddExtraScore(endpoint.Reserve, int64(math.Round(weight*multiplier)))
	// } else {
	endpoint.IncPathCounter(endpoint.Reserve)
	// }

	if SeatAvailability(seatClass) != SaNonReserved {
		endpoint.AddExtraScore(endpoint.Reserve, config.ReservedSeatExtraScore)
	}

	return reserveResp, nil
}

func (c *Client) CommitReservation(ctx context.Context, reservationID int, cardToken string, opt ...ClientOption) error {
	var (
		successCode  = http.StatusOK
		opts         = newClientOptions(successCode, opt...)
		endpointPath = endpoint.GetPath(endpoint.CommitReservation)
		u            = *c.baseURL
	)
	u.Path = filepath.Join(u.Path, endpointPath)

	lgr := zap.S()

	// FIXME: 一応構造体にする？
	lgr.Infow("予約確定処理",
		"reservation_id", reservationID,
		"card_token", cardToken,
	)

	b, err := json.Marshal(&CommitReservationRequest{
		ReservationID: reservationID,
		CardToken:     cardToken,
	})
	if err != nil {
		return bencherror.NewApplicationError(err, "POST %s: Marshalに失敗しました", endpointPath)
	}

	req, err := c.sess.newRequest(ctx, http.MethodPost, u.String(), bytes.NewBuffer(b))
	if err != nil {
		return bencherror.NewApplicationError(err, "POST %s: リクエストの作成に失敗しました", endpointPath)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.sess.do(req)
	if err != nil {
		return bencherror.NewWrapError(err, "POST %s: リクエストに失敗しました", endpointPath)
	}
	defer resp.Body.Close()

	var commitReservationResp *CommitReservationResponse
	if resp.StatusCode == successCode {
		if err := json.NewDecoder(resp.Body).Decode(&commitReservationResp); err != nil {
			return bencherror.NewApplicationError(err, "POST %s: JSONのUnmarshalに失敗しました", endpointPath)
		}
	}

	if opts.autoAssert && resp.StatusCode == successCode {
		if err := assertCommitReservation(ctx, endpointPath, commitReservationResp); err != nil {
			return err
		}
	}

	if resp.StatusCode == successCode {
		if err := ReservationCache.Commit(reservationID); err != nil {
			bencherror.SystemErrs.AddError(bencherror.NewCriticalError(err, "POST %s: 存在しない予約へのCommitを行おうとしました", endpointPath))
		}
	}

	if err := bencherror.NewHTTPStatusCodeError(req, resp, opts.wantStatusCode); err != nil {
		return bencherror.NewApplicationError(err, "POST %s: ステータスコードが不正です: got=%d, want=%d", endpointPath, resp.StatusCode, opts.wantStatusCode)
	}

	endpoint.IncPathCounter(endpoint.CommitReservation)

	return nil
}

func (c *Client) ListReservations(ctx context.Context, opt ...ClientOption) (ListReservationsResponse, error) {
	var (
		successCode  = http.StatusOK
		opts         = newClientOptions(successCode, opt...)
		endpointPath = endpoint.GetPath(endpoint.ListReservations)
		u            = *c.baseURL
	)
	u.Path = filepath.Join(u.Path, endpointPath)

	req, err := c.sess.newRequest(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return ListReservationsResponse{}, bencherror.NewApplicationError(err, "GET %s: リクエストに失敗しました", endpointPath)
	}

	resp, err := c.sess.do(req)
	if err != nil {
		return ListReservationsResponse{}, bencherror.NewWrapError(err, "GET %s: リクエストに失敗しました", endpointPath)
	}
	defer resp.Body.Close()

	var listReservationResp ListReservationsResponse
	if resp.StatusCode == successCode {
		if err := json.NewDecoder(resp.Body).Decode(&listReservationResp); err != nil {
			return ListReservationsResponse{}, bencherror.NewApplicationError(err, "GET %s: 予約のMarshalに失敗しました", endpointPath)
		}
	}

	if err := bencherror.NewHTTPStatusCodeError(req, resp, opts.wantStatusCode); err != nil {
		return ListReservationsResponse{}, bencherror.NewApplicationError(err, "GET %s: ステータスコードが不正です: got=%d, want=%d", endpointPath, resp.StatusCode, opts.wantStatusCode)
	}

	endpoint.IncPathCounter(endpoint.ListReservations)

	return listReservationResp, nil
}

func (c *Client) ShowReservation(ctx context.Context, reservationID int, opt ...ClientOption) (ShowReservationResponse, error) {
	var (
		successCode  = http.StatusOK
		opts         = newClientOptions(successCode, opt...)
		endpointPath = endpoint.GetDynamicPath(endpoint.ShowReservation, reservationID)
		u            = *c.baseURL
	)
	u.Path = filepath.Join(u.Path, endpointPath)

	req, err := c.sess.newRequest(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, bencherror.NewApplicationError(err, "GET %s: リクエストに失敗しました", endpointPath)
	}

	resp, err := c.sess.do(req)
	if err != nil {
		return nil, bencherror.NewWrapError(err, "GET %s: リクエストに失敗しました", endpointPath)
	}

	var showReservationResp ShowReservationResponse
	if resp.StatusCode == successCode {
		if err := json.NewDecoder(resp.Body).Decode(&showReservationResp); err != nil {
			return nil, bencherror.NewApplicationError(err, "GET %s: Unmarshalに失敗しました", endpointPath)
		}
	}

	if err := bencherror.NewHTTPStatusCodeError(req, resp, opts.wantStatusCode); err != nil {
		return nil, bencherror.NewApplicationError(err, "GET %s: ステータスコードが不正です: got=%d, want=%d", endpointPath, resp.StatusCode, opts.wantStatusCode)
	}

	endpoint.IncDynamicPathCounter(endpoint.ShowReservation)

	return showReservationResp, nil
}

func (c *Client) CancelReservation(ctx context.Context, reservationID int, opt ...ClientOption) error {
	var (
		successCode  = http.StatusOK
		opts         = newClientOptions(successCode, opt...)
		endpointPath = endpoint.GetDynamicPath(endpoint.CancelReservation, reservationID)
		u            = *c.baseURL
	)
	u.Path = filepath.Join(u.Path, endpointPath)

	lgr := zap.S()

	lgr.Infow("予約キャンセル処理",
		"reservation_id", reservationID,
	)

	req, err := c.sess.newRequest(ctx, http.MethodPost, u.String(), nil)
	if err != nil {
		return bencherror.NewApplicationError(err, "POST %s: リクエストに失敗しました", endpointPath)
	}

	resp, err := c.sess.do(req)
	if err != nil {
		return bencherror.NewWrapError(err, "POST %s: リクエストに失敗しました", endpointPath)
	}
	defer resp.Body.Close()

	var cancelReservationResponse *CancelReservationResponse
	if resp.StatusCode == successCode {
		if err := json.NewDecoder(resp.Body).Decode(&cancelReservationResponse); err != nil {
			return bencherror.NewApplicationError(err, "POST %s: JSONのUnmarshalに失敗しました", endpointPath)
		}
	}

	if opts.autoAssert && resp.StatusCode == successCode {
		if err := assertCancelReservation(ctx, endpointPath, c, reservationID, cancelReservationResponse); err != nil {
			return err
		}
	}

	if resp.StatusCode == successCode {
		if err := ReservationCache.Cancel(reservationID); err != nil {
			// FIXME: こういうベンチマーカーの異常は、利用者向けには一般的なメッセージで運営に連絡して欲しいと書き、運営向けにSlackに通知する
			bencherror.SystemErrs.AddError(bencherror.NewCriticalError(err, "存在しない予約のCancelを実施しようとしました"))
		}
	}

	if err := bencherror.NewHTTPStatusCodeError(req, resp, opts.wantStatusCode); err != nil {
		return bencherror.NewApplicationError(err, "POST %s: ステータスコードが不正です: got=%d, want=%d", endpointPath, resp.StatusCode, opts.wantStatusCode)
	}

	endpoint.IncDynamicPathCounter(endpoint.CancelReservation)

	return nil
}

func (c *Client) DownloadAsset(ctx context.Context, path string) ([]byte, error) {
	var (
		successCode = http.StatusOK
		u           = *c.baseURL
	)
	u.Path = filepath.Join(u.Path, path)

	req, err := c.sess.newRequest(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return []byte{}, bencherror.PreTestErrs.AddError(bencherror.NewCriticalError(err, "GET %s: 静的ファイルのダウンロードに失敗しました", path))
	}

	resp, err := c.sess.do(req)
	if err != nil {
		return []byte{}, bencherror.PreTestErrs.AddError(bencherror.NewWrapError(err, "GET %s: 静的ファイルのダウンロードに失敗しました", path))
	}
	defer resp.Body.Close()

	if err := bencherror.NewHTTPStatusCodeError(req, resp, successCode); err != nil {
		return []byte{}, bencherror.PreTestErrs.AddError(bencherror.NewCriticalError(err, "GET %s: ステータスコードが不正です: got=%d, want=%d", path, resp.StatusCode, http.StatusOK))
	}

	return ioutil.ReadAll(resp.Body)
}
