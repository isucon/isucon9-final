package scenario

import (
	"context"
	"math/rand"
	"time"

	"github.com/chibiegg/isucon9-final/bench/internal/bencherror"
	"github.com/chibiegg/isucon9-final/bench/internal/isutraindb"
	"github.com/chibiegg/isucon9-final/bench/internal/xrandom"
	"github.com/chibiegg/isucon9-final/bench/isutrain"
	"github.com/chibiegg/isucon9-final/bench/payment"
)

func registerUserAndLogin(ctx context.Context, client *isutrain.Client, user *isutrain.User) error {
	/* ユーザー作成しログインする */

	err := client.Signup(ctx, user.Email, user.Password)
	if err != nil {
		return err
	}

	err = client.Login(ctx, user.Email, user.Password)
	if err != nil {
		return err
	}

	return nil
}

// 曖昧予約
func createSimpleReservation(ctx context.Context, client *isutrain.Client, user *isutrain.User, useAt time.Time, departure, arrival, trainClass string, adult, child int) (*isutrain.ReserveResponse, error) {
	/* 予約を作成する */

	// lgr := zap.S()

	// 決済サービスのクライアントを作成
	paymentClient, err := payment.NewClient()
	if err != nil {
		return nil, err
	}

	_, err = client.ListStations(ctx)
	if err != nil {
		return nil, err
	}

	trains, err := client.SearchTrains(ctx, useAt, departure, arrival, trainClass, adult, child)
	if err != nil {
		return nil, err
	}

	if len(trains) == 0 {
		err := bencherror.NewSimpleCriticalError("列車検索結果が空でした")
		return nil, err
	}

	train := trains[0]

	reserveResp, err := client.Reserve(ctx,
		train.Class, train.Name,
		"premium", isutrain.TrainSeats{},
		departure, arrival, useAt,
		0, adult, child)
	if err != nil {
		return nil, err
	}

	cardToken, err := paymentClient.RegistCard(ctx, "11111111", "222", "10/50")
	if err != nil {
		return nil, err
	}

	err = client.CommitReservation(ctx, reserveResp.ReservationID, cardToken)
	if err != nil {
		return nil, err
	}

	return reserveResp, nil
}

// 列車種別以外指定で予約
func createSpecifiedReservation(ctx context.Context, client *isutrain.Client, user *isutrain.User, useAt time.Time, departure, arrival string, adult, child int) (*isutrain.ReserveResponse, error) {

	paymentClient, err := payment.NewClient()
	if err != nil {
		return nil, err
	}

	_, err = client.ListStations(ctx)
	if err != nil {
		return nil, bencherror.BenchmarkErrs.AddError(err)
	}

	trains, err := client.SearchTrains(ctx, useAt, departure, arrival, "", adult, child)
	if err != nil {
		return nil, bencherror.BenchmarkErrs.AddError(err)
	}

	if len(trains) == 0 {
		err := bencherror.NewSimpleCriticalError("列車検索結果が空でした")
		return nil, bencherror.BenchmarkErrs.AddError(err)
	}

	trainIdx := rand.Intn(len(trains))
	train := trains[trainIdx]
	carNum := xrandom.GetRandomCarNumber(train.Class, "reserved")
	listTrainSeatsResp, err := client.SearchTrainSeats(ctx,
		useAt,
		train.Class, train.Name, carNum, departure, arrival)
	if err != nil {
		return nil, bencherror.BenchmarkErrs.AddError(err)
	}

	availSeats := FilterTrainSeats(listTrainSeatsResp, 2)

	reserveResp, err := client.Reserve(ctx,
		train.Class, train.Name,
		isutraindb.GetSeatClass(train.Class, carNum),
		availSeats, departure, arrival, useAt,
		carNum, 1, 1,
	)
	if err != nil {
		return nil, bencherror.BenchmarkErrs.AddError(err)
	}

	cardToken, err := paymentClient.RegistCard(ctx, "11111111", "222", "10/50")
	if err != nil {
		return nil, bencherror.BenchmarkErrs.AddError(err)
	}

	err = client.CommitReservation(ctx, reserveResp.ReservationID, cardToken)
	if err != nil {
		return nil, bencherror.BenchmarkErrs.AddError(err)
	}

	return reserveResp, nil
}

func FilterTrainSeats(resp *isutrain.SearchTrainSeatsResponse, count int) isutrain.TrainSeats {
	availSeats := isutrain.TrainSeats{}
	for _, seat := range resp.Seats {
		if len(availSeats) == count {
			break
		}
		if !seat.IsOccupied {
			availSeats = append(availSeats, seat)
		}
	}

	return availSeats
}
