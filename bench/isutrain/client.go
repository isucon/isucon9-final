package isutrain

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/chibiegg/isucon9-final/bench/internal/consts"
	"github.com/chibiegg/isucon9-final/bench/util"
)

type Isutrain struct {
	BaseURL string
}

func NewIsutrain(baseURL string) *Isutrain {
	return &Isutrain{
		BaseURL: baseURL,
	}
}

func (i *Isutrain) Register(username, password string) (*http.Response, error) {
	var (
		uri    = fmt.Sprintf("%s%s", i.BaseURL, consts.RegisterPath)
		client http.Client
		form   = url.Values{}
	)
	form.Set("username", username)
	form.Set("password", password)

	req, err := http.NewRequest(http.MethodPost, uri, bytes.NewBufferString(form.Encode()))
	if err != nil {
		return nil, err
	}

	return client.Do(req)
}

func (i *Isutrain) Login(username, password string) (*http.Response, error) {
	var (
		uri    = fmt.Sprintf("%s%s", i.BaseURL, consts.LoginPath)
		client http.Client
		form   = url.Values{}
	)
	form.Set("username", username)
	form.Set("password", password)

	req, err := http.NewRequest(http.MethodPost, uri, bytes.NewBufferString(form.Encode()))
	if err != nil {
		return nil, err
	}

	return client.Do(req)
}

func (i *Isutrain) SearchTrains(useAt time.Time, from, to string) (*http.Response, error) {
	var (
		uri    = fmt.Sprintf("%s%s", i.BaseURL, consts.SearchTrainsPath)
		client http.Client
	)
	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()
	query.Set("use_at", util.FormatISO8601(useAt))
	query.Set("from", from)
	query.Set("to", to)
	req.URL.RawQuery = query.Encode()

	return client.Do(req)
}

func (i *Isutrain) ListTrainSeats(train_class, train_name string) (*http.Response, error) {
	var (
		uri    = fmt.Sprintf("%s%s", i.BaseURL, consts.ListTrainSeatsPath)
		client http.Client
	)
	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()
	query.Set("train_class", train_class)
	query.Set("train_name", train_name)
	req.URL.RawQuery = query.Encode()

	return client.Do(req)
}

func (i *Isutrain) Reserve() (*http.Response, error) {
	var (
		uri    = fmt.Sprintf("%s%s", i.BaseURL, consts.ReservePath)
		client http.Client
		// form = url.Values{}
	)
	req, err := http.NewRequest(http.MethodPost, uri, nil)
	if err != nil {
		return nil, err
	}

	return client.Do(req)
}

func (i *Isutrain) CommitReservation(reservationID int) (*http.Response, error) {
	var (
		uri    = fmt.Sprintf("%s%s", i.BaseURL, consts.BuildCommitReservationPath(reservationID))
		client http.Client
	)
	req, err := http.NewRequest(http.MethodPost, uri, nil)
	if err != nil {
		return nil, err
	}

	return client.Do(req)
}

func (i *Isutrain) ListReservations() (*http.Response, error) {
	var (
		uri    = fmt.Sprintf("%s%s", i.BaseURL, consts.ListReservationsPath)
		client http.Client
	)
	// TODO: クッキー
	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	return client.Do(req)
}

func (i *Isutrain) CancelReservation(reservationID int) (*http.Response, error) {
	var (
		uri = fmt.Sprintf("%s%s", i.BaseURL, consts.BuildCancelReservationPath(reservationID))
		client http.Client
	)
	req, err := http.NewRequest(http.MethodDelete, uri, nil)
	if err != nil {
		return nil, err
	}

	return client.Do(req)
}
