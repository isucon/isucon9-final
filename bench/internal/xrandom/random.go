package xrandom

import "math/rand"

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
