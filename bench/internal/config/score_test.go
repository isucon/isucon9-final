package config

import (
	"math"
	"testing"
	"time"

	"gopkg.in/go-playground/assert.v1"
)

func TestFareMultiplier(t *testing.T) {
	tests := []struct {
		trainClass         string
		seatClass          string
		startDate          time.Time
		wantFareMultiplier float64
	}{
		{
			trainClass:         "最速",
			seatClass:          "premium",
			startDate:          time.Date(2020, 01, 01, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 15.000,
		},
		{
			trainClass:         "最速",
			seatClass:          "reserved",
			startDate:          time.Date(2020, 01, 01, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 9.375,
		},
		{
			trainClass:         "最速",
			seatClass:          "non-reserved",
			startDate:          time.Date(2020, 01, 01, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 7.500,
		},
		{
			trainClass:         "中間",
			seatClass:          "premium",
			startDate:          time.Date(2020, 01, 01, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 10.000,
		},
		{
			trainClass:         "中間",
			seatClass:          "reserved",
			startDate:          time.Date(2020, 01, 01, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 6.250,
		},
		{
			trainClass:         "中間",
			seatClass:          "non-reserved",
			startDate:          time.Date(2020, 01, 01, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 5.000,
		},
		{
			trainClass:         "遅いやつ",
			seatClass:          "premium",
			startDate:          time.Date(2020, 01, 01, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 8.000,
		},
		{
			trainClass:         "遅いやつ",
			seatClass:          "reserved",
			startDate:          time.Date(2020, 01, 01, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 5.000,
		},
		{
			trainClass:         "遅いやつ",
			seatClass:          "non-reserved",
			startDate:          time.Date(2020, 01, 01, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 4.000,
		},
		{
			trainClass:         "最速",
			seatClass:          "premium",
			startDate:          time.Date(2020, 01, 06, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 3.000,
		},
		{
			trainClass:         "最速",
			seatClass:          "reserved",
			startDate:          time.Date(2020, 01, 06, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 1.875,
		},
		{
			trainClass:         "最速",
			seatClass:          "non-reserved",
			startDate:          time.Date(2020, 01, 06, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 1.500,
		},
		{
			trainClass:         "中間",
			seatClass:          "premium",
			startDate:          time.Date(2020, 01, 06, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 2.000,
		},
		{
			trainClass:         "中間",
			seatClass:          "reserved",
			startDate:          time.Date(2020, 01, 06, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 1.250,
		},
		{
			trainClass:         "中間",
			seatClass:          "non-reserved",
			startDate:          time.Date(2020, 01, 06, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 1.000,
		},
		{
			trainClass:         "遅いやつ",
			seatClass:          "premium",
			startDate:          time.Date(2020, 01, 06, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 1.600,
		},
		{
			trainClass:         "遅いやつ",
			seatClass:          "reserved",
			startDate:          time.Date(2020, 01, 06, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 1.000,
		},
		{
			trainClass:         "遅いやつ",
			seatClass:          "non-reserved",
			startDate:          time.Date(2020, 01, 06, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 0.800,
		},
		{
			trainClass:         "最速",
			seatClass:          "premium",
			startDate:          time.Date(2020, 03, 13, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 9.000,
		},
		{
			trainClass:         "最速",
			seatClass:          "reserved",
			startDate:          time.Date(2020, 03, 13, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 5.625,
		},
		{
			trainClass:         "最速",
			seatClass:          "non-reserved",
			startDate:          time.Date(2020, 03, 13, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 4.500,
		},
		{
			trainClass:         "中間",
			seatClass:          "premium",
			startDate:          time.Date(2020, 03, 13, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 6.000,
		},
		{
			trainClass:         "中間",
			seatClass:          "reserved",
			startDate:          time.Date(2020, 03, 13, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 3.750,
		},
		{
			trainClass:         "中間",
			seatClass:          "non-reserved",
			startDate:          time.Date(2020, 03, 13, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 3.000,
		},
		{
			trainClass:         "遅いやつ",
			seatClass:          "premium",
			startDate:          time.Date(2020, 03, 13, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 4.800,
		},
		{
			trainClass:         "遅いやつ",
			seatClass:          "reserved",
			startDate:          time.Date(2020, 03, 13, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 3.000,
		},
		{
			trainClass:         "遅いやつ",
			seatClass:          "non-reserved",
			startDate:          time.Date(2020, 03, 13, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 2.400,
		},
		{
			trainClass:         "最速",
			seatClass:          "premium",
			startDate:          time.Date(2020, 04, 01, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 3.000,
		},
		{
			trainClass:         "最速",
			seatClass:          "reserved",
			startDate:          time.Date(2020, 04, 01, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 1.875,
		},
		{
			trainClass:         "最速",
			seatClass:          "non-reserved",
			startDate:          time.Date(2020, 04, 01, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 1.500,
		},
		{
			trainClass:         "中間",
			seatClass:          "premium",
			startDate:          time.Date(2020, 04, 01, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 2.000,
		},
		{
			trainClass:         "中間",
			seatClass:          "reserved",
			startDate:          time.Date(2020, 04, 01, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 1.250,
		},
		{
			trainClass:         "中間",
			seatClass:          "non-reserved",
			startDate:          time.Date(2020, 04, 01, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 1.000,
		},
		{
			trainClass:         "遅いやつ",
			seatClass:          "premium",
			startDate:          time.Date(2020, 04, 01, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 1.600,
		},
		{
			trainClass:         "遅いやつ",
			seatClass:          "reserved",
			startDate:          time.Date(2020, 04, 01, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 1.000,
		},
		{
			trainClass:         "遅いやつ",
			seatClass:          "non-reserved",
			startDate:          time.Date(2020, 04, 01, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 0.800,
		},
		{
			trainClass:         "最速",
			seatClass:          "premium",
			startDate:          time.Date(2020, 04, 24, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 15.000,
		},
		{
			trainClass:         "最速",
			seatClass:          "reserved",
			startDate:          time.Date(2020, 04, 24, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 9.375,
		},
		{
			trainClass:         "最速",
			seatClass:          "non-reserved",
			startDate:          time.Date(2020, 04, 24, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 7.500,
		},
		{
			trainClass:         "中間",
			seatClass:          "premium",
			startDate:          time.Date(2020, 04, 24, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 10.000,
		},
		{
			trainClass:         "中間",
			seatClass:          "reserved",
			startDate:          time.Date(2020, 04, 24, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 6.250,
		},
		{
			trainClass:         "中間",
			seatClass:          "non-reserved",
			startDate:          time.Date(2020, 04, 24, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 5.000,
		},
		{
			trainClass:         "遅いやつ",
			seatClass:          "premium",
			startDate:          time.Date(2020, 04, 24, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 8.000,
		},
		{
			trainClass:         "遅いやつ",
			seatClass:          "reserved",
			startDate:          time.Date(2020, 04, 24, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 5.000,
		},
		{
			trainClass:         "遅いやつ",
			seatClass:          "non-reserved",
			startDate:          time.Date(2020, 04, 24, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 4.000,
		},
		{
			trainClass:         "最速",
			seatClass:          "premium",
			startDate:          time.Date(2020, 05, 11, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 3.000,
		},
		{
			trainClass:         "最速",
			seatClass:          "reserved",
			startDate:          time.Date(2020, 05, 11, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 1.875,
		},
		{
			trainClass:         "最速",
			seatClass:          "non-reserved",
			startDate:          time.Date(2020, 05, 11, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 1.500,
		},
		{
			trainClass:         "中間",
			seatClass:          "premium",
			startDate:          time.Date(2020, 05, 11, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 2.000,
		},
		{
			trainClass:         "中間",
			seatClass:          "reserved",
			startDate:          time.Date(2020, 05, 11, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 1.250,
		},
		{
			trainClass:         "中間",
			seatClass:          "non-reserved",
			startDate:          time.Date(2020, 05, 11, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 1.000,
		},
		{
			trainClass:         "遅いやつ",
			seatClass:          "premium",
			startDate:          time.Date(2020, 05, 11, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 1.600,
		},
		{
			trainClass:         "遅いやつ",
			seatClass:          "reserved",
			startDate:          time.Date(2020, 05, 11, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 1.000,
		},
		{
			trainClass:         "遅いやつ",
			seatClass:          "non-reserved",
			startDate:          time.Date(2020, 05, 11, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 0.800,
		},
		{
			trainClass:         "最速",
			seatClass:          "premium",
			startDate:          time.Date(2020, 8, 07, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 9.000,
		},
		{
			trainClass:         "最速",
			seatClass:          "reserved",
			startDate:          time.Date(2020, 8, 07, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 5.625,
		},
		{
			trainClass:         "最速",
			seatClass:          "non-reserved",
			startDate:          time.Date(2020, 8, 07, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 4.500,
		},
		{
			trainClass:         "中間",
			seatClass:          "premium",
			startDate:          time.Date(2020, 8, 07, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 6.000,
		},
		{
			trainClass:         "中間",
			seatClass:          "reserved",
			startDate:          time.Date(2020, 8, 07, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 3.750,
		},
		{
			trainClass:         "中間",
			seatClass:          "non-reserved",
			startDate:          time.Date(2020, 8, 07, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 3.000,
		},
		{
			trainClass:         "遅いやつ",
			seatClass:          "premium",
			startDate:          time.Date(2020, 8, 07, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 4.800,
		},
		{
			trainClass:         "遅いやつ",
			seatClass:          "reserved",
			startDate:          time.Date(2020, 8, 07, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 3.000,
		},
		{
			trainClass:         "遅いやつ",
			seatClass:          "non-reserved",
			startDate:          time.Date(2020, 8, 07, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 2.400,
		},
		{
			trainClass:         "最速",
			seatClass:          "premium",
			startDate:          time.Date(2020, 8, 24, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 3.000,
		},
		{
			trainClass:         "最速",
			seatClass:          "reserved",
			startDate:          time.Date(2020, 8, 24, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 1.875,
		},
		{
			trainClass:         "最速",
			seatClass:          "non-reserved",
			startDate:          time.Date(2020, 8, 24, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 1.500,
		},
		{
			trainClass:         "中間",
			seatClass:          "premium",
			startDate:          time.Date(2020, 8, 24, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 2.000,
		},
		{
			trainClass:         "中間",
			seatClass:          "reserved",
			startDate:          time.Date(2020, 8, 24, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 1.250,
		},
		{
			trainClass:         "中間",
			seatClass:          "non-reserved",
			startDate:          time.Date(2020, 8, 24, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 1.000,
		},
		{
			trainClass:         "遅いやつ",
			seatClass:          "premium",
			startDate:          time.Date(2020, 8, 24, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 1.600,
		},
		{
			trainClass:         "遅いやつ",
			seatClass:          "reserved",
			startDate:          time.Date(2020, 8, 24, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 1.000,
		},
		{
			trainClass:         "遅いやつ",
			seatClass:          "non-reserved",
			startDate:          time.Date(2020, 8, 24, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 0.800,
		},
		{
			trainClass:         "最速",
			seatClass:          "premium",
			startDate:          time.Date(2020, 12, 25, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 15.000,
		},
		{
			trainClass:         "最速",
			seatClass:          "reserved",
			startDate:          time.Date(2020, 12, 25, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 9.375,
		},
		{
			trainClass:         "最速",
			seatClass:          "non-reserved",
			startDate:          time.Date(2020, 12, 25, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 7.500,
		},
		{
			trainClass:         "中間",
			seatClass:          "premium",
			startDate:          time.Date(2020, 12, 25, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 10.000,
		},
		{
			trainClass:         "中間",
			seatClass:          "reserved",
			startDate:          time.Date(2020, 12, 25, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 6.250,
		},
		{
			trainClass:         "中間",
			seatClass:          "non-reserved",
			startDate:          time.Date(2020, 12, 25, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 5.000,
		},
		{
			trainClass:         "遅いやつ",
			seatClass:          "premium",
			startDate:          time.Date(2020, 12, 25, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 8.000,
		},
		{
			trainClass:         "遅いやつ",
			seatClass:          "reserved",
			startDate:          time.Date(2020, 12, 25, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 5.000,
		},
		{
			trainClass:         "遅いやつ",
			seatClass:          "non-reserved",
			startDate:          time.Date(2020, 12, 25, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 4.000,
		},
		// 間の値
		{
			trainClass:         "遅いやつ",
			seatClass:          "non-reserved",
			startDate:          time.Date(2020, 1, 05, 0, 0, 0, 0, time.Local),
			wantFareMultiplier: 4.000,
		},
	}

	round := func(f float64, places int) float64 {
		shift := math.Pow(10, float64(places))
		return math.Floor(f*shift+.5) / shift
	}

	for _, tt := range tests {
		m := GetFareMultiplier(tt.trainClass, tt.seatClass, tt.startDate)
		assert.Equal(t, tt.wantFareMultiplier, round(m, 3))
	}
}
