package isutrain

import (
	"errors"
	"time"
)

var (
	ErrNoSeats = errors.New("座席がありません")
)

// TODO: Statusのconst

// ReservationStatus は予約状態です
type ReservationStatus string

const (
	// Pending は予約が確定していない状態です
	Pending ReservationStatus = "pending"
	// Ok は予約が確定した状態です
	Ok ReservationStatus = "ok"
)

type ReserveResponse struct {
	ReservationID int  `json:"reservation_id"`
	IsOk          bool `json:"is_ok"`
}

type ReserveRequest struct {
	// Train構造体
	TrainClass string `json:"train_class"`
	TrainName  string `json:"train_name"`
	// TrainSeat構造体
	SeatClass string     `json:"seat_class"`
	Seats     TrainSeats `json:"seats"`
	// それ以外
	//// 区間
	Departure string `json:"departure"`
	Arrival   string `json:"arrival"`
	// 日付
	Date   time.Time `json:"date"`
	CarNum int       `json:"car_number"`
	Child  int       `json:"child"`
	Adult  int       `json:"adult"`
}

// ReservationResponse は
type ReservationResponse struct {
	ReservationID int        `json:"reservation_id"`
	Date          string     `json:"date"`
	TrainClass    string     `json:"train_class"`
	TrainName     string     `json:"train_name"`
	CarNum        int        `json:"car_number"`
	SeatClass     string     `json:"seat_class"`
	Amount        int64      `json:"amount"`
	Adult         int        `json:"adult"`
	Child         int        `json:"child"`
	Departure     string     `json:"departure"`
	Arrival       string     `json:"arrival"`
	DepartureTime string     `json:"departure_time"`
	ArrivalTime   string     `json:"arrival_time"`
	Seats         TrainSeats `json:"seats"`
}

type CommitReservationRequest struct {
	ReservationID int    `json:"reservation_id"`
	CardToken     string `json:"card_token"`
}

type ShowReservationResponse struct {
	ReservationID int   `json:"reservation_id"`
	Amount        int64 `json:"amount"`
}
