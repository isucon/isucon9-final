package scenario

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"reflect"
	"time"

	"github.com/chibiegg/isucon9-final/bench/assets"
	"github.com/chibiegg/isucon9-final/bench/internal/bencherror"
	"github.com/chibiegg/isucon9-final/bench/internal/config"
	"github.com/chibiegg/isucon9-final/bench/internal/endpoint"
	"github.com/chibiegg/isucon9-final/bench/isutrain"
	"github.com/chibiegg/isucon9-final/bench/payment"
	"golang.org/x/sync/errgroup"
)

// Guest訪問
var (
	ErrInitialTrainDatasetCount = errors.New("列車初期データセットの件数が一致しません")
	ErrInvalidAssetHash         = errors.New("静的ファイルの内容が正しくありません")
)

// Pretest は、ベンチマーク前のアプリケーションが正常に動作できているか検証し、できていなければFAILとします
func Pretest(ctx context.Context, client *isutrain.Client, paymentClient *payment.Client, assets []*assets.Asset) {
	// 正常 - 取得系
	getGrp := &errgroup.Group{}
	getGrp.Go(func() error {
		return pretestStaticFiles(ctx, client, assets)
	})
	getGrp.Go(func() error {
		return pretestListStations(ctx, client)
	})
	getGrp.Go(func() error {
		return pretestSearchTrains(ctx, client)
	})
	getGrp.Go(func() error {
		return pretestSearchTrainSeats(ctx, client)
	})
	if err := getGrp.Wait(); err != nil {
		return
	}

	// 異常系
	abnormalGrp := &errgroup.Group{}
	abnormalGrp.Go(func() error {
		return pretestAbnormalLogin(ctx, client)
	})
	abnormalGrp.Go(func() error {
		return pretestAbnormalDate(ctx, client)
	})
	abnormalGrp.Go(func() error {
		return pretestAbnormalReserveNonStoppableStation(ctx, client)
	})
	abnormalGrp.Go(func() error {
		return pretestAbnormalReserveInvalidSection(ctx, client)
	})
	if err := abnormalGrp.Wait(); err != nil {
		return
	}

	// 予約状態を変更するpretest
	if err := pretestNormalReservation(ctx, client, paymentClient); err != nil {
		return
	}
	if err := pretestReserveVagueSeats(ctx, client); err != nil {
		return
	}
}

// 静的ファイルチェック
func pretestStaticFiles(ctx context.Context, client *isutrain.Client, assets []*assets.Asset) error {
	for _, asset := range assets {
		b, err := client.DownloadAsset(ctx, asset.Path)
		if err != nil {
			return err
		}

		hash := sha256.Sum256(b)

		if !bytes.Equal(hash[:], asset.Hash[:]) {
			return bencherror.PreTestErrs.AddError(bencherror.NewApplicationError(ErrInvalidAssetHash, "GET /%s: 静的ファイルのハッシュ値が異なります", asset.Path))
		}
	}
	return nil
}

// 正常系

// 駅一覧
func pretestListStations(ctx context.Context, client *isutrain.Client) error {
	endpointPath := endpoint.GetPath(endpoint.ListStations)

	client, err := isutrain.NewClient()
	if err != nil {
		return bencherror.PreTestErrs.AddError(err)
	}

	err = registerUserAndLogin(ctx, client, &isutrain.User{
		Email:    "puser1@example.com",
		Password: "puser1",
	})
	if err != nil {
		return bencherror.PreTestErrs.AddError(err)
	}

	gotStations, err := client.ListStations(ctx)
	if err != nil {
		return bencherror.PreTestErrs.AddError(err)
	}

	if ok := isutrain.IsValidStations(gotStations); !ok {
		return bencherror.PreTestErrs.AddError(bencherror.NewSimpleCriticalError("GET %s: 駅一覧が不正です", endpointPath))
	}

	return nil
}

// 列車検索

