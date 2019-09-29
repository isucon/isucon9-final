package isutrain

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"time"

	"github.com/chibiegg/isucon9-final/bench/internal/bencherror"
	"github.com/chibiegg/isucon9-final/bench/internal/config"
	"github.com/chibiegg/isucon9-final/bench/internal/endpoint"
	"github.com/chibiegg/isucon9-final/bench/internal/util"
	"github.com/morikuni/failure"
	"go.uber.org/zap"
)

var (
	ErrTrainSeatsNotFound = errors.New("列車の座席が見つかりませんでした")
)

// type ClientOption struct {
// 	WantStatusCode int
// 	AutoAssert     bool
// }

type Client struct {
	sess    *Session
	baseURL *url.URL

	loginUser *User
}

func NewClient() (*Client, error) {
	sess, err := NewSession()
	if err != nil {
		return nil, bencherror.NewCriticalError(err, "Isutrainクライアントが作成できません. 運営に確認をお願いいたします")
	}

	u, err := url.Parse(config.TargetBaseURL)
	if err != nil {
		return nil, bencherror.NewCriticalError(err, "Isutrainクライアントが作成できません. 運営に確認をお願いいたします")
	}

	return &Client{
		sess:    sess,
		baseURL: u,
	}, nil
}

func NewClientForInitialize() (*Client, error) {
	sess, err := newSessionForInitialize()
	if err != nil {
		return nil, bencherror.NewCriticalError(err, "Isutrainクライアントが作成できません. 運営に確認をお願いいたします")
	}

	u, err := url.Parse(config.TargetBaseURL)
	if err != nil {
		return nil, bencherror.NewCriticalError(err, "Isutrainクライアントが作成できません. 運営に確認をお願いいたします")
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
	u := *c.baseURL
	endpointPath := endpoint.GetPath(endpoint.Initialize)
	u.Path = filepath.Join(u.Path, endpointPath)

	ctx, cancel := context.WithTimeout(ctx, config.InitializeTimeout)
	defer cancel()

	req, err := c.sess.newRequest(ctx, http.MethodPost, u.String(), nil)
	if err != nil {
		bencherror.InitializeErrs.AddError(bencherror.NewApplicationError(err, "POST %s: リクエストに失敗しました", endpointPath))
		return
	}

	// TODO: 言語を返すようにしたり、キャンペーンを返すようにする場合、ちゃんと設定されていなかったらFAILにする
	resp, err := c.sess.do(req)
	if err != nil {
		bencherror.InitializeErrs.AddError(bencherror.NewApplicationError(err, "POST %s: リクエストに失敗しました", endpointPath))
		return
	}
	defer resp.Body.Close()

	if err := bencherror.NewHTTPStatusCodeError(req, resp, http.StatusOK); err != nil {
		bencherror.InitializeErrs.AddError(failure.Wrap(err, failure.Messagef("POST %s: ステータスコードが不正です: got=%d, want=%d", endpointPath, resp.StatusCode, http.StatusOK)))
		return
	}

	var initializeResp *InitializeResponse
	if err := json.NewDecoder(resp.Body).Decode(&initializeResp); err != nil {
		bencherror.BenchmarkErrs.AddError(failure.Wrap(err, failure.Messagef("POST %s: レスポンスの形式が不正です", endpointPath)))
		return
	}

	if initializeResp.AvailableDays <= 0 {
		bencherror.InitializeErrs.AddError(bencherror.NewCriticalError(err, "POST %s: 予約可能日数は正の整数値でなければなりません: got=%d", endpointPath, initializeResp.AvailableDays))
	}

	if err := config.SetAvailReserveDays(initializeResp.AvailableDays); err != nil {
		bencherror.InitializeErrs.AddError(bencherror.NewCriticalError(err, "POST %s: 予約可能日数の設定に失敗しました", endpointPath))
		return
	}

	endpoint.IncPathCounter(endpoint.Initialize)
}

func (c *Client) Settings(ctx context.Context) (*SettingsResponse, error) {
	u := *c.baseURL
	endpointPath := endpoint.GetPath(endpoint.Settings)
	u.Path = filepath.Join(u.Path, endpointPath)

	req, err := c.sess.newRequest(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, failure.Wrap(err, failure.Messagef("GET %s: 設定情報の取得に失敗しました", endpointPath))
	}

	resp, err := c.sess.do(req)
	if err != nil {
		return nil, failure.Wrap(err, failure.Messagef("GET %s: 設定情報の取得に失敗しました", endpointPath))
	}
	defer resp.Body.Close()

	if err := bencherror.NewHTTPStatusCodeError(req, resp, http.StatusOK); err != nil {
		return nil, failure.Wrap(err, failure.Messagef("GET %s: ステータスコードが不正です: got=%d, want=%d", endpointPath, resp.StatusCode, http.StatusOK))
	}
	// TODO: opts.WantStatusCodes制御

	var settings *SettingsResponse
	if err := json.NewDecoder(resp.Body).Decode(&settings); err != nil {
		return nil, failure.Wrap(err, failure.Messagef("GET %s: レスポンスのUnmarshalに失敗しました", endpointPath))
	}

	return settings, nil
}

func (c *Client) Signup(ctx context.Context, email, password string, opt ...ClientOption) error {
	log.Printf("signup opt = %+v\n", opt)
	opts := newClientOptions(http.StatusOK, opt...)
	u := *c.baseURL
	endpointPath := endpoint.GetPath(endpoint.Signup)
	u.Path = filepath.Join(u.Path, endpointPath)

	b, err := json.Marshal(&User{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return failure.Wrap(err, failure.Messagef("POST %s: リクエストに失敗しました", endpointPath))
	}

	req, err := c.sess.newRequest(ctx, http.MethodPost, u.String(), bytes.NewBuffer(b))
	if err != nil {
		return failure.Wrap(err, failure.Messagef("POST %s: リクエストに失敗しました", endpointPath))
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.sess.do(req)
	if err != nil {
		return failure.Wrap(err, failure.Messagef("POST %s: リクエストに失敗しました", endpointPath))
	}
	defer resp.Body.Close()

	if err := bencherror.NewHTTPStatusCodeError(req, resp, opts.wantStatusCode); err != nil {
		return failure.Wrap(err, failure.Messagef("POST %s: ステータスコードが不正です: got=%d, want=%d", endpointPath, resp.StatusCode, opts.wantStatusCode))
	}

	endpoint.IncPathCounter(endpoint.Signup)

	return nil
}

func (c *Client) Login(ctx context.Context, email, password string, opt ...ClientOption) error {
	opts := newClientOptions(http.StatusOK, opt...)
	u := *c.baseURL
	endpointPath := endpoint.GetPath(endpoint.Login)
	u.Path = filepath.Join(u.Path, endpointPath)

	loginUser := &User{
		Email:    email,
		Password: password,
	}

	b, err := json.Marshal(loginUser)
	if err != nil {
		return failure.Wrap(err, failure.Messagef("POST %s: リクエストに失敗しました", endpointPath))
	}

	req, err := c.sess.newRequest(ctx, http.MethodPost, u.String(), bytes.NewBuffer(b))
	if err != nil {
		return failure.Wrap(err, failure.Messagef("POST %s: リクエストに失敗しました", endpointPath))
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.sess.do(req)
	if err != nil {
		return failure.Wrap(err, failure.Messagef("POST %s: リクエストに失敗しました", endpointPath))
	}
	defer resp.Body.Close()

	if err := bencherror.NewHTTPStatusCodeError(req, resp, opts.wantStatusCode); err != nil {
		return failure.Wrap(err, failure.Messagef("POST %s: ステータスコードが不正です: got=%d, want=%d", endpointPath, resp.StatusCode, http.StatusOK))
	}

	c.loginUser = loginUser

	endpoint.IncPathCounter(endpoint.Login)

	return nil
}

func (c *Client) Logout(ctx context.Context, opt ...ClientOption) error {
	opts := newClientOptions(http.StatusOK, opt...)
	u := *c.baseURL
	endpointPath := endpoint.GetPath(endpoint.Logout)
	u.Path = filepath.Join(u.Path, endpointPath)
	req, err := c.sess.newRequest(ctx, http.MethodPost, u.String(), nil)
	if err != nil {
		return failure.Wrap(err, failure.Messagef("POST %s: リクエストに失敗しました", endpointPath))
	}

	resp, err := c.sess.do(req)
	if err != nil {
		return failure.Wrap(err, failure.Messagef("POST %s: リクエストに失敗しました", endpointPath))
	}
	defer resp.Body.Close()

	if err := bencherror.NewHTTPStatusCodeError(req, resp, opts.wantStatusCode); err != nil {
		return failure.Wrap(err, failure.Messagef("POST %s: ステータスコードが不正です: got=%d, want=%d", endpointPath, resp.StatusCode, http.StatusOK))
	}

	return nil
}

// ListStations は駅一覧列挙APIです
func (c *Client) ListStations(ctx context.Context, opt ...ClientOption) ([]*Station, error) {
	opts := newClientOptions(http.StatusOK, opt...)
	u := *c.baseURL
	endpointPath := endpoint.GetPath(endpoint.ListStations)
	u.Path = filepath.Join(u.Path, endpointPath)

	req, err := c.sess.newRequest(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return []*Station{}, failure.Wrap(err, failure.Messagef("GET %s: リクエストに失敗しました", endpointPath))
	}

	resp, err := c.sess.do(req)
	if err != nil {
		return []*Station{}, failure.Wrap(err, failure.Messagef("GET %s: リクエストに失敗しました", endpointPath))
	}
	defer resp.Body.Close()

	if err := bencherror.NewHTTPStatusCodeError(req, resp, opts.wantStatusCode); err != nil {
		return []*Station{}, failure.Wrap(err, failure.Messagef("GET %s: ステータスコードが不正です: got=%d, want=%d", endpointPath, resp.StatusCode, opts.wantStatusCode))
	}

	var stations []*Station
	if err := json.NewDecoder(resp.Body).Decode(&stations); err != nil {
		return []*Station{}, failure.Wrap(err, failure.Messagef("GET %s: レスポンスのUnmarshalに失敗しました", endpointPath))
	}

	endpoint.IncPathCounter(endpoint.ListStations)

	return stations, nil
}

// SearchTrains は 列車検索APIです
func (c *Client) SearchTrains(ctx context.Context, useAt time.Time, from, to, train_class string, opt ...ClientOption) (Trains, error) {
	opts := newClientOptions(http.StatusOK, opt...)
	u := *c.baseURL
	endpointPath := endpoint.GetPath(endpoint.SearchTrains)
	u.Path = filepath.Join(u.Path, endpointPath)

	failureCtx := failure.Context{
		"use_at": util.FormatISO8601(useAt),
		"from":   from,
		"to":     to,
	}

	req, err := c.sess.newRequest(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return Trains{}, failure.Wrap(err, failure.Messagef("GET %s: 列車検索リクエストに失敗しました", endpointPath), failureCtx)
	}

	query := req.URL.Query()
	query.Set("use_at", util.FormatISO8601(useAt))
	query.Set("train_class", train_class)
	query.Set("from", from)
	query.Set("to", to)
	req.URL.RawQuery = query.Encode()

	resp, err := c.sess.do(req)
	if err != nil {
		return Trains{}, failure.Wrap(err, failure.Messagef("GET %s: 列車検索リクエストに失敗しました", endpointPath), failureCtx)
	}
	defer resp.Body.Close()

	if err := bencherror.NewHTTPStatusCodeError(req, resp, opts.wantStatusCode); err != nil {
		return Trains{}, failure.Wrap(err, failure.Messagef("GET %s: ステータスコードが不正です: got=%d, want=%d", endpointPath, resp.StatusCode, opts.wantStatusCode), failureCtx)
	}

	var trains Trains
	if err := json.NewDecoder(resp.Body).Decode(&trains); err != nil {
		// FIXME: 実装
		return Trains{}, failure.Wrap(err, failure.Messagef("GET %s: レスポンスのUnmarshalに失敗しました", endpointPath), failureCtx)
	}

	endpoint.IncPathCounter(endpoint.SearchTrains)

	return trains, nil
}

func (c *Client) ListTrainSeats(ctx context.Context, date time.Time, trainClass, trainName string, carNum int, departure, arrival string, opt ...ClientOption) (*TrainSeatSearchResponse, TrainSeats, error) {
	opts := newClientOptions(http.StatusOK, opt...)
	u := *c.baseURL
	endpointPath := endpoint.GetPath(endpoint.ListTrainSeats)
	u.Path = filepath.Join(u.Path, endpointPath)

	seats := TrainSeats{}

	failureCtx := failure.Context{
		"date":        util.FormatISO8601(date),
		"train_class": trainClass,
		"train_name":  trainName,
		"car_number":  strconv.Itoa(carNum),
		"from":        departure,
		"to":          arrival,
	}
	lgr := zap.S()

	req, err := c.sess.newRequest(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		lgr.Warnf("座席列挙 リクエスト作成に失敗: %+v", err)
		return nil, seats, failure.Wrap(err, failure.Messagef("GET %s: リクエストに失敗しました", endpointPath), failureCtx)
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
		return nil, seats, failure.Wrap(err, failure.Messagef("GET %s: リクエストに失敗しました", endpointPath), failureCtx)
	}
	defer resp.Body.Close()

	if err := bencherror.NewHTTPStatusCodeError(req, resp, opts.wantStatusCode); err != nil {
		lgr.Warnf("座席列挙 ステータスコードが不正: %+v", err)
		return nil, seats, failure.Wrap(err, failure.Messagef("GET %s: ステータスコードが不正です: got=%d, want=%d", endpointPath, resp.StatusCode, opts.wantStatusCode), failureCtx)
	}

	var listTrainSeatsResp *TrainSeatSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&listTrainSeatsResp); err != nil {
		lgr.Warnf("座席列挙Unmarshal失敗: %+v", err)
		return nil, seats, failure.Wrap(err, failure.Messagef("GET %s: レスポンスのUnmarshalに失敗しました", endpointPath), failureCtx)
	}

	// NotFound、あるいはBadRequestの場合、座席を得ることはできない
	if opts.autoAssert && (resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusBadRequest) {
		validSeats, err := assertListTrainSeats(listTrainSeatsResp, opts.trainSeatsCount)
		if err != nil {
			return nil, seats, failure.Wrap(err, failure.Messagef("GET %s: 座席検索の結果、座席が空になっています", endpointPath))
		}

		seats = validSeats
	}

	endpoint.IncPathCounter(endpoint.ListTrainSeats)

	return listTrainSeatsResp, seats, nil
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
	typ string,
	opt ...ClientOption,
) (*ReserveRequest, *ReservationResponse, error) {
	opts := newClientOptions(http.StatusOK, opt...)
	u := *c.baseURL
	endpointPath := endpoint.GetPath(endpoint.Reserve)
	u.Path = filepath.Join(u.Path, endpointPath)

	failureCtx := failure.Context{
		"train_class": trainClass,
		"train_name":  trainName,
		"seat_class":  seatClass,
		"seats":       fmt.Sprintf("%+v", seats),
		"departure":   departure,
		"arrival":     arrival,
		"date":        util.FormatISO8601(useAt),
		"car_num":     fmt.Sprintf("%d", carNum),
		"child":       fmt.Sprintf("%d", child),
		"adult":       fmt.Sprintf("%d", adult),
		"type":        typ,
	}
	lgr := zap.S()

	reserveReq := &ReserveRequest{
		TrainClass: trainClass,
		TrainName:  trainName,
		SeatClass:  seatClass,
		Seats:      seats,
		Departure:  departure,
		Arrival:    arrival,
		Date:       useAt,
		CarNum:     carNum,
		Child:      child,
		Adult:      adult,
		Type:       typ,
	}

	b, err := json.Marshal(reserveReq)
	if err != nil {
		return reserveReq, nil, err
	}

	lgr.Infof("予約クエリ: %s", string(b))

	req, err := c.sess.newRequest(ctx, http.MethodPost, u.String(), bytes.NewBuffer(b))
	if err != nil {
		lgr.Warnf("予約リクエスト失敗: %+v", err)
		return reserveReq, nil, failure.Wrap(err, failure.Messagef("POST %s: リクエストに失敗しました", endpointPath), failureCtx)
	}

	// FIXME: csrfトークン検証
	// _, err = req.Cookie("csrf_token")
	// if err != nil {
	// 	return nil, failure.Wrap(err, failure.Messagef("POST %s: CSRFトークンが不正です", endpointPath), failureCtx)
	// }

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.sess.do(req)
	if err != nil {
		lgr.Warnf("予約リクエスト失敗: %+v", err)
		return reserveReq, nil, failure.Wrap(err, failure.Messagef("POST %s: リクエストに失敗しました", endpointPath), failureCtx)
	}
	defer resp.Body.Close()

	if err := bencherror.NewHTTPStatusCodeError(req, resp, opts.wantStatusCode); err != nil {
		lgr.Warnf("予約リクエストのレスポンスステータス不正: %+v", err)
		return reserveReq, nil, failure.Wrap(err, failure.Messagef("POST %s: ステータスコードが不正です: got=%d, want=%d", endpointPath, resp.StatusCode, opts.wantStatusCode), failureCtx)
	}

	var reservation *ReservationResponse
	if err := json.NewDecoder(resp.Body).Decode(&reservation); err != nil {
		lgr.Warnf("予約リクエストのUnmarshal失敗: %+v", err)
		return reserveReq, nil, failure.Wrap(err, failure.Messagef("POST %s: JSONのUnmarshalに失敗しました", endpointPath), failureCtx)
	}

	if opts.autoAssert {
		if err := assertReserve(ctx, c, reserveReq, reservation); err != nil {
			return reserveReq, nil, err
		}
	}

	endpoint.IncPathCounter(endpoint.Reserve)
	if SeatAvailability(seatClass) != SaNonReserved {
		endpoint.AddExtraScore(endpoint.Reserve, config.ReservedSeatExtraScore)
	}

	ReservationCache.Add(c.loginUser, reserveReq, reservation.ReservationID)

	return reserveReq, reservation, nil
}

func (c *Client) CommitReservation(ctx context.Context, reservationID int, cardToken string, opt ...ClientOption) error {
	opts := newClientOptions(http.StatusOK, opt...)
	u := *c.baseURL
	endpointPath := endpoint.GetPath(endpoint.CommitReservation)
	u.Path = filepath.Join(u.Path, endpointPath)

	lgr := zap.S()
	failureCtx := failure.Context{
		"reservation_id": fmt.Sprintf("%d", reservationID),
		"card_token":     cardToken,
	}

	// FIXME: 一応構造体にする？
	lgr.Infow("予約確定処理",
		"reservation_id", reservationID,
		"card_token", cardToken,
	)

	b, err := json.Marshal(map[string]interface{}{
		"reservation_id": reservationID,
		"card_token":     cardToken,
	})
	if err != nil {
		return failure.Wrap(err, failure.Messagef("POST %s: Marshalに失敗しました", endpointPath), failureCtx)
	}

	req, err := c.sess.newRequest(ctx, http.MethodPost, u.String(), bytes.NewBuffer(b))
	if err != nil {
		return failure.Wrap(err, failure.Messagef("POST %s: リクエストの作成に失敗しました", endpointPath), failureCtx)
	}

	resp, err := c.sess.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := bencherror.NewHTTPStatusCodeError(req, resp, opts.wantStatusCode); err != nil {
		return failure.Wrap(err, failure.Messagef("POST %s: ステータスコードが不正です: got=%d, want=%d", endpointPath, resp.StatusCode, opts.wantStatusCode), failureCtx)
	}

	endpoint.IncPathCounter(endpoint.CommitReservation)

	return nil
}

func (c *Client) ListReservations(ctx context.Context, opt ...ClientOption) ([]*ReservationResponse, error) {
	opts := newClientOptions(http.StatusOK, opt...)
	u := *c.baseURL
	endpointPath := endpoint.GetPath(endpoint.ListReservations)
	u.Path = filepath.Join(u.Path, endpointPath)

	req, err := c.sess.newRequest(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return []*ReservationResponse{}, failure.Wrap(err, failure.Messagef("GET %s: リクエストに失敗しました", endpointPath))
	}

	resp, err := c.sess.do(req)
	if err != nil {
		return []*ReservationResponse{}, failure.Wrap(err, failure.Messagef("GET %s: リクエストに失敗しました", endpointPath))
	}
	defer resp.Body.Close()

	if err := bencherror.NewHTTPStatusCodeError(req, resp, opts.wantStatusCode); err != nil {
		return []*ReservationResponse{}, failure.Wrap(err, failure.Messagef("GET %s: ステータスコードが不正です: got=%d, want=%d", endpointPath, resp.StatusCode, opts.wantStatusCode))
	}

	var reservations []*ReservationResponse
	if err := json.NewDecoder(resp.Body).Decode(&reservations); err != nil {
		return []*ReservationResponse{}, failure.Wrap(err, failure.Messagef("GET %s: 予約のMarshalに失敗しました", endpointPath))
	}

	endpoint.IncPathCounter(endpoint.ListReservations)

	return reservations, nil
}

func (c *Client) ShowReservation(ctx context.Context, reservationID int, opt ...ClientOption) (*ReservationResponse, error) {
	opts := newClientOptions(http.StatusOK, opt...)
	u := *c.baseURL
	endpointPath := endpoint.GetDynamicPath(endpoint.ShowReservation, reservationID)
	u.Path = filepath.Join(u.Path, endpointPath)

	failureCtx := failure.Context{
		"reservation_id": fmt.Sprintf("%d", reservationID),
	}

	req, err := c.sess.newRequest(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, failure.Wrap(err, failure.Messagef("GET %s: リクエストに失敗しました", endpointPath), failureCtx)
	}

	resp, err := c.sess.do(req)
	if err != nil {
		return nil, failure.Wrap(err, failure.Messagef("GET %s: リクエストに失敗しました", endpointPath), failureCtx)
	}

	if err := bencherror.NewHTTPStatusCodeError(req, resp, opts.wantStatusCode); err != nil {
		return nil, failure.Wrap(err, failure.Messagef("GET %s: ステータスコードが不正です: got=%d, want=%d", endpointPath, resp.StatusCode, opts.wantStatusCode))
	}

	var reservation *ReservationResponse
	if err := json.NewDecoder(resp.Body).Decode(&reservation); err != nil {
		return nil, failure.Wrap(err, failure.Messagef("GET %s: Unmarshalに失敗しました", endpointPath), failureCtx)
	}

	endpoint.IncDynamicPathCounter(endpoint.ShowReservation)

	return reservation, nil
}

func (c *Client) CancelReservation(ctx context.Context, reservationID int, opt ...ClientOption) error {
	opts := newClientOptions(http.StatusOK, opt...)
	u := *c.baseURL
	endpointPath := endpoint.GetDynamicPath(endpoint.CancelReservation, reservationID)
	u.Path = filepath.Join(u.Path, endpointPath)

	failureCtx := failure.Context{
		"reservation_id": fmt.Sprintf("%d", reservationID),
	}

	lgr := zap.S()

	lgr.Infow("予約キャンセル処理",
		"reservation_id", reservationID,
	)

	req, err := c.sess.newRequest(ctx, http.MethodPost, u.String(), nil)
	if err != nil {
		return failure.Wrap(err, failure.Messagef("POST %s: リクエストに失敗しました", endpointPath), failureCtx)
	}

	resp, err := c.sess.do(req)
	if err != nil {
		return failure.Wrap(err, failure.Messagef("POST %s: リクエストに失敗しました", endpointPath), failureCtx)
	}
	defer resp.Body.Close()

	if err := bencherror.NewHTTPStatusCodeError(req, resp, opts.wantStatusCode); err != nil {
		return failure.Wrap(err, failure.Messagef("POST %s: ステータスコードが不正です: got=%d, want=%d", endpointPath, resp.StatusCode, opts.wantStatusCode), failureCtx)
	}

	if opts.autoAssert {
		if err := assertCancelReservation(ctx, c, reservationID); err != nil {
			return err
		}
	}

	endpoint.IncDynamicPathCounter(endpoint.CancelReservation)

	return nil
}

func (c *Client) DownloadAsset(ctx context.Context, path string) ([]byte, error) {
	u := *c.baseURL
	u.Path = filepath.Join(u.Path, path)

	req, err := c.sess.newRequest(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return []byte{}, bencherror.PreTestErrs.AddError(bencherror.NewCriticalError(err, "GET %s: 静的ファイルのダウンロードに失敗しました", path))
	}

	resp, err := c.sess.do(req)
	if err != nil {
		return []byte{}, bencherror.PreTestErrs.AddError(bencherror.NewCriticalError(err, "GET %s: 静的ファイルのダウンロードに失敗しました", path))
	}
	defer resp.Body.Close()

	if err := bencherror.NewHTTPStatusCodeError(req, resp, http.StatusOK); err != nil {
		return []byte{}, bencherror.PreTestErrs.AddError(bencherror.NewCriticalError(err, "GET %s: ステータスコードが不正です: got=%d, want=%d", path, resp.StatusCode, http.StatusOK))
	}

	return ioutil.ReadAll(resp.Body)
}
