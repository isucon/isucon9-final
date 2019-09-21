package endpoint

import "sync/atomic"

type EndpointIdx int

type Endpoint struct {
	path   string
	weight int
	count  int64
}

func (e *Endpoint) inc() {
	atomic.AddInt64(&e.count, 1)
}

func (e *Endpoint) score() int64 {
	return int64(e.weight) * e.count
}
