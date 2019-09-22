package endpoint

import "sync/atomic"

type EndpointIdx int

type Endpoint struct {
	path       string
	weight     int
	count      int64
	extraScore int64
}

func (e *Endpoint) inc() {
	atomic.AddInt64(&e.count, 1)
}

func (e *Endpoint) addExtraScore(extraScore int64) {
	atomic.AddInt64(&e.extraScore, extraScore)
}

func (e *Endpoint) score() int64 {
	return int64(e.weight)*e.count + e.extraScore
}
