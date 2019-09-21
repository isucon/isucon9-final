package memory

import (
	"sync"
	"time"

	"github.com/chibiegg/isucon9-final/bench/isutrain"
)

// TODO: 予約情報を覚えておいて、座席予約の時に
// 取れるはずの予約を誤魔化されてないかちゃんとチェックする

// FIXME: 決済情報のバリデーションができるようにする

type SeatMapKey struct {
	Date                  time.Time
	TrainClass, TrainName string
	CarNum                int
	Row                   int
	Column                string
}

func NewSeatMapKeys(req *isutrain.ReservationRequest) []SeatMapKey {
	keys := []SeatMapKey{}
	for _, seat := range req.Seats {
		keys = append(keys, SeatMapKey{
			Date:       req.Date,
			TrainClass: req.TrainClass,
			TrainName:  req.TrainName,
			CarNum:     req.CarNum,
			Row:        seat.Row,
			Column:     seat.Column,
		})
	}

	return keys
}

type ReservationMem struct {
	mu      sync.RWMutex
	seatMap map[SeatMapKey]struct{}

	reservations []isutrain.SeatReservation
}

func NewReservationMem() *ReservationMem {
	return &ReservationMem{
		seatMap: make(map[SeatMapKey]struct{}, 10000),
	}
}

func (r *ReservationMem) Add(key SeatMapKey) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.seatMap[key] = struct{}{}
}

func (r *ReservationMem) AddFromRequest(req *isutrain.ReservationRequest) {
	for _, seat := range req.Seats {
		r.Add(SeatMapKey{
			Date:       req.Date,
			TrainClass: req.TrainClass,
			TrainName:  req.TrainName,
			CarNum:     req.CarNum,
			Row:        seat.Row,
			Column:     seat.Column,
		})
	}
}

// 予約可能判定
// NOTE: この予約が可能か？を判定する必要があるので、リクエストを受け取り、複数のSeatのどれか１つでも含まれていればNGとする
func (r *ReservationMem) CanReserve(keys []SeatMapKey) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, key := range keys {
		if _, ok := r.seatMap[key]; ok {
			return false
		}
	}

	return true
}
