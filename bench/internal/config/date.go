package config

import (
	"errors"
	"time"

	"go.uber.org/zap"
)

var (
	ErrAvailReserveDaysNotSpecified = errors.New("予約可能日数が指定されていません")
	ErrAvailReserveDaysTooLarge     = errors.New("予約日数が翌年の日付を含んでいます")
)

var (
	// initializeで設定される、予約日数
	AvailReserveDays int

	// 予約可能日数
	maxReserveDays = 159
)

var (
	// 予約受付開始日
	ReservationStartDate time.Time = time.Date(2020, 3, 1, 0, 0, 0, 0, time.Local)
	// 予約受付終了日
	ReservationEndDate time.Time
)

// 予約日数設定
func SetAvailReserveDays(days int) error {
	lgr := zap.S()

	if days == 0 {
		lgr.Warnf("予約日数が指定されていません")
		return ErrAvailReserveDaysNotSpecified
	}

	if days > maxReserveDays {
		lgr.Warnf("予約日数が予約可能日数を超過: 予約日数=%d, 予約可能日数=%d", AvailReserveDays, maxReserveDays)
		return ErrAvailReserveDaysTooLarge
	}

	AvailReserveDays = days
	ReservationEndDate = ReservationStartDate.Add(time.Duration(days) * 24 * time.Hour)

	lgr.Infow("予約日数を設定",
		"指定された予約日数", AvailReserveDays,
		"予約可能日数", maxReserveDays,
		"予約受付開始日", ReservationStartDate,
		"予約受付終了日", ReservationEndDate,
	)

	// FIXME: 日数に応じた負荷レベルを設定

	return nil
}

var (
	OlympicStartDate = time.Date(2020, 7, 24, 0, 0, 0, 0, time.Local)
	OlympicEndDate   = time.Date(2020, 8, 9, 0, 0, 0, 0, time.Local)
)
