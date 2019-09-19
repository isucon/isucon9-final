package mock

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"runtime"
	"time"

	"github.com/chibiegg/isucon9-final/bench/internal/config"
	"github.com/jarcoal/httpmock"
)

// RegisterMockEndpoints はhttpmockのエンドポイントを登録する
// NOTE: httpmock.Activate, httpmock.Deactivateは別途実施する必要があります
func Register() *Mock {
	paymentMock := newPaymentMock()
	isutrainMock := NewMock(paymentMock)
	isutrainMock.LoginDelay = 100 * time.Millisecond
	isutrainMock.ReserveDelay = 100 * time.Millisecond
	isutrainMock.ListStationsDelay = 5000 * time.Millisecond
	isutrainMock.SearchTrainsDelay = 100 * time.Millisecond
	isutrainMock.CommitReservationDelay = 100 * time.Millisecond
	isutrainMock.CancelReservationDelay = 100 * time.Millisecond
	isutrainMock.ListReservationDelay = 100 * time.Millisecond
	isutrainMock.ListTrainSeatsDelay = 100 * time.Millisecond

	baseURL := "http://localhost"
	paymentBaseURL := "http://localhost:5000"

	// GET
	httpmock.RegisterResponder("GET", fmt.Sprintf("%s%s", baseURL, config.IsutrainListStationsPath), func(req *http.Request) (*http.Response, error) {
		body, status := isutrainMock.ListStations(req)
		return httpmock.NewBytesResponse(status, body), nil
	})
	httpmock.RegisterResponder("GET", fmt.Sprintf("%s%s", baseURL, config.IsutrainSearchTrainsPath), func(req *http.Request) (*http.Response, error) {
		body, status := isutrainMock.SearchTrains(req)
		return httpmock.NewBytesResponse(status, body), nil
	})
	httpmock.RegisterResponder("GET", fmt.Sprintf("%s%s", baseURL, config.IsutrainListTrainSeatsPath), func(req *http.Request) (*http.Response, error) {
		body, status := isutrainMock.ListTrainSeats(req)
		return httpmock.NewBytesResponse(status, body), nil
	})
	httpmock.RegisterResponder("GET", fmt.Sprintf("%s%s", baseURL, config.IsutrainListReservationsPath), func(req *http.Request) (*http.Response, error) {
		body, status := isutrainMock.ListReservations(req)
		return httpmock.NewBytesResponse(status, body), nil
	})

	// POST
	httpmock.RegisterResponder("POST", fmt.Sprintf("%s%s", baseURL, config.IsutrainInitializePath), func(req *http.Request) (*http.Response, error) {
		body, status := isutrainMock.Initialize(req)
		return httpmock.NewBytesResponse(status, body), nil
	})
	httpmock.RegisterResponder("POST", fmt.Sprintf("%s%s", baseURL, config.IsutrainRegisterPath), func(req *http.Request) (*http.Response, error) {
		body, status := isutrainMock.Register(req)
		return httpmock.NewBytesResponse(status, body), nil
	})
	httpmock.RegisterResponder("POST", fmt.Sprintf("%s%s", baseURL, config.IsutrainLoginPath), func(req *http.Request) (*http.Response, error) {
		body, status := isutrainMock.Login(req)
		return httpmock.NewBytesResponse(status, body), nil
	})
	httpmock.RegisterResponder("POST", fmt.Sprintf("%s%s", baseURL, config.IsutrainReservePath), func(req *http.Request) (*http.Response, error) {
		body, status := isutrainMock.Reserve(req)
		return httpmock.NewBytesResponse(status, body), nil
	})
	httpmock.RegisterResponder("POST", config.IsutrainMockCommitReservationPath, func(req *http.Request) (*http.Response, error) {
		body, status := isutrainMock.CommitReservation(req)
		return httpmock.NewBytesResponse(status, body), nil
	})

	// DELETE
	httpmock.RegisterResponder("DELETE", config.IsutrainMockCancelReservationPath, func(req *http.Request) (*http.Response, error) {
		body, status := isutrainMock.CancelReservation(req)
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

	// 課金API
	httpmock.RegisterResponder("POST", fmt.Sprintf("%s/initialize", paymentBaseURL), func(req *http.Request) (*http.Response, error) {
		body, status := paymentMock.Initialize()
		return httpmock.NewBytesResponse(status, body), nil
	})
	httpmock.RegisterResponder("GET", fmt.Sprintf("%s/result", paymentBaseURL), func(req *http.Request) (*http.Response, error) {
		body, status := paymentMock.GetResult()
		return httpmock.NewBytesResponse(status, body), nil
	})

	return isutrainMock
}
