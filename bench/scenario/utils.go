package scenario

import (
	"context"
	"time"

	"github.com/chibiegg/isucon9-final/bench/internal/bencherror"
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

func createSimpleReservation(ctx context.Context, client *isutrain.Client, user *isutrain.User, useAt time.Time, departure, arrival, train_class string, adult, child int) (*isutrain.ReserveResponse, error) {
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

	trains, err := client.SearchTrains(ctx, useAt, departure, arrival, train_class)
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

func payForReservation(ctx context.Context, client *isutrain.Client, paymentClient *payment.Client, reservationId int) error {
	cardToken, err := paymentClient.RegistCard(ctx, "11111111", "222", "10/50")
	if err != nil {
		return err
	}

	err = client.CommitReservation(ctx, reservationId, cardToken)
	if err != nil {
		return err
	}

	return nil
}

func filterTrainSeats(resp *isutrain.SearchTrainSeatsResponse, count int) isutrain.TrainSeats {
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
