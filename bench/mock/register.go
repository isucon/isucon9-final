package mock

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"runtime"
	"time"

	"github.com/chibiegg/isucon9-final/bench/internal/endpoint"
	"github.com/jarcoal/httpmock"
)

// RegisterMockEndpoints はhttpmockのエンドポイントを登録する
// NOTE: httpmock.Activate, httpmock.Deactivateは別途実施する必要があります
func Register() (*Mock, error) {
	paymentMock := newPaymentMock()
	isutrainMock, err := NewMock(paymentMock)
	if err != nil {
		return nil, err
	}
	isutrainMock.LoginDelay = 100 * time.Millisecond
	isutrainMock.ReserveDelay = 100 * time.Millisecond
	isutrainMock.ListStationsDelay = 100 * time.Millisecond
	isutrainMock.SearchTrainsDelay = 100 * time.Millisecond
	isutrainMock.CommitReservationDelay = 100 * time.Millisecond
	isutrainMock.CancelReservationDelay = 100 * time.Millisecond
	isutrainMock.ListReservationDelay = 100 * time.Millisecond
	isutrainMock.ListTrainSeatsDelay = 100 * time.Millisecond

	baseURL := "http://localhost"
	paymentBaseURL := "http://localhost:5000"

	// GET
	httpmock.RegisterResponder("GET", fmt.Sprintf("%s%s", baseURL, endpoint.GetPath(endpoint.Settings)), func(req *http.Request) (*http.Response, error) {
		b, err := json.Marshal(map[string]interface{}{
			"payment_api": paymentBaseURL,
		})
		if err != nil {
			return httpmock.NewBytesResponse(http.StatusInternalServerError, []byte(http.StatusText(http.StatusInternalServerError))), nil
		}

		return httpmock.NewBytesResponse(http.StatusOK, b), nil
	})
	httpmock.RegisterResponder("GET", fmt.Sprintf("%s%s", baseURL, endpoint.GetPath(endpoint.ListStations)), func(req *http.Request) (*http.Response, error) {
		body, status := isutrainMock.ListStations(req)
		return httpmock.NewBytesResponse(status, body), nil
	})
	httpmock.RegisterResponder("GET", fmt.Sprintf("%s%s", baseURL, endpoint.GetPath(endpoint.SearchTrains)), func(req *http.Request) (*http.Response, error) {
		body, status := isutrainMock.SearchTrains(req)
		return httpmock.NewBytesResponse(status, body), nil
	})
	httpmock.RegisterResponder("GET", fmt.Sprintf("%s%s", baseURL, endpoint.GetPath(endpoint.ListTrainSeats)), func(req *http.Request) (*http.Response, error) {
		body, status := isutrainMock.SearchTrainSeats(req)
		return httpmock.NewBytesResponse(status, body), nil
	})
	httpmock.RegisterResponder("GET", fmt.Sprintf("%s%s", baseURL, endpoint.GetPath(endpoint.ListReservations)), func(req *http.Request) (*http.Response, error) {
		body, status := isutrainMock.ListReservations(req)
		return httpmock.NewBytesResponse(status, body), nil
	})
	httpmock.RegisterResponder("GET", endpoint.IsutrainMockShowReservationPath, func(req *http.Request) (*http.Response, error) {
		body, status := isutrainMock.ShowReservation(req)
		return httpmock.NewBytesResponse(status, body), nil
	})

	// POST
	httpmock.RegisterResponder("POST", fmt.Sprintf("%s%s", baseURL, endpoint.GetPath(endpoint.Initialize)), func(req *http.Request) (*http.Response, error) {
		body, status := isutrainMock.Initialize(req)
		return httpmock.NewBytesResponse(status, body), nil
	})
	httpmock.RegisterResponder("POST", fmt.Sprintf("%s%s", baseURL, endpoint.GetPath(endpoint.Signup)), func(req *http.Request) (*http.Response, error) {
		body, status := isutrainMock.Signup(req)
		return httpmock.NewBytesResponse(status, body), nil
	})
	httpmock.RegisterResponder("POST", fmt.Sprintf("%s%s", baseURL, endpoint.GetPath(endpoint.Login)), func(req *http.Request) (*http.Response, error) {
		wr, status := isutrainMock.Login(req)
		resp := httpmock.NewBytesResponse(status, wr.Body.Bytes())
		resp.Header.Add("X-Test-Header", "hoge")
		for headerKey, header := range wr.Header() {
			for _, headerValue := range header {
				resp.Header.Add(headerKey, headerValue)
			}
		}
		return resp, nil
	})
	httpmock.RegisterResponder("POST", fmt.Sprintf("%s%s", baseURL, endpoint.GetPath(endpoint.Logout)), func(req *http.Request) (*http.Response, error) {
		wr, status := isutrainMock.Logout(req)
		resp := httpmock.NewBytesResponse(status, wr.Body.Bytes())
		for headerKey, header := range wr.Header() {
			for _, headerValue := range header {
				resp.Header.Add(headerKey, headerValue)
			}
		}
		return resp, nil
	})
	httpmock.RegisterResponder("POST", fmt.Sprintf("%s%s", baseURL, endpoint.GetPath(endpoint.Reserve)), func(req *http.Request) (*http.Response, error) {
		body, status := isutrainMock.Reserve(req)
		return httpmock.NewBytesResponse(status, body), nil
	})
	httpmock.RegisterResponder("POST", endpoint.GetPath(endpoint.CommitReservation), func(req *http.Request) (*http.Response, error) {
		body, status := isutrainMock.CommitReservation(req)
		return httpmock.NewBytesResponse(status, body), nil
	})

	// DELETE
	httpmock.RegisterResponder("DELETE", endpoint.IsutrainMockCancelReservationPath, func(req *http.Request) (*http.Response, error) {
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
	httpmock.RegisterResponder("POST", fmt.Sprintf("%s%s", paymentBaseURL, endpoint.PaymentInitializePath), func(req *http.Request) (*http.Response, error) {
		body, status := paymentMock.Initialize()
		return httpmock.NewBytesResponse(status, body), nil
	})
	httpmock.RegisterResponder("GET", fmt.Sprintf("%s%s", paymentBaseURL, endpoint.PaymentResultPath), func(req *http.Request) (*http.Response, error) {
		body, status := paymentMock.GetResult()
		return httpmock.NewBytesResponse(status, body), nil
	})
	httpmock.RegisterResponder("POST", fmt.Sprintf("%s%s", paymentBaseURL, endpoint.PaymentRegistCardPath), func(req *http.Request) (*http.Response, error) {
		body, status := paymentMock.RegistCard()
		return httpmock.NewBytesResponse(status, body), nil
	})

	return isutrainMock, nil
}
