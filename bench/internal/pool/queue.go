package pool

import (
	"sync"

	"github.com/chibiegg/isucon9-final/bench/isutrain"
)

type queue struct {
	mu       sync.Mutex
	cap      int
	sessions []*isutrain.Session
}

func newQueue(cap int) *queue {
	return &queue{
		cap:      cap,
		sessions: make([]*isutrain.Session, cap),
	}
}

func (q *queue) Enqueue(sess *isutrain.Session) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.sessions = append(q.sessions, sess)
}

func (q *queue) Dequeue() (sess *isutrain.Session) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.sessions) == 0 {
		return nil
	}

	sess = q.sessions[0]
	q.sessions = q.sessions[1:]

	return
}
