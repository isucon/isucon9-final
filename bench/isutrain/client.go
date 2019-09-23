package isutrain

import (
	"bytes"
	"context"
	"encoding/json"
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

type ClientOption struct {
	wantStatusCode int
}

type Client struct {
	sess    *Session
	baseURL *url.URL
}

func NewClient() (*Client, error) {
	sess, err := NewSession()
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(config.TargetBaseURL)
	if err != nil {
		return nil, err
	}

	return &Client{
		sess:    sess,
		baseURL: u,
	}, nil
}

func NewClientForInitialize() (*Client, error) {
	sess, err := newSessionForInitialize()
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(config.TargetBaseURL)
	if err != nil {
		return nil, err
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
	u.Path = filepath.Join(u.Path, endpoint.GetPath(endpoint.Initialize))

	ctx, cancel := context.WithTimeout(ctx, config.InitializeTimeout)
	defer cancel()

	req, err := c.sess.newRequest(ctx, http.MethodPost, u.String(), nil)
	if err != nil {
		bencherror.InitializeErrs.AddError(bencherror.NewApplicationError(err, "POST /initialize: リクエストに失敗しました"))
		return
	}

	// TODO: 言語を返すようにしたり、キャンペーンを返すようにする場合、ちゃんと設定されていなかったらFAILにする
	resp, err := c.sess.do(req)
	if err != nil {
		bencherror.InitializeErrs.AddError(bencherror.NewApplicationError(err, "POST /initialize: リクエストに失敗しました"))
		return
	}
	defer resp.Body.Close()

	if err := bencherror.NewHTTPStatusCodeError(req, resp, http.StatusOK); err != nil {
		bencherror.InitializeErrs.AddError(err)
		return
	}

	// FIXME: 予約可能日数をレスポンスから受け取る
	if err := config.SetAvailReserveDays(30); err != nil {
		bencherror.InitializeErrs.AddError(bencherror.NewCriticalError(err, "POST /initialize: 予約可能日数の設定に失敗しました"))
		return
	}

	endpoint.IncPathCounter(endpoint.Initialize)
}

func (c *Client) Settings(ctx context.Context) (*Settings, error) {
	u := *c.baseURL
	u.Path = filepath.Join(u.Path, endpoint.GetPath(endpoint.Settings))

	req, err := c.sess.newRequest(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, failure.Wrap(err, failure.Message("GET /settings: 設定情報の取得に失敗しました"))
	}

	resp, err := c.sess.do(req)
	if err != nil {
		return nil, failure.Wrap(err, failure.Message("GET /settings: 設定情報の取得に失敗しました"))
	}
	defer resp.Body.Close()

	var settings *Settings
	if err := json.NewDecoder(resp.Body).Decode(&settings); err != nil {
		return nil, failure.Wrap(err, failure.Message("GET /settings: レスポンスのUnmarshalに失敗しました"))
	}

	return settings, nil
}

func (c *Client) Signup(ctx context.Context, email, password string, opts *ClientOption) error {
	var (
		u    = *c.baseURL
		form = url.Values{}
	)
	u.Path = filepath.Join(u.Path, endpoint.GetPath(endpoint.Signup))
	form.Set("email", email)
	form.Set("password", password)

	req, err := c.sess.newRequest(ctx, http.MethodPost, u.String(), bytes.NewBufferString(form.Encode()))
	if err != nil {
		return failure.Wrap(err, failure.Message("POST /register: リクエストに失敗しました"))
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.sess.do(req)
	if err != nil {
		return failure.Wrap(err, failure.Message("POST /register: リクエストに失敗しました"))
	}
	defer resp.Body.Close()

	if opts == nil {
		if err := bencherror.NewHTTPStatusCodeError(req, resp, http.StatusOK); err != nil {
			return failure.Wrap(err, failure.Message("POST /register: ステータスコードが不正です"))
		}
	} else {
		if err := bencherror.NewHTTPStatusCodeError(req, resp, opts.wantStatusCode); err != nil {
			return failure.Wrap(err, failure.Message("POST /register: ステータスコードが不正です"))
		}
	}

	endpoint.IncPathCounter(endpoint.Signup)

	return nil
}

func (c *Client) Login(ctx context.Context, username, password string, opts *ClientOption) error {
	var (
		u    = *c.baseURL
		form = url.Values{}
	)
	u.Path = filepath.Join(u.Path, endpoint.GetPath(endpoint.Login))
	form.Set("username", username)
	form.Set("password", password)

	req, err := c.sess.newRequest(ctx, http.MethodPost, u.String(), bytes.NewBufferString(form.Encode()))
	if err != nil {
		return failure.Wrap(err, failure.Message("POST /login: リクエストに失敗しました"))
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.sess.do(req)
	if err != nil {
		return failure.Wrap(err, failure.Message("POST /login: リクエストに失敗しました"))
	}
	defer resp.Body.Close()

	if opts == nil {
		if err := bencherror.NewHTTPStatusCodeError(req, resp, http.StatusOK); err != nil {
			failure.Wrap(err, failure.Message("POST /login: ステータスコードが不正です"))
			return err
		}
	} else {
		if err := bencherror.NewHTTPStatusCodeError(req, resp, opts.wantStatusCode); err != nil {
			failure.Wrap(err, failure.Message("POST /login: ステータスコードが不正です"))
			return err
		}
	}

	endpoint.IncPathCounter(endpoint.Login)

	return nil
}

func (c *Client) Logout(ctx context.Context, opts *ClientOption) error {
	u := *c.baseURL
	u.Path = filepath.Join(u.Path, endpoint.GetPath(endpoint.Logout))

	req, err := c.sess.newRequest(ctx, http.MethodPost, u.String(), nil)
	if err != nil {
		return failure.Wrap(err, failure.Message("POST /logout: リクエストに失敗しました"))
	}

	resp, err := c.sess.do(req)
	if err != nil {
		return failure.Wrap(err, failure.Message("POST /logout: リクエストに失敗しました"))
	}
	defer resp.Body.Close()

	if opts == nil {
		if err := bencherror.NewHTTPStatusCodeError(req, resp, http.StatusNoContent); err != nil {
			failure.Wrap(err, failure.Message("POST /logout: ステータスコードが不正です"))
			return err
		}
	} else {
		if err := bencherror.NewHTTPStatusCodeError(req, resp, opts.wantStatusCode); err != nil {
			failure.Wrap(err, failure.Message("POST /logout: ステータスコードが不正です"))
			return err
		}
	}

	return nil
}

// ListStations は駅一覧列挙APIです
func (c *Client) ListStations(ctx context.Context, opts *ClientOption) ([]*Station, error) {
	u := *c.baseURL
	u.Path = filepath.Join(u.Path, endpoint.GetPath(endpoint.ListStations))

	req, err := c.sess.newRequest(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return []*Station{}, failure.Wrap(err, failure.Message("GET /stations: リクエストに失敗しました"))
	}

	resp, err := c.sess.do(req)
	if err != nil {
		return []*Station{}, failure.Wrap(err, failure.Message("GET /stations: リクエストに失敗しました"))
	}
	defer resp.Body.Close()

	if opts == nil {
		if err := bencherror.NewHTTPStatusCodeError(req, resp, http.StatusOK); err != nil {
			return []*Station{}, failure.Wrap(err, failure.Message("GET /stations: ステータスコードが不正です"))
		}
	} else {
		if err := bencherror.NewHTTPStatusCodeError(req, resp, opts.wantStatusCode); err != nil {
			return []*Station{}, failure.Wrap(err, failure.Message("GET /stations: ステータスコードが不正です"))
		}
	}

	var stations []*Station
	if err := json.NewDecoder(resp.Body).Decode(&stations); err != nil {
		// FIXME: 実装
		return []*Station{}, failure.Wrap(err, failure.Message("GET /stations: レスポンスのUnmarshalに失敗しました"))
	}

	endpoint.IncPathCounter(endpoint.ListStations)

	return stations, nil
}

// SearchTrains は 列車検索APIです
func (c *Client) SearchTrains(ctx context.Context, useAt time.Time, from, to string, opts *ClientOption) (Trains, error) {
	u := *c.baseURL
	u.Path = filepath.Join(u.Path, endpoint.GetPath(endpoint.SearchTrains))

	lgr := zap.S()

	req, err := c.sess.newRequest(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return Trains{}, err
	}

	query := req.URL.Query()
	query.Set("use_at", util.FormatISO8601(useAt))
	query.Set("train_class", "") // FIXME: 列車種別
	query.Set("from", from)
	query.Set("to", to)
	req.URL.RawQuery = query.Encode()

	lgr.Infof("クエリ: %s\n", query.Encode())

	resp, err := c.sess.do(req)
	if err != nil {
		return Trains{}, err
	}
	defer resp.Body.Close()

	if opts == nil {
		if err := bencherror.NewHTTPStatusCodeError(req, resp, http.StatusOK); err != nil {
			return Trains{}, failure.Wrap(err, failure.Message("GET /trains: ステータスコードが不正です"))
		}
	} else {
		if err := bencherror.NewHTTPStatusCodeError(req, resp, opts.wantStatusCode); err != nil {
			return Trains{}, failure.Wrap(err, failure.Message("GET /trains: ステータスコードが不正です"))
		}
	}

	var trains Trains
	if err := json.NewDecoder(resp.Body).Decode(&trains); err != nil {
		// FIXME: 実装
		return Trains{}, failure.Wrap(err, failure.Message("GET /trains: レスポンスのUnmarshalに失敗しました"))
	}

	endpoint.IncPathCounter(endpoint.SearchTrains)

	return trains, nil
}

func (c *Client) ListTrainSeats(ctx context.Context, date time.Time, trainClass, trainName string, carNum int, departure, arrival string, opts *ClientOption) (*TrainSeatSearchResponse, error) {
	u := *c.baseURL
	u.Path = filepath.Join(u.Path, endpoint.GetPath(endpoint.ListTrainSeats))

	lgr := zap.S()

	req, err := c.sess.newRequest(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, failure.Wrap(err, failure.Message("GET /train/seats: リクエストに失敗しました"))
	}

	query := req.URL.Query()
	query.Set("date", util.FormatISO8601(date))
	query.Set("train_class", trainClass)
	query.Set("train_name", trainName)
	query.Set("car_number", strconv.Itoa(carNum))
	query.Set("from", departure)
	query.Set("to", arrival)
	req.URL.RawQuery = query.Encode()

	lgr.Infof("クエリ: %s\n", query.Encode())

	resp, err := c.sess.do(req)
	if err != nil {
		return nil, failure.Wrap(err, failure.Message("GET /train/seats: リクエストに失敗しました"))
	}
	defer resp.Body.Close()

	if opts == nil {
		if err := bencherror.NewHTTPStatusCodeError(req, resp, http.StatusOK); err != nil {
			return nil, failure.Wrap(err, failure.Message("GET /train/seats: ステータスコードが不正です"))
		}
	} else {
		if err := bencherror.NewHTTPStatusCodeError(req, resp, opts.wantStatusCode); err != nil {
			return nil, failure.Wrap(err, failure.Message("GET /train/seats: ステータスコードが不正です"))
		}
	}

	var listTrainSeatsResp *TrainSeatSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&listTrainSeatsResp); err != nil {
		return nil, failure.Wrap(err, failure.Message("GET /train/seats: レスポンスのUnmarshalに失敗しました"))
	}

	endpoint.IncPathCounter(endpoint.ListTrainSeats)

	return listTrainSeatsResp, nil
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
	opts *ClientOption,
) (*ReservationResponse, error) {
	u := *c.baseURL
	u.Path = filepath.Join(u.Path, endpoint.GetPath(endpoint.Reserve))

	b, err := json.Marshal(&ReservationRequest{
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
	})
	if err != nil {
		return nil, err
	}

	log.Printf("reserve request: %s\n", string(b))

	req, err := c.sess.newRequest(ctx, http.MethodPost, u.String(), bytes.NewBuffer(b))
	if err != nil {
		return nil, failure.Wrap(err, failure.Message("POST /reserve: リクエストに失敗しました"))
	}

	// FIXME: csrfトークン検証
	// _, err = req.Cookie("csrf_token")
	// if err != nil {
	// 	return nil, failure.Wrap(err, failure.Message("POST /reserve: CSRFトークンが不正です"))
	// }

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.sess.do(req)
	if err != nil {
		return nil, failure.Wrap(err, failure.Message("POST /reserve: リクエストに失敗しました"))
	}
	defer resp.Body.Close()

	if opts == nil {
		if err := bencherror.NewHTTPStatusCodeError(req, resp, http.StatusOK); err != nil {
			bencherror.BenchmarkErrs.AddError(failure.Wrap(err, failure.Message("POST /reserve: ステータスコードが不正です")))
			return nil, err
		}
	} else {
		if err := bencherror.NewHTTPStatusCodeError(req, resp, opts.wantStatusCode); err != nil {
			bencherror.BenchmarkErrs.AddError(failure.Wrap(err, failure.Message("POST /reserve: ステータスコードが不正です")))
			return nil, err
		}
	}

	var reservation *ReservationResponse
	if err := json.NewDecoder(resp.Body).Decode(&reservation); err != nil {
		bencherror.BenchmarkErrs.AddError(failure.Wrap(err, failure.Message("POST /reserve: JSONのUnmarshalに失敗しました")))
		return nil, err
	}

	endpoint.IncPathCounter(endpoint.Reserve)
	if SeatAvailability(seatClass) != SaNonReserved {
		endpoint.AddExtraScore(endpoint.Reserve, config.ReservedSeatExtraScore)
	}

	return reservation, nil
}

func (c *Client) CommitReservation(ctx context.Context, reservationID int, cardToken string, opts *ClientOption) error {
	u := *c.baseURL
	u.Path = filepath.Join(u.Path, endpoint.GetPath(endpoint.CommitReservation))

	// FIXME: 一応構造体にする？
	b, err := json.Marshal(map[string]interface{}{
		"reservation_id": reservationID,
		"card_token":     cardToken,
	})
	if err != nil {
		return err
	}

	req, err := c.sess.newRequest(ctx, http.MethodPost, u.String(), bytes.NewBuffer(b))
	if err != nil {
		return err
	}

	resp, err := c.sess.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if opts == nil {
		if err := bencherror.NewHTTPStatusCodeError(req, resp, http.StatusOK); err != nil {
			bencherror.BenchmarkErrs.AddError(failure.Wrap(err, failure.Message("POST /reservation/:id/commit: ステータスコードが不正です")))
			return err
		}
	} else {
		if err := bencherror.NewHTTPStatusCodeError(req, resp, opts.wantStatusCode); err != nil {
			bencherror.BenchmarkErrs.AddError(failure.Wrap(err, failure.Message("POST /reservation/:id/commit: ステータスコードが不正です")))
			return err
		}
	}

	endpoint.IncPathCounter(endpoint.CommitReservation)

	return nil
}

func (c *Client) ListReservations(ctx context.Context, opts *ClientOption) ([]*SeatReservation, error) {
	u := *c.baseURL
	u.Path = filepath.Join(u.Path, endpoint.GetPath(endpoint.ListReservations))

	req, err := c.sess.newRequest(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return []*SeatReservation{}, err
	}

	resp, err := c.sess.do(req)
	if err != nil {
		return []*SeatReservation{}, err
	}
	defer resp.Body.Close()

	if opts == nil {
		if err := bencherror.NewHTTPStatusCodeError(req, resp, http.StatusOK); err != nil {
			bencherror.BenchmarkErrs.AddError(failure.Wrap(err, failure.Message("GET /reservations: ステータスコードが不正です")))
			return []*SeatReservation{}, err
		}
	} else {
		if err := bencherror.NewHTTPStatusCodeError(req, resp, opts.wantStatusCode); err != nil {
			bencherror.BenchmarkErrs.AddError(failure.Wrap(err, failure.Message("GET /reservations: ステータスコードが不正です")))
			return []*SeatReservation{}, err
		}
	}

	var reservations []*SeatReservation
	if err := json.NewDecoder(resp.Body).Decode(&reservations); err != nil {
		return []*SeatReservation{}, err
	}

	endpoint.IncPathCounter(endpoint.ListReservations)

	return reservations, nil
}

func (c *Client) ShowReservation(ctx context.Context, reservationID int, opts *ClientOption) (*SeatReservation, error) {
	u := *c.baseURL
	u.Path = filepath.Join(u.Path, endpoint.GetDynamicPath(endpoint.ShowReservation, reservationID))

	req, err := c.sess.newRequest(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.sess.do(req)
	if err != nil {
		return nil, err
	}

	var reservation *SeatReservation
	if err := json.NewDecoder(resp.Body).Decode(&reservation); err != nil {
		return nil, err
	}

	endpoint.IncDynamicPathCounter(endpoint.ShowReservation)

	return reservation, nil
}

func (c *Client) CancelReservation(ctx context.Context, reservationID int, opts *ClientOption) error {
	u := *c.baseURL
	u.Path = filepath.Join(u.Path, endpoint.GetDynamicPath(endpoint.CancelReservation, reservationID))

	log.Printf("cancel reservation uri: %s\n", u.String())

	req, err := c.sess.newRequest(ctx, http.MethodPost, u.String(), nil)
	if err != nil {
		return err
	}

	resp, err := c.sess.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if opts == nil {
		if err := bencherror.NewHTTPStatusCodeError(req, resp, http.StatusOK); err != nil {
			bencherror.BenchmarkErrs.AddError(failure.Wrap(err, failure.Message("DELETE /reservation/:id/cancel: ステータスコードが不正です")))
			return err
		}
	} else {
		if err := bencherror.NewHTTPStatusCodeError(req, resp, opts.wantStatusCode); err != nil {
			bencherror.BenchmarkErrs.AddError(failure.Wrap(err, failure.Message("DELETE /reservation/:id/cancel: ステータスコードが不正です")))
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
		return []byte{}, bencherror.PreTestErrs.AddError(bencherror.NewCriticalError(err, "GET %s: ステータスコードが不正です", path))
	}

	return ioutil.ReadAll(resp.Body)
}
