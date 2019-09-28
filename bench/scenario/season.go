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
	"github.com/chibiegg/isucon9-final/bench/payment"
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

func reserveSingle(ctx context.Context, useAt time.Time, departure, arrival string, adult, child int) error {
	// lgr := zap.S()

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

	// 決済サービスのクライアントを作成
	paymentClient, err := payment.NewClient()
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	user, err := xrandom.GetRandomUser()
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	err = client.Signup(ctx, user.Email, user.Password, nil)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	err = client.Login(ctx, user.Email, user.Password, nil)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	_, err = client.ListStations(ctx, nil)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	trains, err := client.SearchTrains(ctx, useAt, departure, arrival, "最速", nil)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	if len(trains) == 0 {
		err := bencherror.NewSimpleCriticalError("列車検索結果が空でした")
		return bencherror.BenchmarkErrs.AddError(err)
	}

	train := trains[0]

	reservation, err := client.Reserve(ctx,
		train.Class, train.Name,
		"premium", isutrain.TrainSeats{},
		departure, arrival, useAt,
		0, adult, child, "", nil)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	cardToken, err := paymentClient.RegistCard(ctx, "11111111", "222", "10/50")
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	err = client.CommitReservation(ctx, reservation.ReservationID, cardToken, nil)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	_, err = client.ListReservations(ctx, nil)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	reservation2, err := client.ShowReservation(ctx, reservation.ReservationID, nil)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	if reservation.ReservationID != reservation2.ReservationID {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	if err := client.Logout(ctx, nil); err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	return nil
}

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
			err := reserveSingle(ctx, GoldenWeekStartDate, departure, arrival, 1, 1)
			if err != nil {
				totalErr = err
			}
		}()
	}
	wg.Wait()

	return totalErr
}
