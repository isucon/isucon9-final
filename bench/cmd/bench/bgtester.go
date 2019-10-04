package main

import (
	"context"
	"time"

	"github.com/chibiegg/isucon9-final/bench/internal/bencherror"
	"github.com/chibiegg/isucon9-final/bench/internal/endpoint"
	"github.com/chibiegg/isucon9-final/bench/isutrain"
	"github.com/chibiegg/isucon9-final/bench/scenario"
	"go.uber.org/zap"
)

/**
mysql> SELECT COUNT(*) FROM seat_master WHERE train_class="遅いやつ" AND seat_class="reserved" AND is_smoking_seat=0;
+----------+
| COUNT(*) |
+----------+
|       65 |
+----------+
1 row in set (0.01 sec)

mysql> SELECT COUNT(*) FROM seat_master WHERE train_class="遅いやつ" AND seat_class="reserved" AND is_smoking_seat=1;
+----------+
| COUNT(*) |
+----------+
|        0 |
+----------+
1 row in set (0.01 sec)

mysql>
*/

// BgChecker は、バックグラウンドでSeatAvailabilityの整合性チェックを行います
type BgTester struct {
	client *isutrain.Client

	remainSeats                      int
	reserveDate                      time.Time
	reserveTrainClass                string
	reserveTrainName                 string
	reserveSeatClass                 string
	reserveDeparture, reserveArrival string
	reserveCarNum                    int
}

func newBgTester() (*BgTester, error) {
	ctx := context.Background()

	client, err := isutrain.NewClient()
	if err != nil {
		return nil, err
	}

	user := &isutrain.User{
		Email:    "Clacvuwobfonsakchayraill@example.com",
		Password: "Clacvuwobfonsakchayraill",
	}

	if err := client.Signup(ctx, user.Email, user.Password); err != nil {
		return nil, err
	}

	if err := client.Login(ctx, user.Email, user.Password); err != nil {
		return nil, err
	}

	return &BgTester{
		client:            client,
		remainSeats:       65,
		reserveDate:       time.Date(2020, 1, 1, 20, 0, 0, 0, time.UTC),
		reserveTrainClass: "遅いやつ",
		reserveTrainName:  "163",
		reserveSeatClass:  "reserved",
		reserveDeparture:  "古岡",
		reserveArrival:    "大阪",
		reserveCarNum:     16,
	}, nil
}

func (t *BgTester) reserve(ctx context.Context) error {
	lgr := zap.S()

	searchTrainSeatsResp, err := t.client.SearchTrainSeats(ctx, t.reserveDate, t.reserveTrainClass, t.reserveTrainName,
		t.reserveCarNum, t.reserveDeparture, t.reserveArrival)
	if err != nil {
		return err
	}

	availSeats := scenario.FilterTrainSeats(searchTrainSeatsResp, 2)

	_, err = t.client.Reserve(ctx, t.reserveTrainClass, t.reserveTrainName,
		t.reserveSeatClass, availSeats, t.reserveDeparture, t.reserveArrival, t.reserveDate,
		t.reserveCarNum, 1, 1, isutrain.DisableAssertOpt())
	if err != nil {
		return err
	}

	// showReservationResp, err := t.client.ShowReservation(ctx, reserveResp.ReservationID)
	// if err != nil {
	// 	return err
	// }

	lgr.Infof("%d個の座席を予約しました", len(availSeats))
	t.remainSeats -= len(availSeats)

	return nil
}

func (t *BgTester) bgtestSeatAvailability(ctx context.Context, wantReservedSeatAvailability string) error {
	endpointPath := endpoint.GetPath(endpoint.SearchTrains)
	searchTrainsResp, err := t.client.SearchTrains(ctx, t.reserveDate, t.reserveDeparture, t.reserveArrival, t.reserveTrainClass, 1, 1)
	if err != nil {
		return err
	}

	var targetTrain *isutrain.Train
	for _, train := range searchTrainsResp {
		if train.Name == t.reserveTrainName {
			targetTrain = train
		}
	}
	if targetTrain == nil {
		return nil
	}

	sa := targetTrain.SeatAvailability
	if sa["reserved"] != wantReservedSeatAvailability {
		return bencherror.BenchmarkErrs.AddError(bencherror.NewSimpleCriticalError("GET %s: seatAvailabilityが不正です: want=%s, got=%s", endpointPath, wantReservedSeatAvailability, sa["reserved"]))
	}

	return nil
}

func (t *BgTester) run(ctx context.Context) error {
	if err := t.bgtestSeatAvailability(ctx, "○"); err != nil {
		return err
	}

	for t.remainSeats > 10 {
		t.reserve(ctx)
	}
	if err := t.bgtestSeatAvailability(ctx, "△"); err != nil {
		return err
	}

	for t.remainSeats > 0 {
		t.reserve(ctx)
	}
	if err := t.bgtestSeatAvailability(ctx, "×"); err != nil {
		return err
	}

	return nil
}
