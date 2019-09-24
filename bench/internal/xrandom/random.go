package xrandom

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/chibiegg/isucon9-final/bench/internal/util"
)

// FIXME: チューニングポイントに関わる値について、公平性を保てるように結果的には同じものを舐めるようにしたい
// 固定のものを用意し、起動時にshuffle、それでアクセスする？（ただ、これだと散らばりが悪いと最初になかなかスコアが上がらないチームとそうでないのが出たりする）

func GetRandomStations() string {
	idx := rand.Intn(len(stations))
	return stations[idx]
}

func GetRandomTrainClass() string {
	idx := rand.Intn(len(trainClasses))
	return trainClasses[idx]
}

func GetRandomUseAt() time.Time {
	startTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.Local)
	days := rand.Intn(366)
	return startTime.AddDate(0, 0, days)
}

func GetRandomSection() (string, string) {
	stations1 := stations
	rand.Shuffle(len(stations1), func(i, j int) { stations1[i], stations1[j] = stations1[j], stations1[i] })
	stations2 := stations1[1:]
	rand.Shuffle(len(stations2), func(i, j int) { stations2[i], stations2[j] = stations2[j], stations2[i] })

	return stations1[0], stations2[0]
}

func GetRandomUser() (*User, error) {
	emailRandomStr, err := util.SecureRandomStr(10)
	if err != nil {
		return nil, err
	}
	passwdRandomStr, err := util.SecureRandomStr(20)
	if err != nil {
		return nil, err
	}
	return &User{
		Email:    fmt.Sprintf("%s@example.com", emailRandomStr),
		Password: passwdRandomStr,
	}, nil
}
