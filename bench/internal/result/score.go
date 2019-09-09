package result

import "sync/atomic"

type totalScore struct {
	totalScore uint64
}

var total = new(totalScore)

// IncScore はスコアを加算します
func IncScore(score int) {
	atomic.AddUint64(&total.totalScore, uint64(score))
}

// DecScore はスコアを減算します
func DecScore(score int) {
	atomic.AddUint64(&total.totalScore, -uint64(score))
}
