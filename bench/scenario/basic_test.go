package scenario

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/chibiegg/isucon9-final/bench/internal/consts"
	"github.com/chibiegg/isucon9-final/bench/internal/score"
	"github.com/chibiegg/isucon9-final/bench/internal/util"
	"github.com/chibiegg/isucon9-final/bench/isutrain"
	"github.com/chibiegg/isucon9-final/bench/mock"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestScore(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	isutrain.RegisterMockEndpoints()

	scenario := NewBasicScenario("http://localhost")
	go scenario.Run(context.Background())

	var (
		scores = []int{
			score.RegisterType.Score(),
			score.LoginType.Score(),
			score.SearchTrainsType.Score(),
			score.ListTrainSeatsType.Score(),
			score.ReserveType.Score(),
			score.CommitReservationType.Score(),
			score.ListReservationsType.Score(),
		}
		got, want int
		wg        sync.WaitGroup
	)
	wg.Add(1)
	go func() {
		defer wg.Done()

		var idx int
		for requestResult := range scenario.RequestResultStream() {
			got += requestResult.Type.Score()
			want += scores[idx]

			idx++
		}
	}()

	wg.Wait()
	assert.Equal(t, want, got)
}

func TestScenario(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	var (
		client http.Client
		query  url.Values
		form   url.Values
	)

	m := mock.Register()

	// 初期化する
	req, err := http.NewRequest(http.MethodPost, consts.InitializePath, nil)
	assert.NoError(t, err)

	_, err = client.Do(req)
	assert.NoError(t, err)

	// 登録する
	form = url.Values{}
	form.Set("username", "hoge")
	form.Set("password", "hoge")
	req, err = http.NewRequest(http.MethodPost, consts.RegisterPath, bytes.NewBufferString(form.Encode()))
	assert.NoError(t, err)

	_, err = client.Do(req)
	assert.NoError(t, err)

	// ログインする
	req, err = http.NewRequest(http.MethodPost, consts.LoginPath, bytes.NewBufferString(form.Encode()))
	assert.NoError(t, err)

	_, err = client.Do(req)
	assert.NoError(t, err)

	// 列車を検索
	req, err = http.NewRequest(http.MethodGet, consts.SearchTrainsPath, nil)
	assert.NoError(t, err)

	query = req.URL.Query()
	query.Set("use_at", util.FormatISO8601(time.Now()))
	query.Set("from", "東京")
	query.Set("to", "大阪")
	req.URL.RawQuery = query.Encode()

	_, err = client.Do(req)
	assert.NoError(t, err)

	// どれか選んで座席を列挙
	req, err = http.NewRequest(http.MethodGet, consts.ListTrainSeatsPath, nil)
	assert.NoError(t, err)

	query = req.URL.Query()
	query.Set("train_class", "のぞみ")
	query.Set("train_name", "96号")
	req.URL.RawQuery = query.Encode()

	_, err = client.Do(req)
	assert.NoError(t, err)

	// 座席を選び、予約
	form = url.Values{}
	req, err = http.NewRequest(http.MethodPost, consts.ReservePath, bytes.NewBufferString(form.Encode()))
	assert.NoError(t, err)

	_, err = client.Do(req)
	assert.NoError(t, err)

	// 予約を確定
	req, err = http.NewRequest(http.MethodPost, consts.BuildCommitReservationPath(1111), nil)
	assert.NoError(t, err)

	_, err = client.Do(req)
	assert.NoError(t, err)

	// 予約を確認
	req, err = http.NewRequest(http.MethodGet, consts.ListReservationsPath, nil)
	assert.NoError(t, err)

	resp, err := client.Do(req)
	assert.NoError(t, err)

	body, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	resp.Body.Close()

	var reservations SeatReservations
	assert.NoError(t, json.Unmarshal(body, &reservations))
	assert.Len(t, reservations, 1)

	// 予約をキャンセル
	req, err = http.NewRequest(http.MethodDelete, consts.BuildCancelReservationPath(1111), nil)
	assert.NoError(t, err)

	_, err = client.Do(req)
	assert.NoError(t, err)
}
