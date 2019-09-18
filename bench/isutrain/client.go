package isutrain

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/chibiegg/isucon9-final/bench/internal/bencherror"
	"github.com/chibiegg/isucon9-final/bench/internal/config"
	"github.com/chibiegg/isucon9-final/bench/internal/util"
	"github.com/morikuni/failure"
)

type ClientOption struct {
	wantStatusCode int
}

type Client struct {
	sess    *session
	baseURL string
}

func NewClient(targetBaseURL string) (*Client, error) {
	sess, err := newSession()
	if err != nil {
		return nil, err
	}
	return &Client{
		sess:    sess,
		baseURL: targetBaseURL,
	}, nil
}

func NewClientForInitialize(targetBaseURL string) (*Client, error) {
	sess, err := newSessionForInitialize()
	if err != nil {
		return nil, err
	}
	return &Client{
		sess:    sess,
		baseURL: targetBaseURL,
	}, nil
}

// ReplaceMockTransport は、clientの利用するhttp.RoundTripperを、DefaultTransportに差し替えます
// NOTE: httpmockはhttp.DefaultTransportを利用するため、モックテストの時この関数を利用する
func (c *Client) ReplaceMockTransport() {
	c.sess.httpClient.Transport = http.DefaultTransport
}

func (c *Client) Initialize(ctx context.Context) {
	uri := fmt.Sprintf("%s%s", c.baseURL, config.InitializePath)

	ctx, cancel := context.WithTimeout(ctx, config.InitializeTimeout)
	defer cancel()

	req, err := c.sess.newRequest(ctx, http.MethodPost, uri, nil)
	if err != nil {
		bencherror.InitializeErrs.AddError(err)
		return
	}

	// TODO: 言語を返すようにしたり、キャンペーンを返すようにする場合、ちゃんと設定されていなかったらFAILにする
	resp, err := c.sess.do(req)
	if err != nil {
		bencherror.InitializeErrs.AddError(err)
		return
	}
	defer resp.Body.Close()

	if err := bencherror.NewHTTPStatusCodeError(resp, http.StatusAccepted); err != nil {
		bencherror.InitializeErrs.AddError(err)
		return
	}
}

func (c *Client) Register(ctx context.Context, username, password string, opts *ClientOption) error {
	var (
		uri  = fmt.Sprintf("%s%s", c.baseURL, config.RegisterPath)
		form = url.Values{}
	)
	form.Set("username", username)
	form.Set("password", password)

	req, err := c.sess.newRequest(ctx, http.MethodPost, uri, bytes.NewBufferString(form.Encode()))
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
		if err := bencherror.NewHTTPStatusCodeError(resp, http.StatusAccepted); err != nil {
			bencherror.BenchmarkErrs.AddError(failure.Wrap(err, failure.Message("POST /register: ステータスコードが不正です")))
			return err
		}
	} else {
		if err := bencherror.NewHTTPStatusCodeError(resp, opts.wantStatusCode); err != nil {
			bencherror.BenchmarkErrs.AddError(failure.Wrap(err, failure.Message("POST /register: ステータスコードが不正です")))
			return err
		}
	}

	return nil
}

func (c *Client) Login(ctx context.Context, username, password string, opts *ClientOption) error {
	var (
		uri  = fmt.Sprintf("%s%s", c.baseURL, config.LoginPath)
		form = url.Values{}
	)
	form.Set("username", username)
	form.Set("password", password)

	req, err := c.sess.newRequest(ctx, http.MethodPost, uri, bytes.NewBufferString(form.Encode()))
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
		if err := bencherror.NewHTTPStatusCodeError(resp, http.StatusAccepted); err != nil {
			bencherror.BenchmarkErrs.AddError(failure.Wrap(err, failure.Message("POST /login: ステータスコードが不正です")))
			return err
		}
	} else {
		if err := bencherror.NewHTTPStatusCodeError(resp, opts.wantStatusCode); err != nil {
			bencherror.BenchmarkErrs.AddError(failure.Wrap(err, failure.Message("POST /login: ステータスコードが不正です")))
			return err
		}
	}

	return nil
}

