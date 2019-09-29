package isutrain

import (
	"errors"
	"sync"
	"time"

	"github.com/chibiegg/isucon9-final/bench/internal/isutraindb"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// FIXME: 料金計算
//距離運賃(円) * 期間倍率(繁忙期なら2倍等) * 車両クラス倍率(急行・各停等) * 座席クラス倍率(プレミアム・指定席・自由席)

var (
	ErrCommitReservation = errors.New("予約の確定に失敗しました")
	ErrCancelReservation = errors.New("予約のキャンセルに失敗しました")
	ErrCanNotReserve     = errors.New("予約済みの座席が指定されたため予約できません")
)

// NOTE: 区間の考慮
// * 発駅が範囲内に入っている
// * 着駅が範囲内に入って入る
// * 発駅、着駅が範囲外で、ちょうど覆って入る

// TODO: 予約情報を覚えておいて、座席予約の時に
// 取れるはずの予約を誤魔化されてないかちゃんとチェックする

// TODO: 決済情報のバリデーションができるようにする

// TODO: 未予約の予約を取得できるものがあるといい

type ReservationCacheEntry struct {
	// ユーザ情報
	User *User

	// 予約情報
	ID int

	Date                  time.Time
	Departure, Arrival    string
	TrainClass, TrainName string
	CarNum                int

	SeatClass string
	Seats     TrainSeats

	Adult, Child int

	UseAt time.Time
}

// Fare は大人１人あたりの運賃を算出します
func (r *ReservationCacheEntry) Fare() (int, error) {
	var (
		distanceFare, err = isutraindb.GetDistanceFare(r.Departure, r.Arrival)
		fareMultiplier    = isutraindb.GetFareMultiplier(r.TrainClass, r.SeatClass, r.UseAt)
	)
	if err != nil {
		return -1, err
	}

	lgr := zap.S()
	lgr.Infow("運賃取得情報",
		"reservation_id", r.ID,
		"departure", r.Departure,
		"arrival", r.Arrival,
		"train_class", r.TrainClass,
		"seat_class", r.SeatClass,
		"use_at", r.UseAt,
	)
	lgr.Infow("運賃",
		"distance_fare", distanceFare,
		"fare_multiplier", fareMultiplier,
	)

	return int(float64(distanceFare) * fareMultiplier), nil
}

// Amount は、大人と子供を考慮し、合計の運賃を算出します
func (r *ReservationCacheEntry) Amount() (int, error) {
	fare, err := r.Fare()
	if err != nil {
		return -1, err
	}

	var (
		adultFare = fare * r.Adult
		// 子供は半額
		childFare = (fare * r.Child) / 2
	)

	lgr := zap.S()
	lgr.Infow("Amount",
		"adult", r.Adult,
		"child", r.Child,
		"adult_fare", adultFare,
		"child_fare", childFare,
	)
	return adultFare + childFare, nil
}

var (
	// RCache は、webappの予約に関する情報が適切か検証するために用いられるキャッシュです
	ReservationCache = newReservationCache()
)

type reservationCache struct {
	mu sync.RWMutex
	// reservationID -> ReservationCacheEntry
	reservations map[int]*ReservationCacheEntry
	// reservations []*ReservationCacheEntry
}

func newReservationCache() *reservationCache {
	return &reservationCache{
		reservations: map[int]*ReservationCacheEntry{},
		// reservations: []*ReservationCacheEntry{},
	}
}

func (r *reservationCache) Reservation(reservationID int) (*ReservationCacheEntry, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	reservation, ok := r.reservations[reservationID]
	if !ok {
		return nil, false
	}

	return reservation, true
}

// 予約可能判定
// NOTE: この予約が可能か？を判定する必要があるので、リクエストを受け取り、複数のSeatのどれか１つでも含まれていればNGとする
func (r *reservationCache) CanReserve(req *ReserveRequest) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	lgr := zap.S()

	canReserveWithOverwrap := func(reservation *ReservationCacheEntry) (bool, error) {
		reqKudari, err := isKudari(req.Departure, req.Arrival)
		if err != nil {
			lgr.Warnf("予約可能判定の 下り判定でエラーが発生: %+v", err)
			return false, err
		}

		resKudari, err := isKudari(reservation.Departure, reservation.Arrival)
		if err != nil {
			lgr.Warnf("予約可能判定の 下り判定でエラーが発生: %+v", err)
			return false, err
		}

		// 上りと下りが一致しなければ、予約として被らない
		if reqKudari != resKudari {
			return true, nil
		}

		if reqKudari {
			overwrap, err := isKudariOverwrap(reservation.Departure, reservation.Arrival, req.Departure, req.Arrival)
			if err != nil {
				lgr.Warnf("予約可能判定の 区間重複判定呼び出しでエラーが発生: %+v", err)
				return false, err
			}

			if overwrap {
				return false, nil
			}
		} else {
			// NOTE: 下りベースの判定関数を用いるため、上りの場合は乗車・降車を入れ替えて渡す
			overwrap, err := isKudariOverwrap(reservation.Arrival, reservation.Departure, req.Arrival, req.Departure)
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

	eg := errgroup.Group{}
	for _, r := range r.reservations {
		reservation := r
		eg.Go(func() error {
			if !req.Date.Equal(reservation.Date) {
				return nil
			}
			if req.TrainClass != reservation.TrainClass || req.TrainName != reservation.TrainName {
				return nil
			}
			// 区間
			if ok, err := canReserveWithOverwrap(reservation); ok || err != nil {
				return err
			}
			// 車両
			if req.CarNum != reservation.CarNum {
				return nil
			}
			// 座席
			for _, seat := range req.Seats {
				for _, existSeat := range reservation.Seats {
					if seat.Row == existSeat.Row && seat.Column == existSeat.Column {
						return ErrCanNotReserve
					}
				}
			}

			return nil
		})
	}
	if err := eg.Wait(); errors.Is(err, ErrCanNotReserve) {
		return false, nil
	} else if err != nil {
		lgr.Warnf("予約可能判定の予約チェックループにて、区間重複チェック呼び出しエラーが発生: %+v", err)
		return false, err
	}

	return true, nil
}

func (r *reservationCache) Add(user *User, req *ReserveRequest, reservationID int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// TODO: webappから意図的にreservationIDを細工して変に整合性つけることができないか考える
	r.reservations[reservationID] = &ReservationCacheEntry{
		User:       user,
		ID:         reservationID,
		Date:       req.Date,
		Departure:  req.Departure,
		Arrival:    req.Arrival,
		TrainClass: req.TrainClass,
		TrainName:  req.TrainName,
		CarNum:     req.CarNum,
		SeatClass:  req.SeatClass,
		Seats:      req.Seats,
		Adult:      req.Adult,
		Child:      req.Child,
		UseAt:      req.Date,
	}
}

func (r *reservationCache) Cancel(reservationID int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.reservations, reservationID)

	return nil
}

func (r *reservationCache) Range(f func(reservation *ReservationCacheEntry)) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, reservation := range r.reservations {
		f(reservation)
	}
}
