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

func SeasonGoldenWeekScenario(ctx context.Context, goldenweekDate time.Time) error {
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
				return
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

			_, err = createSimpleReservation(ctx, client, user, goldenweekDate, departure, arrival, "最速", 1, 1)
			if err != nil {
				bencherror.BenchmarkErrs.AddError(err)
				totalErr = err
			}
		}()
	}
	wg.Wait()

	return totalErr
}

func reserveForOlympic(ctx context.Context, scenarioIdx int, user *isutrain.User, departure, arrival string) error {
	lgr := zap.S()
	defer lgr.Infof("[season:SeasonOlympicScenario] Done %d", scenarioIdx)

	var (
		useAt        = xrandom.GetRandomUseAtByOlympicDate()
		adult, child = xrandom.GetRandomNumberOfPeople()
	)

	client, err := isutrain.NewClient()
	if err != nil {
		return err
	}

	if config.Debug {
		client.ReplaceMockTransport()
	}

	err = registerUserAndLogin(ctx, client, user)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	_, err = client.ListStations(ctx)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	_, err = createSpecifiedReservation(ctx, client, user, useAt, departure, arrival, adult, child)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	return nil
}

func vagueReserveForOlympic(ctx context.Context, scenarioIdx int, user *isutrain.User, departure, arrival string) error {
	lgr := zap.S()
	defer lgr.Infof("[season:SeasonOlympicScenario] Done %d", scenarioIdx)

	var (
		useAt        = xrandom.GetRandomUseAtByOlympicDate()
		adult, child = xrandom.GetRandomNumberOfPeople()
	)

	client, err := isutrain.NewClient()
	if err != nil {
		return err
	}

	if config.Debug {
		client.ReplaceMockTransport()
	}

	err = registerUserAndLogin(ctx, client, user)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	_, err = client.ListStations(ctx)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	lgr.Infow("指定検索: ",
		"user", user,
		"use_at", useAt,
		"departure", departure,
		"arrival", arrival,
		"adult", adult,
		"child", child,
	)
	_, err = createSimpleReservation(ctx, client, user, useAt, departure, arrival, "遅いやつ", adult, child)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	return nil
}

func SeasonOlympicScenario(ctx context.Context) error {
	// zapロガー取得 (参考: https://qiita.com/emonuh/items/28dbee9bf2fe51d28153#sugaredlogger )
	lgr := zap.S()

	// ここからシナリオ
	lgr.Info("[season:SeasonOlympicScenario] オリンピックのシナリオ")

	eg := &errgroup.Group{}

	// 東京に行きたい人
	for i := 0; i < 3; i++ {
		var (
			scenarioIdx       = i
			tokyo, anyStation = xrandom.GetRandomSectionWithTokyo()
			user, err         = xrandom.GetRandomUser()
		)
		if err != nil {
			return bencherror.SystemErrs.AddError(err)
		}
		eg.Go(func() error {
			return reserveForOlympic(ctx, scenarioIdx, user, anyStation, tokyo)
		})
	}

	// 東京から離れたい人
	for i := 3; i < 5; i++ {
		var (
			scenarioIdx       = i
			tokyo, anyStation = xrandom.GetRandomSectionWithTokyo()
			user, err         = xrandom.GetRandomUser()
		)
		if err != nil {
			return bencherror.SystemErrs.AddError(err)
		}
		eg.Go(func() error {
			return vagueReserveForOlympic(ctx, scenarioIdx, user, tokyo, anyStation)
		})
	}

	return eg.Wait()
}
