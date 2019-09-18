package mock

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"runtime"

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

	// Assets
	_, file, _, _ := runtime.Caller(0)
	testDir := filepath.Join(filepath.Dir(file), "testdata")
	httpmock.RegisterResponder("GET", fmt.Sprintf("%s/css/app.css", baseURL), func(req *http.Request) (*http.Response, error) {
		b, err := ioutil.ReadFile(filepath.Join(testDir, "/css/app.css"))
		if err != nil {
			return httpmock.NewBytesResponse(http.StatusInternalServerError, []byte(http.StatusText(http.StatusInternalServerError))), err
		}
		return httpmock.NewBytesResponse(http.StatusOK, b), nil
	})
	httpmock.RegisterResponder("GET", fmt.Sprintf("%s/img/logo.svg", baseURL), func(req *http.Request) (*http.Response, error) {
		b, err := ioutil.ReadFile(filepath.Join(testDir, "/img/logo.svg"))
		if err != nil {
			return httpmock.NewBytesResponse(http.StatusInternalServerError, []byte(http.StatusText(http.StatusInternalServerError))), err
		}
		return httpmock.NewBytesResponse(http.StatusOK, b), nil
	})
	httpmock.RegisterResponder("GET", fmt.Sprintf("%s/js/app.js", baseURL), func(req *http.Request) (*http.Response, error) {
		b, err := ioutil.ReadFile(filepath.Join(testDir, "/js/app.js"))
		if err != nil {
			return httpmock.NewBytesResponse(http.StatusInternalServerError, []byte(http.StatusText(http.StatusInternalServerError))), err
		}
		return httpmock.NewBytesResponse(http.StatusOK, b), nil
	})
	httpmock.RegisterResponder("GET", fmt.Sprintf("%s/js/chunk.js", baseURL), func(req *http.Request) (*http.Response, error) {
		b, err := ioutil.ReadFile(filepath.Join(testDir, "/js/chunk.js"))
		if err != nil {
			return httpmock.NewBytesResponse(http.StatusInternalServerError, []byte(http.StatusText(http.StatusInternalServerError))), err
		}
		log.Println(string(b))
		return httpmock.NewBytesResponse(http.StatusOK, b), nil
	})
	httpmock.RegisterResponder("GET", fmt.Sprintf("%s/favicon.ico", baseURL), func(req *http.Request) (*http.Response, error) {
		b, err := ioutil.ReadFile(filepath.Join(testDir, "/favicon.ico"))
		if err != nil {
			return httpmock.NewBytesResponse(http.StatusInternalServerError, []byte(http.StatusText(http.StatusInternalServerError))), err
		}
		return httpmock.NewBytesResponse(http.StatusOK, b), nil
	})
	httpmock.RegisterResponder("GET", fmt.Sprintf("%s/index.html", baseURL), func(req *http.Request) (*http.Response, error) {
		b, err := ioutil.ReadFile(filepath.Join(testDir, "/index.html"))
		if err != nil {
			return httpmock.NewBytesResponse(http.StatusInternalServerError, []byte(http.StatusText(http.StatusInternalServerError))), err
		}
		return httpmock.NewBytesResponse(http.StatusOK, b), nil
	})

	return m
}
