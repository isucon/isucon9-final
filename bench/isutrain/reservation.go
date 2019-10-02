package isutrain

import (
	"errors"
)

var (
	ErrNoSeats = errors.New("座席がありません")
)

// 予約API /api/train/reserve
type (
	ReserveRequest struct {
		Date          string     `json:"date"`
		TrainName     string     `json:"train_name"`
		TrainClass    string     `json:"train_class"`
		CarNum        int        `json:"car_number"`
		IsSmokingSeat bool       `json:"is_smoking_seat"`
		SeatClass     string     `json:"seat_class"`
		Departure     string     `json:"departure"`
		Arrival       string     `json:"arrival"`
		Child         int        `json:"child"`
		Adult         int        `json:"adult"`
		Column        string     `json:"Column"`
		Seats         TrainSeats `json:"seats"`
	}

	ReserveResponse struct {
		ReservationID int  `json:"reservation_id"`
		Amount        int  `json:"amount"`
		IsOk          bool `json:"is_ok"`
	}
)

// 予約詳細・列挙API
type (
	ShowReservationResponse *Reservation

	ListReservationsResponse []*Reservation

	Reservation struct {
		ReservationID int                `json:"reservation_id"`
		Date          string             `json:"date"`
		TrainClass    string             `json:"train_class"`
		TrainName     string             `json:"train_name"`
		CarNumber     int                `json:"car_number"`
		SeatClass     string             `json:"seat_class"`
		Amount        int                `json:"amount"`
		Adult         int                `json:"adult"`
		Child         int                `json:"child"`
		Departure     string             `json:"departure"`
		Arrival       string             `json:"arrival"`
		DepartureTime string             `json:"departure_time"`
		ArrivalTime   string             `json:"arrival_time"`
		Seats         []*ReservationSeat `json:"seats"`
	}

	ReservationSeat struct {
		ReservationID int    `json:"reservation_id,omitempty" db:"reservation_id"`
		CarNumber     int    `json:"car_number,omitempty" db:"car_number"`
		SeatRow       int    `json:"seat_row" db:"seat_row"`
		SeatColumn    string `json:"seat_column" db:"seat_column"`
	}
)

// ReservationStatus は予約状態です
type ReservationStatus string

const (
	// Pending は予約が確定していない状態です
	Pending ReservationStatus = "pending"
	// Ok は予約が確定した状態です
	Ok ReservationStatus = "ok"
)

// ReservationResponse は
// type ReservationResponse struct {
// 	ReservationID int        `json:"reservation_id"`
// 	Date          string     `json:"date"`
// 	TrainClass    string     `json:"train_class"`
// 	TrainName     string     `json:"train_name"`
// 	CarNum        int        `json:"car_number"`
// 	SeatClass     string     `json:"seat_class"`
// 	Amount        int64      `json:"amount"`
// 	Adult         int        `json:"adult"`
// 	Child         int        `json:"child"`
// 	Departure     string     `json:"departure"`
// 	Arrival       string     `json:"arrival"`
// 	DepartureTime string     `json:"departure_time"`
// 	ArrivalTime   string     `json:"arrival_time"`
// 	Seats         TrainSeats `json:"seats"`
// }

type CommitReservationRequest struct {
	ReservationID int    `json:"reservation_id"`
	CardToken     string `json:"card_token"`
}
