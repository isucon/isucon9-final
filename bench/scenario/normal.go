package scenario

import (
	"context"
	"log"
	"math/rand"
	"time"

	"github.com/chibiegg/isucon9-final/bench/internal/bencherror"
	"github.com/chibiegg/isucon9-final/bench/internal/config"
	"github.com/chibiegg/isucon9-final/bench/internal/xrandom"
	"github.com/chibiegg/isucon9-final/bench/isutrain"
	"github.com/chibiegg/isucon9-final/bench/payment"
	"go.uber.org/zap"
)

// NormalScenario は基本的な予約フローのシナリオです
func NormalScenario(ctx context.Context) error {
	lgr := zap.S()

	client, err := isutrain.NewClient()
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	paymentClient, err := payment.NewClient()
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	if config.Debug {
		client.ReplaceMockTransport()
	}

	user, err := xrandom.GetRandomUser()
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	err = client.Signup(ctx, user.Email, user.Password, nil)
	if err != nil {
		lgr.Infow("ユーザ登録失敗",
			"error", err,
			"email", user.Email,
			"password", user.Password,
		)
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

	useAt := xrandom.GetRandomUseAt()
	departure, arrival := xrandom.GetRandomSection()
	trains, err := client.SearchTrains(ctx, useAt, departure, arrival, nil)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	if len(trains) == 0 {
		return nil
	}

	trainIdx := rand.Intn(len(trains))
	train := trains[trainIdx]
	log.Printf("[normal:NormalScenario] train=%+v\n", train)
	log.Printf("[normal:NormalScenario] trainClass=%s trainName=%s\n", train.Class, train.Name)
	carNum := 8
	seatResp, err := client.ListTrainSeats(ctx,
		useAt,
		train.Class, train.Name, carNum, departure, arrival, nil)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	reservation, err := client.Reserve(ctx, train.Class, train.Name, xrandom.GetSeatClass(train.Class, carNum), seatResp.Seats[:2], "東京", "大阪",
		useAt,
		carNum, 1, 1, "isle", nil)
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

// 予約キャンセル含む(Commit後にキャンセル)
func NormalCancelScenario(ctx context.Context) error {
	client, err := isutrain.NewClient()
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	paymentClient, err := payment.NewClient()
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	if config.Debug {
		client.ReplaceMockTransport()
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

	var (
		departure = xrandom.GetRandomStations()
		arrival   = xrandom.GetRandomStations()
	)
	_, err = client.SearchTrains(ctx, time.Now().AddDate(1, 0, 0), departure, arrival, nil)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	_, err = client.ListTrainSeats(ctx, time.Now().AddDate(1, 0, 0), "こだま", "96号", 1, departure, arrival, nil)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	reservation, err := client.Reserve(ctx, "こだま", "69号", "premium", isutrain.TrainSeats{}, "東京", "名古屋", time.Now().AddDate(1, 0, 0), 1, 1, 1, "isle", nil)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	cardToken, err := paymentClient.RegistCard(ctx, "11111111", "222", "20/20")
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

	err = client.CancelReservation(ctx, reservation.ReservationID, nil)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	return nil
}

// 曖昧検索シナリオ
func NormalAmbigiousSearchScenario(ctx context.Context) error {
	return nil
}

// オリンピック開催期間に負荷をあげるシナリオ
// FIXME: initializeで指定された日数に応じ、負荷レベルを変化させる
func NormalOlympicParticipantsScenario(ctx context.Context) error {

	// NOTE: webappから見て、明らかに負荷が上がったと感じるレベルに持ってく必要がある
	// NOTE: 指定できる最大の日数で負荷をかける際、飽和しないようにする
	// NOTE:

	return nil
}