func pretestSearchTrains(ctx context.Context, client *isutrain.Client) error {
	// 初期状態で、いくつか試す
	// 必ずこれは空にならないというパターンを試す
	endpointPath := endpoint.GetPath(endpoint.SearchTrains)

	err := registerUserAndLogin(ctx, client, &isutrain.User{
		Email:    "puser2@example.com",
		Password: "puser2",
	})
	if err != nil {
		return bencherror.PreTestErrs.AddError(err)
	}

	randIdx := rand.Intn(len(pretestSearchTrainsTests))
	randTest := pretestSearchTrainsTests[randIdx]

	resp, err := client.SearchTrains(ctx, randTest.useAt, randTest.from, randTest.to, randTest.trainClass, randTest.adult, randTest.child)
	if err != nil {
		return bencherror.PreTestErrs.AddError(err)
	}

	if len(resp) == 0 {
		bencherror.PreTestErrs.AddError(bencherror.NewSimpleCriticalError("GET %s: 検索結果を返すべき条件で返せておらず、空です", endpointPath))
	}

	for _, train := range resp {
		wantSeatAvailability, ok := randTest.wantSeatAvailability[train.Name]
		if !ok {
			return bencherror.PreTestErrs.AddError(bencherror.NewSimpleCriticalError("GET %s: 検索条件に対し、不正な列車名がレスポンスに含まれています: got=%s", endpointPath, train.Name))
		}
		if !reflect.DeepEqual(train.SeatAvailability, wantSeatAvailability) {
			return bencherror.PreTestErrs.AddError(bencherror.NewSimpleCriticalError("GET %s: seat_availabilityが不正です: want=%+v, got=%v", endpointPath, wantSeatAvailability, train.SeatAvailability))
		}

		wantFareInformation := randTest.wantFareInformation[train.Name]
		if !reflect.DeepEqual(train.FareInformation, wantFareInformation) {
			return bencherror.PreTestErrs.AddError(bencherror.NewSimpleCriticalError("GET %s: fare_informationが不正です: want=%+v, got=%v", endpointPath, wantFareInformation, train.FareInformation))
		}

		wantDepartedAt := randTest.wantDepartedAt[train.Name]
		if train.DepartedAt != wantDepartedAt {
			return bencherror.PreTestErrs.AddError(bencherror.NewSimpleCriticalError("GET %s: DepartedAtが不正です: want=%s, got=%s", endpointPath, wantDepartedAt, train.DepartedAt))
		}

		wantArrivedAt := randTest.wantArrivedAt[train.Name]
		if train.ArrivedAt != wantArrivedAt {
			return bencherror.PreTestErrs.AddError(bencherror.NewSimpleCriticalError("GET %s: ArrivedAtが不正です: want=%s, got=%s", endpointPath, wantArrivedAt, train.ArrivedAt))
		}
	}

	return nil
}

func pretestSearchTrainsEmptyTrainClass(ctx context.Context, client *isutrain.Client) {
	// 空のtrain_classを設定してリクエストした場合にも検索ができるはず
	// https://github.com/chibiegg/isucon9-final/blob/master/webapp/go/main.go#L507
}

// 座席検索

func pretestSearchTrainSeats(ctx context.Context, client *isutrain.Client) error {
	endpointPath := endpoint.GetPath(endpoint.SearchTrains)

	err := registerUserAndLogin(ctx, client, &isutrain.User{
		Email:    "puser4@example.com",
		Password: "puser4",
	})
	if err != nil {
		return bencherror.PreTestErrs.AddError(err)
	}

	randIdx := rand.Intn(len(pretestSearchTrainSeatsTests))
	randTest := pretestSearchTrainSeatsTests[randIdx]

	resp, err := client.SearchTrainSeats(ctx,
		randTest.date, randTest.trainClass, randTest.trainName,
		randTest.carNum, randTest.departure, randTest.arrival)
	if err != nil {
		return bencherror.PreTestErrs.AddError(err)
	}

	wantDate := randTest.date.Format("2006/01/02")
	if resp.Date != wantDate {
		return bencherror.PreTestErrs.AddError(bencherror.NewSimpleCriticalError("GET %s: レスポンスに含まれる日付が不正です: want=%s, got=%s", endpointPath, wantDate, resp.Date))
	}
	if resp.TrainClass != randTest.trainClass {
		return bencherror.PreTestErrs.AddError(bencherror.NewSimpleCriticalError("GET %s: レスポンスに含まれる列車種別が不正です: want=%s, got=%s", endpointPath, randTest.trainClass, resp.TrainClass))
	}
	if resp.TrainName != randTest.trainName {
		return bencherror.PreTestErrs.AddError(bencherror.NewSimpleCriticalError("GET %s: レスポンスに含まれる列車名が不正です: want=%s, got=%s", endpointPath, randTest.trainName, resp.TrainName))
	}
	if resp.CarNumber != randTest.carNum {
		return bencherror.PreTestErrs.AddError(bencherror.NewSimpleCriticalError("GET %s: レスポンスに含まれる車両番号が不正です: want=%s, got=%s", endpointPath, randTest.carNum, resp.CarNumber))
	}
	if !randTest.wantSeats.IsSame(resp.Seats) {
		return bencherror.PreTestErrs.AddError(bencherror.NewSimpleCriticalError("GET %s: レスポンスに含まれる座席が不正です", endpointPath))
	}
	if !randTest.wantCars.IsSame(resp.Cars) {
		return bencherror.PreTestErrs.AddError(bencherror.NewSimpleCriticalError("GET %s: レスポンスに含まれる車両が不正です", endpointPath))
	}

	return nil
}

