package mock

import (
	"fmt"
	"net/http"

	"github.com/chibiegg/isucon9-final/bench/internal/config"
	"github.com/jarcoal/httpmock"
)

// RegisterMockEndpoints はhttpmockのエンドポイントを登録する
// NOTE: httpmock.Activate, httpmock.Deactivateは別途実施する必要があります
func Register() *Mock {
	m := NewMock()
	baseURL := "http://localhost"

	// GET
	httpmock.RegisterResponder("GET", fmt.Sprintf("%s%s", baseURL, config.ListStationsPath), func(req *http.Request) (*http.Response, error) {
		body, status := m.ListStations(req)
		return httpmock.NewBytesResponse(status, body), nil
	})
	httpmock.RegisterResponder("GET", fmt.Sprintf("%s%s", baseURL, config.SearchTrainsPath), func(req *http.Request) (*http.Response, error) {
		body, status := m.SearchTrains(req)
		return httpmock.NewBytesResponse(status, body), nil
	})
	httpmock.RegisterResponder("GET", fmt.Sprintf("%s%s", baseURL, config.ListTrainSeatsPath), func(req *http.Request) (*http.Response, error) {
		body, status := m.ListTrainSeats(req)
		return httpmock.NewBytesResponse(status, body), nil
	})
	httpmock.RegisterResponder("GET", fmt.Sprintf("%s%s", baseURL, config.ListReservationsPath), func(req *http.Request) (*http.Response, error) {
		body, status := m.ListReservations(req)
		return httpmock.NewBytesResponse(status, body), nil
	})

	// POST
	httpmock.RegisterResponder("POST", fmt.Sprintf("%s%s", baseURL, config.InitializePath), func(req *http.Request) (*http.Response, error) {
		body, status := m.Initialize(req)
		return httpmock.NewBytesResponse(status, body), nil
	})
	httpmock.RegisterResponder("POST", fmt.Sprintf("%s%s", baseURL, config.RegisterPath), func(req *http.Request) (*http.Response, error) {
		body, status := m.Register(req)
		return httpmock.NewBytesResponse(status, body), nil
	})
	httpmock.RegisterResponder("POST", fmt.Sprintf("%s%s", baseURL, config.LoginPath), func(req *http.Request) (*http.Response, error) {
		body, status := m.Login(req)
		return httpmock.NewBytesResponse(status, body), nil
	})
	httpmock.RegisterResponder("POST", fmt.Sprintf("%s%s", baseURL, config.ReservePath), func(req *http.Request) (*http.Response, error) {
		body, status := m.Reserve(req)
		return httpmock.NewBytesResponse(status, body), nil
	})
	httpmock.RegisterResponder("POST", config.MockCommitReservationPath, func(req *http.Request) (*http.Response, error) {
		body, status := m.CommitReservation(req)
		return httpmock.NewBytesResponse(status, body), nil
	})

	// DELETE
	httpmock.RegisterResponder("DELETE", config.MockCancelReservationPath, func(req *http.Request) (*http.Response, error) {
		body, status := m.CancelReservation(req)
		return httpmock.NewBytesResponse(status, body), nil
	})

	return m
}
