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
	"github.com/chibiegg/isucon9-final/bench/internal/consts"
	"github.com/chibiegg/isucon9-final/bench/internal/util"
)

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
	uri := fmt.Sprintf("%s%s", c.baseURL, consts.InitializePath)

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

	return
}

func (c *Client) Register(ctx context.Context, username, password string) error {
	var (
		uri  = fmt.Sprintf("%s%s", c.baseURL, consts.RegisterPath)
		form = url.Values{}
	)
	form.Set("username", username)
	form.Set("password", password)

	req, err := c.sess.newRequest(ctx, http.MethodPost, uri, bytes.NewBufferString(form.Encode()))
	if err != nil {
		return err
	}

	resp, err := c.sess.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		// FIXME: 実装
	}

	return nil
}

func (c *Client) Login(ctx context.Context, username, password string) error {
	var (
		uri  = fmt.Sprintf("%s%s", c.baseURL, consts.LoginPath)
		form = url.Values{}
	)
	form.Set("username", username)
	form.Set("password", password)

	req, err := c.sess.newRequest(ctx, http.MethodPost, uri, bytes.NewBufferString(form.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.sess.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// FIXME: 実装
	}

	return nil
}

// ListStations は駅一覧列挙APIです
func (c *Client) ListStations(ctx context.Context) ([]*Station, error) {
	uri := fmt.Sprintf("%s%s", c.baseURL, consts.ListStationsPath)

	req, err := c.sess.newRequest(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return []*Station{}, err
	}

	resp, err := c.sess.do(req)
	if err != nil {
		return []*Station{}, err
	}
	defer resp.Body.Close()

	var stations []*Station
	if err := json.NewDecoder(resp.Body).Decode(&stations); err != nil {
		// FIXME: 実装
		return []*Station{}, err
	}

	return stations, nil
}

// SearchTrains は 列車検索APIです
func (c *Client) SearchTrains(ctx context.Context, useAt time.Time, from, to string) ([]*TrainSearchResponse, error) {
	uri := fmt.Sprintf("%s%s", c.baseURL, consts.SearchTrainsPath)

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

	var trains []*TrainSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&trains); err != nil {
		// FIXME: 実装
		return []*TrainSearchResponse{}, err
	}

	return trains, nil
}

func (c *Client) ListTrainSeats(ctx context.Context, train_class, train_name string) (TrainSeats, error) {
	uri := fmt.Sprintf("%s%s", c.baseURL, consts.ListTrainSeatsPath)

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

	var seats TrainSeats
	if err := json.NewDecoder(resp.Body).Decode(&seats); err != nil {
		return TrainSeats{}, err
	}

	return seats, nil
}

func (c *Client) Reserve(ctx context.Context) (*ReservationResponse, error) {
	var (
		uri = fmt.Sprintf("%s%s", c.baseURL, consts.ReservePath)
		// form = url.Values{}
	)

	req, err := c.sess.newRequest(ctx, http.MethodPost, uri, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.sess.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var reservation *ReservationResponse
	if err := json.NewDecoder(resp.Body).Decode(&reservation); err != nil {
		return nil, err
	}

	return reservation, nil
}

func (c *Client) CommitReservation(ctx context.Context, reservationID string) error {
	uri := fmt.Sprintf("%s%s", c.baseURL, consts.BuildCommitReservationPath(reservationID))

	req, err := c.sess.newRequest(ctx, http.MethodPost, uri, nil)
	if err != nil {
		return err
	}

	resp, err := c.sess.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		// FIXME: 実装
	}

	return nil
}

func (c *Client) ListReservations(ctx context.Context) ([]*SeatReservation, error) {
	uri := fmt.Sprintf("%s%s", c.baseURL, consts.ListReservationsPath)

	req, err := c.sess.newRequest(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return []*SeatReservation{}, err
	}

	resp, err := c.sess.do(req)
	if err != nil {
		return []*SeatReservation{}, err
	}
	defer resp.Body.Close()

	var reservations []*SeatReservation
	if err := json.NewDecoder(resp.Body).Decode(&reservations); err != nil {
		return []*SeatReservation{}, err
	}

	return reservations, nil
}

func (c *Client) CancelReservation(ctx context.Context, reservationID string) error {
	uri := fmt.Sprintf("%s%s", c.baseURL, consts.BuildCancelReservationPath(reservationID))

	req, err := c.sess.newRequest(ctx, http.MethodDelete, uri, nil)
	if err != nil {
		return err
	}

	resp, err := c.sess.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		// FIXME: 実装
	}

	return nil
}