// 予約

func pretestNormalReservation(ctx context.Context, client *isutrain.Client, paymentClient *payment.Client) error {
	// FIXME: 最初から登録されて入る、２つくらいのユーザで試す

	user := &isutrain.User{
		Email:    "hoge@example.com",
		Password: "hoge",
	}

	if err := client.Signup(ctx, user.Email, user.Password); err != nil {
		return bencherror.PreTestErrs.AddError(err)
	}

	if err := client.Login(ctx, user.Email, user.Password); err != nil {
		return bencherror.PreTestErrs.AddError(err)
	}

	_, err := client.ListStations(ctx)
	if err != nil {
		return bencherror.PreTestErrs.AddError(err)
	}

	var (
		useAt              = time.Date(2020, 1, 1, 10, 0, 0, 0, time.UTC)
		departure, arrival = "東京", "大阪"
		trainClass         = "最速"
		trainName          = "49"
		carNum             = 9
		seatClass          = "premium"
		adult, child       = 1, 1
	)
	_, err = client.SearchTrains(ctx, useAt, departure, arrival, "最速", adult, child)
	if err != nil {
		return bencherror.PreTestErrs.AddError(err)
	}

	// FIXME: 日付、列車クラス、名前、車両番号、乗車駅降車駅を指定
	searchTrainSeatsResp, err := client.SearchTrainSeats(ctx,
		useAt,
		trainClass, trainName, carNum, departure, arrival)
	if err != nil {
		return bencherror.PreTestErrs.AddError(err)
	}

	availSeats := filterTrainSeats(searchTrainSeatsResp, 2)

	reserveResp, err := client.Reserve(ctx, trainClass, trainName, seatClass, availSeats, departure, arrival,
		useAt,
		carNum, child, adult)
	if err != nil {
		return bencherror.PreTestErrs.AddError(err)
	}

	// 予約が、予約詳細に反映されているか
	showReservationResp, err := client.ShowReservation(ctx, reserveResp.ReservationID)
	if err != nil {
		return bencherror.PreTestErrs.AddError(err)
	}

	if reserveResp.ReservationID != showReservationResp.ReservationID {
		return bencherror.PreTestErrs.AddError(bencherror.NewSimpleCriticalError("GET %s: 予約レスポンスに含まれる予約IDと、予約詳細に含まれる予約IDが異なります: want=%d, got=%d", endpoint.GetPath(endpoint.ShowReservation), reserveResp.ReservationID, showReservationResp.ReservationID))
	}

	// 予約が、予約一覧に反映されているか
	listReservationResp, err := client.ListReservations(ctx)
	if err != nil {
		return bencherror.PreTestErrs.AddError(err)
	}

	var found bool
	for _, reservation := range listReservationResp {
		if reservation.ReservationID == reserveResp.ReservationID {
			found = true
			break
		}
	}

	if !found {
		return bencherror.PreTestErrs.AddError(bencherror.NewSimpleCriticalError("GET %s: 予約一覧に、予約したはずの予約IDが含まれていません: want=%d", endpoint.GetPath(endpoint.ListReservations), reserveResp.ReservationID))
	}

	cardToken, err := paymentClient.RegistCard(ctx, "11111111", "222", "10/50")
	if err != nil {
		return bencherror.PreTestErrs.AddError(err)
	}

	if err = client.CommitReservation(ctx, reserveResp.ReservationID, cardToken); err != nil {
		return bencherror.PreTestErrs.AddError(err)
	}

	if err := client.CancelReservation(ctx, reserveResp.ReservationID); err != nil {
		return bencherror.PreTestErrs.AddError(err)
	}

	return nil
}

