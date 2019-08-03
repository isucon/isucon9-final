package isutrain

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/chibiegg/isucon9-final/bench/internal/config"
	"github.com/chibiegg/isucon9-final/bench/internal/consts"
	"github.com/chibiegg/isucon9-final/bench/internal/util"
)

// FIXME: Isutrain -> Client
type Isutrain struct {
	BaseURL string
}

func NewIsutrain(baseURL string) *Isutrain {
	return &Isutrain{
		BaseURL: baseURL,
	}
}

func (i *Isutrain) Initialize(ctx context.Context) (*http.Response, error) {
	uri := fmt.Sprintf("%s%s", i.BaseURL, consts.InitializePath)

	ctx, cancel := context.WithTimeout(ctx, config.InitializeTimeout)
	defer cancel()

	req, err := util.NewIsutrainRequest(http.MethodPost, uri, nil)
	if err != nil {
		return nil, err
	}

	return http.DefaultClient.Do(req.WithContext(ctx))
}

func (i *Isutrain) Register(ctx context.Context, username, password string) (*http.Response, error) {
	var (
		uri  = fmt.Sprintf("%s%s", i.BaseURL, consts.RegisterPath)
		form = url.Values{}
	)
	form.Set("username", username)
	form.Set("password", password)

	ctx, cancel := context.WithTimeout(ctx, config.IsutrainAPITimeout)
	defer cancel()

	req, err := util.NewIsutrainRequest(http.MethodPost, uri, bytes.NewBufferString(form.Encode()))
	if err != nil {
		return nil, err
	}

	return http.DefaultClient.Do(req.WithContext(ctx))
}

func (i *Isutrain) Login(ctx context.Context, username, password string) (*http.Response, error) {
	var (
		uri  = fmt.Sprintf("%s%s", i.BaseURL, consts.LoginPath)
		form = url.Values{}
	)
	form.Set("username", username)
	form.Set("password", password)

	ctx, cancel := context.WithTimeout(ctx, config.IsutrainAPITimeout)
	defer cancel()

	req, err := util.NewIsutrainRequest(http.MethodPost, uri, bytes.NewBufferString(form.Encode()))
	if err != nil {
		return nil, err
	}

	return http.DefaultClient.Do(req.WithContext(ctx))
}

// SearchTrains は 列車検索APIです
func (i *Isutrain) SearchTrains(ctx context.Context, useAt time.Time, from, to string) (*http.Response, error) {
	uri := fmt.Sprintf("%s%s", i.BaseURL, consts.SearchTrainsPath)

	ctx, cancel := context.WithTimeout(ctx, config.IsutrainAPITimeout)
	defer cancel()

	req, err := util.NewIsutrainRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()
	query.Set("use_at", util.FormatISO8601(useAt))
	query.Set("train_class", "") // FIXME: 列車種別
	query.Set("from", from)
	query.Set("to", to)
	req.URL.RawQuery = query.Encode()

	return http.DefaultClient.Do(req.WithContext(ctx))
}

func (i *Isutrain) ListTrainSeats(ctx context.Context, train_class, train_name string) (*http.Response, error) {
	uri := fmt.Sprintf("%s%s", i.BaseURL, consts.ListTrainSeatsPath)

	ctx, cancel := context.WithTimeout(ctx, config.IsutrainAPITimeout)
	defer cancel()

	req, err := util.NewIsutrainRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()
	query.Set("train_class", train_class)
	query.Set("train_name", train_name)
	req.URL.RawQuery = query.Encode()

	return http.DefaultClient.Do(req.WithContext(ctx))
}

func (i *Isutrain) Reserve(ctx context.Context) (*http.Response, error) {
	var (
		uri = fmt.Sprintf("%s%s", i.BaseURL, consts.ReservePath)
		// form = url.Values{}
	)

	ctx, cancel := context.WithTimeout(ctx, config.IsutrainAPITimeout)
	defer cancel()

	req, err := util.NewIsutrainRequest(http.MethodPost, uri, nil)
	if err != nil {
		return nil, err
	}

	return http.DefaultClient.Do(req.WithContext(ctx))
}

func (i *Isutrain) CommitReservation(ctx context.Context, reservationID int) (*http.Response, error) {
	uri := fmt.Sprintf("%s%s", i.BaseURL, consts.BuildCommitReservationPath(reservationID))

	ctx, cancel := context.WithTimeout(ctx, config.IsutrainAPITimeout)
	defer cancel()

	req, err := http.NewRequest(http.MethodPost, uri, nil)
	if err != nil {
		return nil, err
	}

	return http.DefaultClient.Do(req.WithContext(ctx))
}

func (i *Isutrain) ListReservations(ctx context.Context) (*http.Response, error) {
	uri := fmt.Sprintf("%s%s", i.BaseURL, consts.ListReservationsPath)

	ctx, cancel := context.WithTimeout(ctx, config.IsutrainAPITimeout)
	defer cancel()

	// TODO: クッキー
	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	return http.DefaultClient.Do(req.WithContext(ctx))
}

func (i *Isutrain) CancelReservation(ctx context.Context, reservationID int) (*http.Response, error) {
	uri := fmt.Sprintf("%s%s", i.BaseURL, consts.BuildCancelReservationPath(reservationID))

	ctx, cancel := context.WithTimeout(ctx, config.IsutrainAPITimeout)
	defer cancel()

	req, err := http.NewRequest(http.MethodDelete, uri, nil)
	if err != nil {
		return nil, err
	}

	return http.DefaultClient.Do(req.WithContext(ctx))
}
