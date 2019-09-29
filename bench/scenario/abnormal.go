package scenario

import (
	"context"
	"math/rand"
	"net/http"

	"github.com/chibiegg/isucon9-final/bench/internal/bencherror"
	"github.com/chibiegg/isucon9-final/bench/internal/config"
	"github.com/chibiegg/isucon9-final/bench/internal/util"
	"github.com/chibiegg/isucon9-final/bench/internal/xrandom"
	"github.com/chibiegg/isucon9-final/bench/isutrain"
)

func AbnormalLoginScenario(ctx context.Context) error {
	var (
		email, err1    = util.SecureRandomStr(10)
		password, err2 = util.SecureRandomStr(10)
	)
	if err1 != nil {
		return err1
	}
	if err2 != nil {
		return err2
	}

	client, err := isutrain.NewClient()
	if err != nil {
		return err
	}

	err = client.Login(ctx, email, password, &isutrain.ClientOption{
		WantStatusCode: http.StatusUnauthorized,
	})
	if err != nil {
		return err
	}

	return nil
}

// 指定列車の運用区間外で予約を取ろうとして、きちんと弾かれるかチェック
func AbnormalReserveWrongSection(ctx context.Context) error {

	client, err := isutrain.NewClient()
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	if config.Debug {
		client.ReplaceMockTransport()
	}

	err = registerUserAndLogin(ctx, client)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	_, err = client.ListStations(ctx, nil)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	useAt := xrandom.GetRandomUseAt()
	trains, err := client.SearchTrains(ctx, useAt, "東京", "大阪", "最速", nil)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	trainIdx := rand.Intn(len(trains))
	train := trains[trainIdx]
	carNum := 5
	seatResp, err := client.ListTrainSeats(ctx,
		useAt,
		train.Class, train.Name, carNum, "東京", "大阪", nil)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	validSeats, err := assertListTrainSeats(seatResp, 2)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	_, err = client.Reserve(ctx,
		train.Class, train.Name,
		xrandom.GetSeatClass(train.Class, carNum),
		validSeats, "山田", "夷太寺", useAt,
		carNum, 1, 1, "", nil,
	)
	if err == nil {
		err = bencherror.NewSimpleCriticalError("予約できない区間が予約できました")
		return bencherror.BenchmarkErrs.AddError(err)
	}

	return nil
}

// 列車の指定号車に存在しない席を予約しようとし、エラーになるかチェック
func AbnormalReserveWrongSeat(ctx context.Context) error {

	client, err := isutrain.NewClient()
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	if config.Debug {
		client.ReplaceMockTransport()
	}

	err = registerUserAndLogin(ctx, client)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	_, err = client.ListStations(ctx, nil)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	useAt := xrandom.GetRandomUseAt()
	departure, arrival := xrandom.GetRandomSection()
	trains, err := client.SearchTrains(ctx, useAt, departure, arrival, "", nil)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	trainIdx := rand.Intn(len(trains))
	train := trains[trainIdx]
	carNum := xrandom.GetRandomCarNumber(train.Class, "reserved")
	seatResp, err := client.ListTrainSeats(ctx,
		useAt,
		train.Class, train.Name, carNum, departure, arrival, nil)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	validSeats, err := assertListTrainSeats(seatResp, 2)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	validSeats[0].Row = 30
	validSeats[1].Column = "G"

	_, err = client.Reserve(ctx,
		train.Class, train.Name,
		xrandom.GetSeatClass(train.Class, carNum),
		validSeats, departure, arrival, useAt,
		carNum, 1, 1, "", nil,
	)
	if err == nil {
		err = bencherror.NewSimpleCriticalError("予約できない座席が予約できました")
		return bencherror.BenchmarkErrs.AddError(err)
	}

	return nil
}

func AbnormalReserveWithCSRFTokenScenario(ctx context.Context) error {
	return nil
}