// https://github.com/chibiegg/isucon9-final/blob/master/webapp/go/main.go#L1102
func pretestReserveVagueSeats(ctx context.Context, client *isutrain.Client) error {
	// 細かく見ずに、曖昧検索で結果が返ってきてそうなら通す

	err := registerUserAndLogin(ctx, client, &isutrain.User{
		Email:    "puser5@example.com",
		Password: "puser5",
	})
	if err != nil {
		return bencherror.PreTestErrs.AddError(err)
	}

	d := time.Date(2020, 1, 1, 6, 50, 0, 0, time.UTC)
	_, err = client.Reserve(ctx, "最速", "1", "premium", isutrain.TrainSeats{}, "東京", "大阪", d, 8, 1, 1)
	if err != nil {
		return bencherror.PreTestErrs.AddError(err)
	}

	err = client.Logout(ctx)
	if err != nil {
		return bencherror.PreTestErrs.AddError(err)
	}

	return nil
}

// ユーティリティ

func pretestAbnormalDate(ctx context.Context, client *isutrain.Client) error {
	var (
		availDate = config.ReservationStartDate.AddDate(0, 0, config.AvailableDays)
		d         = availDate.AddDate(0, 0, 1)
	)

	client, err := isutrain.NewClient()
	if err != nil {
		return bencherror.PreTestErrs.AddError(err)
	}

	err = registerUserAndLogin(ctx, client, &isutrain.User{
		Email:    "puser6@example.com",
		Password: "puser6",
	})
	if err != nil {
		return bencherror.PreTestErrs.AddError(err)
	}

	_, err = client.SearchTrains(ctx, d, "東京", "大阪", "最速", 1, 1, isutrain.StatusCodeOpt(http.StatusNotFound))
	if err != nil {
		fmt.Println(err.Error())
		return bencherror.PreTestErrs.AddError(err)
	}

	return nil
}

// 異常系

// PreTestAbnormalLogin は不正なパスワードでのログインを試みます
func pretestAbnormalLogin(ctx context.Context, client *isutrain.Client) error {
	if err := client.Login(ctx, "FikyavwocZear@example.com", "jieldirAwsabyonsInd", isutrain.StatusCodeOpt(http.StatusForbidden)); err != nil {
		return bencherror.PreTestErrs.AddError(err)
	}

	return nil
}

// https://github.com/chibiegg/isucon9-final/blob/master/webapp/go/main.go#L960
func pretestAbnormalReserveNonStoppableStation(ctx context.Context, client *isutrain.Client) error {
	endpointPath := endpoint.GetPath(endpoint.Reserve)
	d := time.Date(2020, 1, 1, 6, 0, 0, 0, time.UTC)
	// Express
	_, err := client.Reserve(ctx, "最速", "1", "premium", isutrain.TrainSeats{}, "古岡", "大阪", d, 8, 1, 1, isutrain.StatusCodeOpt(http.StatusBadRequest))
	if err != nil {
		return bencherror.PreTestErrs.AddError(bencherror.NewSimpleCriticalError("POST %s: 最速が止まらない駅で予約可能です", endpointPath))
	}

	// SemiExpress
	_, err = client.Reserve(ctx, "中間", "3", "reserved", isutrain.TrainSeats{}, "東京", "絵寒町", d, 8, 1, 1, isutrain.StatusCodeOpt(http.StatusBadRequest))
	if err != nil {
		return bencherror.PreTestErrs.AddError(bencherror.NewSimpleCriticalError("POST %s: 中間が止まらない駅で予約可能です", endpointPath))
	}

	return nil
}

// https://github.com/chibiegg/isucon9-final/blob/master/webapp/go/main.go#L1077
func pretestAbnormalReserveInvalidSection(ctx context.Context, client *isutrain.Client) error {
	endpointPath := endpoint.GetPath(endpoint.Reserve)
	d := time.Date(2020, 1, 1, 6, 50, 0, 0, time.UTC)
	// 適当に選んだ列車が走らない区間を選ぶ
	_, err := client.Reserve(ctx, "最速", "12", "premium", isutrain.TrainSeats{}, "大阪", "東京", d, 8, 1, 1, isutrain.StatusCodeOpt(http.StatusBadRequest))
	if err != nil {
		return bencherror.PreTestErrs.AddError(bencherror.NewSimpleCriticalError("POST %s: 列車が運行してない区間の予約が可能です", endpointPath))
	}

	return nil
}
