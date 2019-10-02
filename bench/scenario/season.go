package scenario

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

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

func SeasonGoldenWeekScenario(ctx context.Context) error {
	// zapロガー取得 (参考: https://qiita.com/emonuh/items/28dbee9bf2fe51d28153#sugaredlogger )
	lgr := zap.S()

	// ここからシナリオ
	lgr.Info("[season:GoldenWeekScenario] GWのシナリオ")

	var wg sync.WaitGroup
	var totalErr error

	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer lgr.Infof("[season:GoldenWeekScenario] Done %d", i)

			departure, arrival := xrandom.GetTokaiRandomSection()

			client, err := isutrain.NewClient()
			if err != nil {
				bencherror.BenchmarkErrs.AddError(err)
				return
			}

			if config.Debug {
				client.ReplaceMockTransport()
			}

			if config.Debug {
				client.ReplaceMockTransport()
			}

			user, err := xrandom.GetRandomUser()
			if err != nil {
				bencherror.SystemErrs.AddError(err)
				return
			}

			err = registerUserAndLogin(ctx, client, user)
			if err != nil {
				bencherror.BenchmarkErrs.AddError(err)
				return
			}

			_, err = client.ListStations(ctx)
			if err != nil {
				bencherror.BenchmarkErrs.AddError(err)
				return
			}

			_, err = createSimpleReservation(ctx, client, user, GoldenWeekStartDate, departure, arrival, "最速", 1, 1)
			if err != nil {
				bencherror.BenchmarkErrs.AddError(err)
				totalErr = err
			}
		}()
	}
	wg.Wait()

	return totalErr
}

func SeasonOlympicScenario(ctx context.Context) error {
	// zapロガー取得 (参考: https://qiita.com/emonuh/items/28dbee9bf2fe51d28153#sugaredlogger )
	lgr := zap.S()

	// ここからシナリオ
	lgr.Info("[season:SeasonOlympicScenario] オリンピックのシナリオ")

	eg := &errgroup.Group{}
	reserveForOlympic := func(scenarioIdx int, departure, arrival string) error {
		defer lgr.Infof("[season:SeasonOlympicScenario] Done %d", scenarioIdx)

		var (
			useAt        = xrandom.GetRandomUseAtByOlympicDate()
			adult, child = xrandom.GetRandomNumberOfPeople()
		)

		client, err := isutrain.NewClient()
		if err != nil {
			bencherror.BenchmarkErrs.AddError(err)
			return nil
		}

		if config.Debug {
			client.ReplaceMockTransport()
		}

		user, err := xrandom.GetRandomUser()
		if err != nil {
			bencherror.SystemErrs.AddError(err)
			return nil
		}

		err = registerUserAndLogin(ctx, client, user)
		if err != nil {
			bencherror.BenchmarkErrs.AddError(err)
			return nil
		}

		_, err = client.ListStations(ctx)
		if err != nil {
			bencherror.BenchmarkErrs.AddError(err)
			return nil
		}

		_, err = createSimpleReservation(ctx, client, user, useAt, departure, arrival, "最速", adult, child)
		if err != nil {
			return bencherror.BenchmarkErrs.AddError(err)
		}

		return nil
	}

	// 東京に行きたい人
	for i := 0; i < 3; i++ {
		var (
			scenarioIdx       = i
			tokyo, anyStation = xrandom.GetRandomSectionWithTokyo()
		)
		eg.Go(func() error {
			return reserveForOlympic(scenarioIdx, anyStation, tokyo)
		})
	}

	// 東京から離れたい人
	for i := 3; i < 5; i++ {
		var (
			scenarioIdx       = i
			tokyo, anyStation = xrandom.GetRandomSectionWithTokyo()
		)
		eg.Go(func() error {
			return reserveForOlympic(scenarioIdx, tokyo, anyStation)
		})
	}

	return eg.Wait()
}
