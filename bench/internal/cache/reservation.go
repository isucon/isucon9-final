package cache

import (
	"errors"
	"sync"
	"time"

	"github.com/chibiegg/isucon9-final/bench/internal/bencherror"
	"github.com/chibiegg/isucon9-final/bench/isutrain"
	"go.uber.org/zap"
)

// FIXME: 料金計算
//距離運賃(円) * 期間倍率(繁忙期なら2倍等) * 車両クラス倍率(急行・各停等) * 座席クラス倍率(プレミアム・指定席・自由席)

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

	lgr := zap.S()

	canReserveWithOverwrap := func(reservation *Reservation) (bool, error) {
		reqKudari, err := isKudari(req.Departure, req.Arrival)
		if err != nil {
			lgr.Warnf("予約可能判定の 下り判定でエラーが発生: %+v", err)
			return false, err
		}

		resKudari, err := isKudari(reservation.Origin, reservation.Destination)
		if err != nil {
			lgr.Warnf("予約可能判定の 下り判定でエラーが発生: %+v", err)
			return false, err
		}

		// 上りと下りが一致しなければ、予約として被らない
		if reqKudari != resKudari {
			return true, nil
		}

		if reqKudari {
			overwrap, err := isKudariOverwrap(reservation.Origin, reservation.Destination, req.Departure, req.Arrival)
			if err != nil {
				lgr.Warnf("予約可能判定の 区間重複判定呼び出しでエラーが発生: %+v", err)
				return false, err
			}

			if overwrap {
				return false, nil
			}
		} else {
			// NOTE: 下りベースの判定関数を用いるため、上りの場合は乗車・降車を入れ替えて渡す
			overwrap, err := isKudariOverwrap(reservation.Destination, reservation.Origin, req.Arrival, req.Departure)
			if err != nil {
				lgr.Warnf("予約可能判定の 区間重複判定呼び出しでエラーが発生: %+v", err)
				return false, err
			}

			if overwrap {
				return false, nil
			}
		}

		return true, nil
	}

	for _, reservation := range r.reservations {
		if !req.Date.Equal(reservation.Date) {
			continue
		}
		if req.TrainClass != reservation.TrainClass || req.TrainName != reservation.TrainName {
			continue
		}
		// 区間
		if ok, err := canReserveWithOverwrap(reservation); ok {
			if err != nil {
				lgr.Warnf("予約可能判定の予約チェックループにて、区間重複チェック呼び出しエラーが発生: %+v", err)
				return false, err
			}
			continue
		} else if err != nil {
			lgr.Warnf("予約可能判定の予約チェックループにて、区間重複チェック呼び出しエラーが発生: %+v", err)
		}
		// 車両
		if req.CarNum != reservation.CarNum {
			continue
		}
		// 座席
		for _, seat := range req.Seats {
			for _, existSeat := range reservation.Seats {
				if seat.Row == existSeat.Row && seat.Column == existSeat.Column {
					return false, nil
				}
			}
		}
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
		Origin:      req.Departure,
		Destination: req.Arrival,
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