// ListStations は駅一覧列挙APIです
func (c *Client) ListStations(ctx context.Context, opts *ClientOption) ([]*Station, error) {
	uri := fmt.Sprintf("%s%s", c.baseURL, config.ListStationsPath)

	req, err := c.sess.newRequest(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return []*Station{}, err
	}

	resp, err := c.sess.do(req)
	if err != nil {
		return []*Station{}, err
	}
	defer resp.Body.Close()

	if opts == nil {
		if err := bencherror.NewHTTPStatusCodeError(resp, http.StatusOK); err != nil {
			bencherror.BenchmarkErrs.AddError(failure.Wrap(err, failure.Message("GET /stations: ステータスコードが不正です")))
			return []*Station{}, err
		}
	} else {
		if err := bencherror.NewHTTPStatusCodeError(resp, opts.wantStatusCode); err != nil {
			bencherror.BenchmarkErrs.AddError(failure.Wrap(err, failure.Message("GET /stations: ステータスコードが不正です")))
			return []*Station{}, err
		}
	}

	var stations []*Station
	if err := json.NewDecoder(resp.Body).Decode(&stations); err != nil {
		// FIXME: 実装
		return []*Station{}, err
	}

	return stations, nil
}

// SearchTrains は 列車検索APIです
func (c *Client) SearchTrains(ctx context.Context, useAt time.Time, from, to string, opts *ClientOption) ([]*TrainSearchResponse, error) {
	uri := fmt.Sprintf("%s%s", c.baseURL, config.SearchTrainsPath)

	req, err := c.sess.newRequest(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return []*TrainSearchResponse{}, err
	}

	query := req.URL.Query()
	query.Set("use_at", util.FormatISO8601(useAt))
	query.Set("train_class", "") // FIXME: 列車種別
	query.Set("from", from)
	query.Set("to", to)
	req.URL.RawQuery = query.Encode()

	resp, err := c.sess.do(req)
	if err != nil {
		return []*TrainSearchResponse{}, err
	}
	defer resp.Body.Close()

	if opts == nil {
		if err := bencherror.NewHTTPStatusCodeError(resp, http.StatusOK); err != nil {
			bencherror.BenchmarkErrs.AddError(failure.Wrap(err, failure.Message("GET /trains: ステータスコードが不正です")))
			return []*TrainSearchResponse{}, err
		}
	} else {
		if err := bencherror.NewHTTPStatusCodeError(resp, opts.wantStatusCode); err != nil {
			bencherror.BenchmarkErrs.AddError(failure.Wrap(err, failure.Message("GET /trains: ステータスコードが不正です")))
			return []*TrainSearchResponse{}, err
		}
	}

	var trains []*TrainSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&trains); err != nil {
		// FIXME: 実装
		return []*TrainSearchResponse{}, err
	}

	return trains, nil
}

func (c *Client) ListTrainSeats(ctx context.Context, train_class, train_name string, opts *ClientOption) (TrainSeats, error) {
	uri := fmt.Sprintf("%s%s", c.baseURL, config.ListTrainSeatsPath)

	req, err := c.sess.newRequest(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return TrainSeats{}, err
	}

	query := req.URL.Query()
	query.Set("train_class", train_class)
	query.Set("train_name", train_name)
	req.URL.RawQuery = query.Encode()

	resp, err := c.sess.do(req)
	if err != nil {
		return TrainSeats{}, err
	}
	defer resp.Body.Close()

	if opts == nil {
		if err := bencherror.NewHTTPStatusCodeError(resp, http.StatusOK); err != nil {
			bencherror.BenchmarkErrs.AddError(failure.Wrap(err, failure.Message("GET /train/seats: ステータスコードが不正です")))
			return TrainSeats{}, err
		}
	} else {
		if err := bencherror.NewHTTPStatusCodeError(resp, opts.wantStatusCode); err != nil {
			bencherror.BenchmarkErrs.AddError(failure.Wrap(err, failure.Message("GET /train/seats: ステータスコードが不正です")))
			return TrainSeats{}, err
		}
	}

	var seats TrainSeats
	if err := json.NewDecoder(resp.Body).Decode(&seats); err != nil {
		return TrainSeats{}, err
	}

	return seats, nil
}

