package cache

import (
	"errors"
	"log"
	"sync"
	"time"

	"github.com/chibiegg/isucon9-final/bench/internal/bencherror"
	"github.com/chibiegg/isucon9-final/bench/isutrain"
)

var (
	ErrCommitReservation = errors.New("予約の確定に失敗しました")
	ErrCancelReservation = errors.New("予約のキャンセルに失敗しました")
)

// FIXME: 区間の考慮
// * 発駅が範囲内に入っている
// * 着駅が範囲内に入って入る
// * 発駅、着駅が範囲外で、ちょうど覆って入る

// TODO: 予約情報を覚えておいて、座席予約の時に
// 取れるはずの予約を誤魔化されてないかちゃんとチェックする

// FIXME: 決済情報のバリデーションができるようにする

// FIXME: 未予約の予約を取得できるものがあるといい

type ReservationResult struct {
	keys                []SeatMapKey
	origin, destination string
	amount              int64
}

func NewReservationResult() *ReservationResult {
	return &ReservationResult{
		keys: []SeatMapKey{},
	}
}

type SeatMapKey struct {
	Date                  time.Time
	TrainClass, TrainName string
	CarNum                int
	Row                   int
	Column                string
}

type Reservation struct {
	ID     string
	Amount int64

	// 検索条件周り
	Date                  time.Time
	Origin, Destination   string
	TrainClass, TrainName string
	CarNum                int

	Seats isutrain.TrainSeats
}

type ReservationCache struct {
	mu           sync.RWMutex
	reservations []*Reservation
}

func NewReservationMem() *ReservationCache {
	return &ReservationCache{
		reservations: []*Reservation{},
	}
}

// 予約可能判定
// NOTE: この予約が可能か？を判定する必要があるので、リクエストを受け取り、複数のSeatのどれか１つでも含まれていればNGとする
func (r *ReservationCache) CanReserve(req *isutrain.ReservationRequest) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	canReserveWithOverwrap := func(reservation *Reservation) (bool, error) {
		nobori, err := isNobori(req.Origin, req.Destination)
		if err != nil {
			log.Println("上りエラー")
			log.Println(err)
			return false, err
		}

		if nobori {
			log.Println("上り")
			overwrap, err := isOverwrap(req.Origin, req.Destination, reservation.Origin, reservation.Destination)
			if err != nil {
				return false, err
			}

			if overwrap {
				return false, nil
			}
		} else {
			log.Println("下り")
			overwrap, err := isOverwrap(reservation.Origin, reservation.Destination, reservation.Origin, reservation.Destination)
			if err != nil {
				return false, err
			}

			if overwrap {
				return false, nil
			}
		}

		return true, nil
	}

	log.Println("iterate reservations")
	for _, reservation := range r.reservations {
		log.Println("look at a reservation")
		if !req.Date.Equal(reservation.Date) {
			continue
		}
		log.Println("date checking")
		log.Printf("req trainclass=%s, trainname=%s | reservation trainclass=%s, trainname=%s\n", req.TrainClass, req.TrainName, reservation.TrainClass, reservation.TrainName)
		if req.TrainClass != reservation.TrainClass || req.TrainName != reservation.TrainName {
			continue
		}
		log.Printf("train checking")
		// 区間
		log.Printf("req origin=%s destination=%s | reservation origin=%s destination=%s\n", req.Origin, req.Destination, reservation.Origin, reservation.Destination)
		if ok, err := canReserveWithOverwrap(reservation); ok {
			if err != nil {
				log.Printf("overwrap error: %+v\n", err)
				return false, err
			}
			continue
		}
		log.Println("overwrap checking")
		// 車両
		if req.CarNum != reservation.CarNum {
			continue
		}
		log.Println("carnum checking")
		// 座席
		for _, seat := range req.Seats {
			for _, existSeat := range reservation.Seats {
				if seat.Row == existSeat.Row && seat.Column == existSeat.Column {
					return false, nil
				}
			}
		}
		log.Println("seat is not same. ok.")
	}

	return true, nil
}

func (r *ReservationCache) Add(req *isutrain.ReservationRequest, reservationID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// TODO: webappから意図的にreservationIDを細工して変に整合性つけることができないか考える
	r.reservations = append(r.reservations, &Reservation{
		ID:          reservationID,
		Date:        req.Date,
		Origin:      req.Origin,
		Destination: req.Destination,
		TrainClass:  req.TrainClass,
		TrainName:   req.TrainName,
		CarNum:      req.CarNum,
		Seats:       req.Seats,
	})
}

func (r *ReservationCache) Commit(reservationID string, amount int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, reservation := range r.reservations {
		if reservation.ID == reservationID {
			reservation.Amount = amount
			return nil
		}
	}

	return bencherror.NewApplicationError(ErrCommitReservation, "予約が存在しません")
}

func (r *ReservationCache) Cancel(reservationID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for idx, reservation := range r.reservations {
		if reservation.ID == reservationID {
			r.reservations = append(r.reservations[:idx], r.reservations[idx+1:]...)
			return nil
		}
	}

	return bencherror.NewApplicationError(ErrCancelReservation, "予約が存在しません")
}
