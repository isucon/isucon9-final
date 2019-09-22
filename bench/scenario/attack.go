package scenario

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/chibiegg/isucon9-final/bench/internal/bencherror"
	"github.com/chibiegg/isucon9-final/bench/internal/config"
	"github.com/chibiegg/isucon9-final/bench/isutrain"
)

// FIXME: 適当に10個生成するようにしてるけど、設定できるように

// 検索しまくる
func AttackSearchScenario(ctx context.Context) error {
	var searchGrp sync.WaitGroup

	// SearchTrains
	searchTrainCtx, cancelSearchTrain := context.WithTimeout(ctx, config.AttackSearchTrainTimeout)
	defer cancelSearchTrain()
	for i := 0; i < 10; i++ {
		searchGrp.Add(1)
		go func() {
			defer searchGrp.Done()

			client, err := isutrain.NewClient()
			if err != nil {
				bencherror.BenchmarkErrs.AddError(bencherror.NewCriticalError(err, "Isutrainクライアントが作成できません. 運営に確認をお願いいたします"))
				return
			}

			if config.Debug {
				client.ReplaceMockTransport()
			}

			err = client.Login(ctx, "hoge", "hoge", nil)
			if err != nil {
				bencherror.BenchmarkErrs.AddError(bencherror.NewApplicationError(err, "ユーザログインができません"))
				return
			}

			for {
				select {
				case <-searchTrainCtx.Done():
					return
				default:
					_, err := client.SearchTrains(
						searchTrainCtx,
						time.Now(),
						"東京",
						"名古屋",
						nil,
					)
					if err != nil {
						bencherror.BenchmarkErrs.AddError(bencherror.NewApplicationError(err, "列車検索に失敗しました"))
						return
					}
				}
			}
		}()
	}

	// ListTrainSeats
	listTrainSeatsCtx, cancelListTrainSeats := context.WithTimeout(ctx, config.AttackListTrainSeatsTimeout)
	defer cancelListTrainSeats()
	for i := 0; i < 10; i++ {
		searchGrp.Add(1)
		go func() {
			defer searchGrp.Done()

			client, err := isutrain.NewClient()
			if err != nil {
				bencherror.BenchmarkErrs.AddError(bencherror.NewCriticalError(err, "Isutrainクライアントが作成できません. 運営に確認をお願いいたします"))
				return
			}

			if config.Debug {
				client.ReplaceMockTransport()
			}

			err = client.Login(ctx, "hoge", "hoge", nil)
			if err != nil {
				bencherror.BenchmarkErrs.AddError(bencherror.NewApplicationError(err, "ユーザログインができません"))
				return
			}

			for {
				select {
				case <-listTrainSeatsCtx.Done():
					return
				default:
					_, err := client.ListTrainSeats(
						listTrainSeatsCtx,
						"ほげ",
						"ほげ",
						1,
						"東京",
						"名古屋",
						nil,
					)
					if err != nil {
						bencherror.BenchmarkErrs.AddError(bencherror.NewApplicationError(err, "列車の座席列挙に失敗しました"))
						return
					}
				}
			}
		}()
	}

	searchGrp.Wait()
	return nil
}

// ログインしまくる (ログイン失敗もする. また、失敗するはずが成功したりしたら失格扱いにする)
func AttackLoginScenario(ctx context.Context) error {
	var loginGrp sync.WaitGroup

	loginCtx, cancelLogin := context.WithTimeout(ctx, 20*time.Second)
	defer cancelLogin()
	for i := 0; i < 10; i++ {
		loginGrp.Add(1)
		go func() {
			defer loginGrp.Done()

			for {
				select {
				case <-loginCtx.Done():
					return
				default:
					// TODO: リソースリークしないかチェック
					client, err := isutrain.NewClient()
					if err != nil {
						bencherror.BenchmarkErrs.AddError(bencherror.NewCriticalError(err, "Isutrainクライアントが作成できません. 運営に確認をお願いいたします"))
						return
					}

					if config.Debug {
						client.ReplaceMockTransport()
					}

					err = client.Login(loginCtx, "hoge", "hoge", nil)
					if err != nil {
						bencherror.BenchmarkErrs.AddError(bencherror.NewApplicationError(err, "ユーザログインができません"))
						return
					}

					msecs := rand.Intn(1000)
					time.Sleep(time.Duration(msecs) * time.Millisecond)
				}
			}
		}()
	}

	loginGrp.Wait()
	return nil
}

// 予約済みユーザについて、予約確認しまくる
// FIXME: 予約済みユーザを取ってくる仕組みづくりが必要
func AttackListReservationsScenario(ctx context.Context) error {
	return nil
}

// 予約済みの条件で予約を試みる
// 一応、予約キャンセルするのを虎視眈々と狙っている利用者からのリクエスト、という設定
func AttackReserveForReserved(ctx context.Context) error {
	return nil
}

// 他人の予約をキャンセルしようとする
// ちゃんと弾けなかったら失格
func AttackReserveForOtherReservation(ctx context.Context) error {
	return nil
}
