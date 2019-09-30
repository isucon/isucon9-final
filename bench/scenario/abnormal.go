package scenario

import (
	"context"
	"math/rand"
	"net/http"
	"time"

	"github.com/chibiegg/isucon9-final/bench/internal/bencherror"
	"github.com/chibiegg/isucon9-final/bench/internal/config"
	"github.com/chibiegg/isucon9-final/bench/internal/isutraindb"
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

	err = client.Login(ctx, email, password, isutrain.StatusCodeOpt(http.StatusUnauthorized))
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
	trains, err := client.SearchTrains(ctx, useAt, "東京", "大阪", "最速")
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	trainIdx := rand.Intn(len(trains))
	train := trains[trainIdx]
	carNum := 5
	listTrainSeatsResp, err := client.ListTrainSeats(ctx,
		useAt,
		train.Class, train.Name, carNum, "東京", "大阪")
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	availSeats := filterTrainSeats(listTrainSeatsResp, 2)

	_, err = client.Reserve(ctx,
		train.Class, train.Name,
		isutraindb.GetSeatClass(train.Class, carNum),
		availSeats, "山田", "夷太寺", useAt,
		carNum, 1, 1, "",
		isutrain.StatusCodeOpt(http.StatusBadRequest))

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

	user, err := xrandom.GetRandomUser()
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	err = registerUserAndLogin(ctx, client, user)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	useAt := time.Date(2020, 1, 1, 10, 0, 0, 0, time.UTC)
	departure, arrival := "東京", "大阪"
	trains, err := client.SearchTrains(ctx, useAt, departure, arrival, "最速")
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	trainIdx := rand.Intn(len(trains))
	train := trains[trainIdx]
	carNum := xrandom.GetRandomCarNumber(train.Class, "reserved")
	listTrainSeatsResp, err := client.ListTrainSeats(ctx,
		useAt,
		train.Class, train.Name, carNum, departure, arrival)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	availSeats := filterTrainSeats(listTrainSeatsResp, 2)

	// FIXME: seatが空になるケースがあるので、上の座席列挙で固定の条件で検索をかける必要がある
	availSeats[0].Row = 30
	availSeats[1].Column = "G"

	_, err = client.Reserve(ctx,
		train.Class, train.Name,
		isutraindb.GetSeatClass(train.Class, carNum),
		availSeats, departure, arrival, useAt,
		carNum, 1, 1, "",
		isutrain.StatusCodeOpt(http.StatusNotFound))
	if err == nil {
		return bencherror.BenchmarkErrs.AddError(bencherror.NewCriticalError(err, "予約できない座席が予約できました"))
	}

	return nil
}

func AbnormalReserveWithCSRFTokenScenario(ctx context.Context) error {
	return nil
}