func (c *Client) Reserve(ctx context.Context, opts *ClientOption) (*ReservationResponse, error) {
	var (
		uri = fmt.Sprintf("%s%s", c.baseURL, config.ReservePath)
		// form = url.Values{}
	)

	req, err := c.sess.newRequest(ctx, http.MethodPost, uri, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.sess.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if opts == nil {
		if err := bencherror.NewHTTPStatusCodeError(resp, http.StatusAccepted); err != nil {
			bencherror.BenchmarkErrs.AddError(failure.Wrap(err, failure.Message("POST /reserve: ステータスコードが不正です")))
			return nil, err
		}
	} else {
		if err := bencherror.NewHTTPStatusCodeError(resp, opts.wantStatusCode); err != nil {
			bencherror.BenchmarkErrs.AddError(failure.Wrap(err, failure.Message("POST /reserve: ステータスコードが不正です")))
			return nil, err
		}
	}

	var reservation *ReservationResponse
	if err := json.NewDecoder(resp.Body).Decode(&reservation); err != nil {
		bencherror.BenchmarkErrs.AddError(failure.Wrap(err, failure.Message("POST /reserve: JSONのUnmarshalに失敗しました")))
		return nil, err
	}

	return reservation, nil
}

func (c *Client) CommitReservation(ctx context.Context, reservationID string, opts *ClientOption) error {
	uri := fmt.Sprintf("%s%s", c.baseURL, config.BuildCommitReservationPath(reservationID))

	req, err := c.sess.newRequest(ctx, http.MethodPost, uri, nil)
	if err != nil {
		return err
	}

	resp, err := c.sess.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if opts == nil {
		if err := bencherror.NewHTTPStatusCodeError(resp, http.StatusAccepted); err != nil {
			bencherror.BenchmarkErrs.AddError(failure.Wrap(err, failure.Message("POST /reservation/:id/commit: ステータスコードが不正です")))
			return err
		}
	} else {
		if err := bencherror.NewHTTPStatusCodeError(resp, opts.wantStatusCode); err != nil {
			bencherror.BenchmarkErrs.AddError(failure.Wrap(err, failure.Message("POST /reservation/:id/commit: ステータスコードが不正です")))
			return err
		}
	}

	return nil
}

func (c *Client) ListReservations(ctx context.Context, opts *ClientOption) ([]*SeatReservation, error) {
	uri := fmt.Sprintf("%s%s", c.baseURL, config.ListReservationsPath)

	req, err := c.sess.newRequest(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return []*SeatReservation{}, err
	}

	resp, err := c.sess.do(req)
	if err != nil {
		return []*SeatReservation{}, err
	}
	defer resp.Body.Close()

	if opts == nil {
		if err := bencherror.NewHTTPStatusCodeError(resp, http.StatusOK); err != nil {
			bencherror.BenchmarkErrs.AddError(failure.Wrap(err, failure.Message("GET /reservations: ステータスコードが不正です")))
			return []*SeatReservation{}, err
		}
	} else {
		if err := bencherror.NewHTTPStatusCodeError(resp, opts.wantStatusCode); err != nil {
			bencherror.BenchmarkErrs.AddError(failure.Wrap(err, failure.Message("GET /reservations: ステータスコードが不正です")))
			return []*SeatReservation{}, err
		}
	}

	var reservations []*SeatReservation
	if err := json.NewDecoder(resp.Body).Decode(&reservations); err != nil {
		return []*SeatReservation{}, err
	}

	return reservations, nil
}

func (c *Client) CancelReservation(ctx context.Context, reservationID string, opts *ClientOption) error {
	uri := fmt.Sprintf("%s%s", c.baseURL, config.BuildCancelReservationPath(reservationID))

	req, err := c.sess.newRequest(ctx, http.MethodDelete, uri, nil)
	if err != nil {
		return err
	}

	resp, err := c.sess.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if opts == nil {
		if err := bencherror.NewHTTPStatusCodeError(resp, http.StatusNoContent); err != nil {
			bencherror.BenchmarkErrs.AddError(failure.Wrap(err, failure.Message("DELETE /reservation/:id/cancel: ステータスコードが不正です")))
			return err
		}
	} else {
		if err := bencherror.NewHTTPStatusCodeError(resp, opts.wantStatusCode); err != nil {
			bencherror.BenchmarkErrs.AddError(failure.Wrap(err, failure.Message("DELETE /reservation/:id/cancel: ステータスコードが不正です")))
			return err
		}
	}

	return nil
}
