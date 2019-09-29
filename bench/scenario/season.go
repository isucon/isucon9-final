package scenario

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/chibiegg/isucon9-final/bench/internal/bencherror"
	"github.com/chibiegg/isucon9-final/bench/internal/config"
	"github.com/chibiegg/isucon9-final/bench/internal/xrandom"
	"github.com/chibiegg/isucon9-final/bench/isutrain"
)

/*
 繁忙期

 GW 2020-04-29 to 2020-05-06
 五輪 2020-07-23 to 2020-08-10

*/

var (
	GoldenWeekStartDate time.Time = time.Date(2020, 4, 29, 0, 0, 0, 0, time.Local)
	GoldenWeekEndDate   time.Time = time.Date(2020, 5, 6, 15, 0, 0, 0, time.Local)
)

func GoldenWeekScenario(ctx context.Context) error {
	// zapロガー取得 (参考: https://qiita.com/emonuh/items/28dbee9bf2fe51d28153#sugaredlogger )
	lgr := zap.S()

	// ISUTRAIN APIのクライアントを作成
	client, err := isutrain.NewClient()
	if err != nil {
		// 実行中のエラーは `bencherror.BenchmarkErrs.AddError(err)` に投げる
		return bencherror.BenchmarkErrs.AddError(err)
	}

	// デバッグの場合はモックに差し替える
	// NOTE: httpmockというライブラリが、http.Transporterを差し替えてエンドポイントをマウントする都合上、この処理が必要です
	//       この処理がないと、テスト実行時に存在しない宛先にリクエストを送り、失敗します
	if config.Debug {
		client.ReplaceMockTransport()
	}

	// ここからシナリオ
	lgr.Info("[season:GoldenWeekScenario] GWのシナリオ")

	var wg sync.WaitGroup
	var totalErr error

	departure, arrival := xrandom.GetTokaiRandomSection()

	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer lgr.Infof("[season:GoldenWeekScenario] Done %d", i)

			client, err := isutrain.NewClient()
			if err != nil {
				bencherror.BenchmarkErrs.AddError(err)
				return
			}

			if config.Debug {
				client.ReplaceMockTransport()
			}

			user, err := xrandom.GetRandomUser()
			if err != nil {
				bencherror.BenchmarkErrs.AddError(err)
				return
			}

			err = registerUserAndLogin(ctx, client, user)
			if err != nil {
				bencherror.BenchmarkErrs.AddError(err)
				return
			}

			_, err = client.ListStations(ctx, nil)
			if err != nil {
				bencherror.BenchmarkErrs.AddError(err)
				return
			}

			_, err = createSimpleReservation(ctx, client, GoldenWeekStartDate, departure, arrival, "最速", 1, 1)
			if err != nil {
				bencherror.BenchmarkErrs.AddError(err)
				totalErr = err
			}
		}()
	}
	wg.Wait()

	return totalErr
}
