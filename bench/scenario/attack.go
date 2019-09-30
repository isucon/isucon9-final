package scenario

import (
	"context"
	"math/rand"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	// "go.uber.org/zap"

	"github.com/chibiegg/isucon9-final/bench/internal/bencherror"
	"github.com/chibiegg/isucon9-final/bench/internal/config"
	"github.com/chibiegg/isucon9-final/bench/internal/isutraindb"
	"github.com/chibiegg/isucon9-final/bench/internal/xrandom"
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
				bencherror.BenchmarkErrs.AddError(err)
				return
			}

			if config.Debug {
				client.ReplaceMockTransport()
			}

			user, err := xrandom.GetRandomUser()
			if err != nil {
				bencherror.BenchmarkErrs.AddError(err)
			}
			err = client.Login(ctx, user.Email, user.Password)
			if err != nil {
				bencherror.BenchmarkErrs.AddError(err)
				return
			}

			for {
				select {
				case <-searchTrainCtx.Done():
					return
				default:
					var (
						useAt    = xrandom.GetRandomUseAt()
						from, to = xrandom.GetRandomSection()
					)
					_, err := client.SearchTrains(searchTrainCtx, useAt, from, to, "")
					if err != nil {
						bencherror.BenchmarkErrs.AddError(err)
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
				bencherror.BenchmarkErrs.AddError(err)
				return
			}

			if config.Debug {
				client.ReplaceMockTransport()
			}

			user, err := xrandom.GetRandomUser()
			if err != nil {
				bencherror.BenchmarkErrs.AddError(bencherror.NewCriticalError(err, "ユーザを作成できません. 運営に確認をお願いいたします"))
			}
			err = client.Login(ctx, user.Email, user.Password)
			if err != nil {
				bencherror.BenchmarkErrs.AddError(err)
				return
			}

			for {
				select {
				case <-listTrainSeatsCtx.Done():
					return
				default:
					var (
						useAt              = xrandom.GetRandomUseAt()
						departure, arrival = xrandom.GetRandomSection()
					)
					trains, err := client.SearchTrains(ctx, useAt, departure, arrival, "")
					if err != nil {
						bencherror.BenchmarkErrs.AddError(err)
					}
					if len(trains) == 0 {
						break
					}

					trainIdx := rand.Intn(len(trains))
					train := trains[trainIdx]
					carNum := 8

					_, err = client.ListTrainSeats(listTrainSeatsCtx, useAt, train.Class, train.Name, carNum, train.Departure, train.Arrival)
					if err != nil {
						bencherror.BenchmarkErrs.AddError(err)
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

	// 正常ログイン
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

					err = client.Login(loginCtx, "hoge", "hoge")
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

	// 異常

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

	user, err := xrandom.GetRandomUser()
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	err = registerUserAndLogin(ctx, client, user)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	_, err = client.ListStations(ctx)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	useAt := xrandom.GetRandomUseAt()
	departure, arrival := xrandom.GetRandomSection()
	trains, err := client.SearchTrains(ctx, useAt, departure, arrival, "遅いやつ")
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	if len(trains) == 0 {
		err := bencherror.NewSimpleCriticalError("列車検索結果が空でした")
		return bencherror.BenchmarkErrs.AddError(err)
	}

	trainIdx := rand.Intn(len(trains))
	train := trains[trainIdx]
	carNum := 9
	listTrainSeatsResp, err := client.ListTrainSeats(ctx,
		useAt,
		train.Class, train.Name, carNum, departure, arrival)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	availSeats := filterTrainSeats(listTrainSeatsResp, 2)

	// １回目の予約
	_, err = client.Reserve(ctx,
		train.Class, train.Name,
		isutraindb.GetSeatClass(train.Class, carNum), availSeats,
		departure, arrival, useAt,
		carNum, 1, 1, "")
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	wg := new(sync.WaitGroup)

	var successCount uint64

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := client.Reserve(ctx,
				train.Class, train.Name,
				isutraindb.GetSeatClass(train.Class, carNum), availSeats,
				departure, arrival, useAt,
				carNum, 1, 1, "", isutrain.StatusCodeOpt(http.StatusBadRequest))
			if err != nil {
				return
			}

			atomic.AddUint64(&successCount, 1)
		}()
	}

	wg.Wait()

	return nil
}

// 他人の予約をキャンセルしようとする
// ちゃんと弾けなかったら失格
func AttackReserveForOtherReservation(ctx context.Context) error {
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

	var (
		user1, user1Err = xrandom.GetRandomUser()
		user2, user2Err = xrandom.GetRandomUser()
	)
	if user1Err != nil {
		return bencherror.BenchmarkErrs.AddError(user1Err)
	}
	if user2Err != nil {
		return bencherror.BenchmarkErrs.AddError(user2Err)
	}

	err = registerUserAndLogin(ctx, client, user1)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	useAt := xrandom.GetRandomUseAt()
	departure, arrival := xrandom.GetRandomSection()
	reservation, err := createSimpleReservation(ctx, client, user1, useAt, departure, arrival, "遅いやつ", 1, 1)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	err = client.Logout(ctx)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	// 異なるユーザーでログインする
	err = registerUserAndLogin(ctx, client, user2)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	err = client.CancelReservation(ctx, reservation.ReservationID)
	if err == nil {
		err = bencherror.NewSimpleCriticalError("他のユーザーの予約がキャンセルできました")
		return bencherror.BenchmarkErrs.AddError(err)
	}

	return nil
}
