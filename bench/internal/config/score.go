package config

import (
	"time"

	"go.uber.org/zap"
)

const (
	ApplicationPenaltyWeight = 5

	// TrivialPenaltyThreshold = 200
	TrivialPenaltyThreshold = 1
	TrivialPenaltyWeight    = 5
	// TrivialPenaltyWeight    = 5000
	// TrivialPenaltyPerCount  = 100
	TrivialPenaltyPerCount = 10
)

const (
	ReservedSeatExtraScore = 100
)

type FareMultiplierQuery struct {
	TrainClass string
	SeatClass  string
}

var (
	fareMultiplierMap = map[FareMultiplierQuery]float64{
		FareMultiplierQuery{TrainClass: "最速", SeatClass: "premium"}:        3.0,
		FareMultiplierQuery{TrainClass: "最速", SeatClass: "reserved"}:       1.875,
		FareMultiplierQuery{TrainClass: "最速", SeatClass: "non-reserved"}:   1.5,
		FareMultiplierQuery{TrainClass: "中間", SeatClass: "premium"}:        2.0,
		FareMultiplierQuery{TrainClass: "中間", SeatClass: "reserved"}:       1.25,
		FareMultiplierQuery{TrainClass: "中間", SeatClass: "non-reserved"}:   1.0,
		FareMultiplierQuery{TrainClass: "遅いやつ", SeatClass: "premium"}:      1.6,
		FareMultiplierQuery{TrainClass: "遅いやつ", SeatClass: "reserved"}:     1.0,
		FareMultiplierQuery{TrainClass: "遅いやつ", SeatClass: "non-reserved"}: 0.8,
	}

	seasons = []time.Time{
		// 正月
		time.Date(2020, 1, 01, 0, 0, 0, 0, time.Local),
		time.Date(2020, 1, 06, 0, 0, 0, 0, time.Local),
		// 春休み
		time.Date(2020, 3, 13, 0, 0, 0, 0, time.Local),
		time.Date(2020, 4, 01, 0, 0, 0, 0, time.Local),
		// GW
		time.Date(2020, 4, 24, 0, 0, 0, 0, time.Local),
		time.Date(2020, 5, 11, 0, 0, 0, 0, time.Local),
		// 夏休み
		time.Date(2020, 8, 07, 0, 0, 0, 0, time.Local),
		time.Date(2020, 8, 24, 0, 0, 0, 0, time.Local),
		// 年越し
		time.Date(2020, 12, 25, 0, 0, 0, 0, time.Local),
	}
)

func GetFareMultiplier(trainClass, seatClass string, useAt time.Time) float64 {
	lgr := zap.S()

	fareMultiplier := fareMultiplierMap[FareMultiplierQuery{TrainClass: trainClass, SeatClass: seatClass}]

	var seasonMultiplier float64
	switch {
	case useAt.Equal(seasons[0]) || (useAt.After(seasons[0]) && useAt.Before(seasons[1])):
		seasonMultiplier = 5.0
	case useAt.Equal(seasons[1]) || (useAt.After(seasons[1]) && useAt.Before(seasons[2])):
		seasonMultiplier = 1.0
	case useAt.Equal(seasons[2]) || (useAt.After(seasons[2]) && useAt.Before(seasons[3])):
		seasonMultiplier = 3.0
	case useAt.Equal(seasons[3]) || (useAt.After(seasons[3]) && useAt.Before(seasons[4])):
		seasonMultiplier = 1.0
	case useAt.Equal(seasons[4]) || (useAt.After(seasons[4]) && useAt.Before(seasons[5])):
		seasonMultiplier = 5.0
	case useAt.Equal(seasons[5]) || (useAt.After(seasons[5]) && useAt.Before(seasons[6])):
		seasonMultiplier = 1.0
	case useAt.Equal(seasons[6]) || (useAt.After(seasons[6]) && useAt.Before(seasons[7])):
		seasonMultiplier = 3.0
	case useAt.Equal(seasons[7]) || (useAt.After(seasons[7]) && useAt.Before(seasons[8])):
		seasonMultiplier = 1.0
	case useAt.Equal(seasons[8]):
		seasonMultiplier = 5.0
	}

	if fareMultiplier == 0 || seasonMultiplier == 0 {
		lgr.Warnw("運賃倍率もしくは期間倍率が不正です",
			"fare_multiplier", fareMultiplier,
			"season_multiplier", seasonMultiplier,
		)
	}

	return fareMultiplier * seasonMultiplier
}
