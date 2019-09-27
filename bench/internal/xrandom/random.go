package xrandom

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/chibiegg/isucon9-final/bench/internal/bencherror"
	"github.com/chibiegg/isucon9-final/bench/internal/config"
	"github.com/chibiegg/isucon9-final/bench/internal/util"
	"github.com/chibiegg/isucon9-final/bench/isutrain"
)

func GetRandomStations() string {
	idx := rand.Intn(len(stations))
	return stations[idx]
}

func GetRandomTrainClass() string {
	idx := rand.Intn(len(trainClasses))
	return trainClasses[idx]
}

func GetRandomUseAt() time.Time {
	var (
		hour   = util.RandRangeIntn(6, 15)
		minute = util.RandRangeIntn(0, 59)
		sec    = util.RandRangeIntn(0, 59)
	)
	startTime := config.ReservationStartDate.Add(time.Duration(hour*60*60+minute*60+sec) * time.Second)
	// startTime := time.Date(2020, 1, 1, hour, minute, sec, 0, time.Local)
	days := rand.Intn(config.AvailableDays - 1)

	useAt := startTime.AddDate(0, 0, days)
	return useAt
}

func GetRandomSection() (station1 string, station2 string) {
	localStations := stations
	randIndexes := rand.Perm(len(localStations))

	return localStations[randIndexes[0]], localStations[randIndexes[1]]
}

func GetTokaiRandomSection() (string, string) {
	stations1 := tokaiStations
	rand.Shuffle(len(stations1), func(i, j int) { stations1[i], stations1[j] = stations1[j], stations1[i] })
	stations2 := stations1[1:]
	rand.Shuffle(len(stations2), func(i, j int) { stations2[i], stations2[j] = stations2[j], stations2[i] })

	return stations1[0], stations2[0]
}

func GetRandomUser() (*isutrain.User, error) {
	emailRandomStr, err := util.SecureRandomStr(10)
	if err != nil {
		return nil, bencherror.NewCriticalError(err, "ユーザを作成できません. 運営に確認をお願いいたします")
	}
	passwdRandomStr, err := util.SecureRandomStr(20)
	if err != nil {
		return nil, bencherror.NewCriticalError(err, "ユーザを作成できません. 運営に確認をお願いいたします")
	}
	return &isutrain.User{
		Email:    fmt.Sprintf("%s@example.com", emailRandomStr),
		Password: passwdRandomStr,
	}, nil
}

func GetRandomCarNumber(trainClass, seatClass string) int {
	l := []int{}

	for carNum := 1; carNum <= 16; carNum++ {
		if GetSeatClass(trainClass, carNum) == seatClass {
			l = append(l, carNum)
		}
	}

	idx := rand.Intn(len(l))
	return l[idx]
}
