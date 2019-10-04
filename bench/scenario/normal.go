package scenario

import (
	"context"
	"math/rand"
	"time"

	"github.com/chibiegg/isucon9-final/bench/internal/bencherror"
	"github.com/chibiegg/isucon9-final/bench/internal/config"
	"github.com/chibiegg/isucon9-final/bench/internal/isutraindb"
	"github.com/chibiegg/isucon9-final/bench/internal/xrandom"
	"github.com/chibiegg/isucon9-final/bench/isutrain"
	"github.com/chibiegg/isucon9-final/bench/payment"
)

// NormalScenario は基本的な予約フローのシナリオです
func NormalScenario(ctx context.Context) error {
	client, err := isutrain.NewClient()
	if err != nil {
		return err
	}

	paymentClient, err := payment.NewClient()
	if err != nil {
		return err
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
		return bencherror.BenchmarkErrs.AddError(err)
	}

	_, err = client.ListStations(ctx)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	useAt := xrandom.GetRandomUseAt()
	departure, arrival := xrandom.GetRandomSection()
	adult, child := xrandom.GetRandomNumberOfPeople()
	trains, err := client.SearchTrains(ctx, useAt, departure, arrival, "", adult, child)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	if len(trains) == 0 {
		return bencherror.BenchmarkErrs.AddError(bencherror.NewSimpleApplicationError("列車検索の結果が空です"))
	}

	trainIdx := rand.Intn(len(trains))
	train := trains[trainIdx]
	carNum := xrandom.GetRandomCarNumber(train.Class, "premium")
	listTrainSeatsResp, err := client.SearchTrainSeats(ctx,
		useAt,
		train.Class, train.Name, carNum, departure, arrival)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	availSeats := filterTrainSeats(listTrainSeatsResp, 2)

	reserveResp, err := client.Reserve(ctx,
		train.Class, train.Name,
		isutraindb.GetSeatClass(train.Class, carNum), availSeats,
		departure, arrival, useAt,
		carNum, 1, 1)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	cardToken, err := paymentClient.RegistCard(ctx, "11111111", "222", "10/50")
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	err = client.CommitReservation(ctx, reserveResp.ReservationID, cardToken)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	_, err = client.ListReservations(ctx)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	reservation2, err := client.ShowReservation(ctx, reserveResp.ReservationID)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	if reserveResp.ReservationID != reservation2.ReservationID {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	if err := client.Logout(ctx); err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	return nil
}

// 予約キャンセル含む(Commit後にキャンセル)
func NormalCancelScenario(ctx context.Context) error {
	client, err := isutrain.NewClient()
	if err != nil {
		return err
	}

	paymentClient, err := payment.NewClient()
	if err != nil {
		return err
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
		return bencherror.BenchmarkErrs.AddError(err)
	}

	_, err = client.ListStations(ctx)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	var (
		useAt              = xrandom.GetRandomUseAt()
		departure, arrival = xrandom.GetRandomSection()
		adult, child       = xrandom.GetRandomNumberOfPeople()
	)
	trains, err := client.SearchTrains(ctx, useAt, departure, arrival, "", adult, child)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	if len(trains) == 0 {
		err := bencherror.NewSimpleCriticalError("列車検索結果が空でした")
		return bencherror.BenchmarkErrs.AddError(err)
	}

	trainIdx := rand.Intn(len(trains))
	train := trains[trainIdx]
	carNum := xrandom.GetRandomCarNumber(train.Class, "reserved")
	listTrainSeatsResp, err := client.SearchTrainSeats(ctx,
		useAt,
		train.Class, train.Name, carNum, departure, arrival)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	availSeats := filterTrainSeats(listTrainSeatsResp, 2)

	reserveResp, err := client.Reserve(ctx,
		train.Class, train.Name,
		isutraindb.GetSeatClass(train.Class, carNum),
		availSeats, departure, arrival, useAt,
		carNum, 1, 1,
	)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	cardToken, err := paymentClient.RegistCard(ctx, "11111111", "222", "10/50")
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	err = client.CommitReservation(ctx, reserveResp.ReservationID, cardToken)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	_, err = client.ListReservations(ctx)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	reservation2, err := client.ShowReservation(ctx, reserveResp.ReservationID)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	if reserveResp.ReservationID != reservation2.ReservationID {
		return bencherror.BenchmarkErrs.AddError(bencherror.NewSimpleCriticalError("予約確認で得られる予約IDが一致していません: got=%d, want=%d", reservation2.ReservationID, reserveResp.ReservationID))
	}

	err = client.CancelReservation(ctx, reserveResp.ReservationID)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	if err := client.Logout(ctx); err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	return nil
}

// 曖昧検索シナリオ
func NormalVagueSearchScenario(ctx context.Context) error {
	client, err := isutrain.NewClient()
	if err != nil {
		return err
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
		return bencherror.BenchmarkErrs.AddError(err)
	}

	user, err = xrandom.GetRandomUser()
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}
	if err = client.Signup(ctx, user.Email, user.Password); err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	if err := client.Login(ctx, user.Email, user.Password); err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	_, err = client.Reserve(ctx,
		"最速", "1", "premium", isutrain.TrainSeats{},
		"東京", "大阪", time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		1, 1, 1)
	if err != nil {
		bencherror.BenchmarkErrs.AddError(err)
	}

	return nil
}

func NormalManyCancelScenario(ctx context.Context, counter int) error {

	client, err := isutrain.NewClient()
	if err != nil {
		return err
	}

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

	var retErr error

	cancelIds := []int{}

	// たくさん予約を作る
	for i := 0; i < counter; i++ {
		useAt := xrandom.GetRandomUseAt()
		departure, arrival := xrandom.GetRandomSection()
		reservation, err := createSimpleReservation(ctx, client, user, useAt, departure, arrival, "遅いやつ", 3, 3)
		if err != nil {
			bencherror.BenchmarkErrs.AddError(err)
			retErr = err
			continue
		}
		cancelIds = append(cancelIds, reservation.ReservationID)
	}

	// 全部キャンセルする
	for _, reservationId := range cancelIds {
		err = client.CancelReservation(ctx, reservationId)
		if err != nil {
			bencherror.BenchmarkErrs.AddError(err)
			retErr = err
		}
	}

	if err := client.Logout(ctx); err != nil {
		bencherror.BenchmarkErrs.AddError(err)
		retErr = err
	}

	return retErr
}

func NormalManyAmbigiousSearchScenario(ctx context.Context, counter int) error {
	client, err := isutrain.NewClient()
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	if config.Debug {
		client.ReplaceMockTransport()
	}

	useAt := xrandom.GetRandomUseAt()
	departure, arrival := xrandom.GetRandomSection()

	var retErr error

	for i := 0; i < counter; i++ {
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

		_, err = createSimpleReservation(ctx, client, user, useAt, departure, arrival, "遅いやつ", 3, 3)
		if err != nil {
			bencherror.BenchmarkErrs.AddError(err)
			retErr = err
		}
	}

	return retErr
}
