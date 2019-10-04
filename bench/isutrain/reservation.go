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
		ReservationID int              `json:"reservation_id"`
		Date          string           `json:"date"`
		TrainClass    string           `json:"train_class"`
		TrainName     string           `json:"train_name"`
		CarNumber     int              `json:"car_number"`
		SeatClass     string           `json:"seat_class"`
		Amount        int              `json:"amount"`
		Adult         int              `json:"adult"`
		Child         int              `json:"child"`
		Departure     string           `json:"departure"`
		Arrival       string           `json:"arrival"`
		DepartureTime string           `json:"departure_time"`
		ArrivalTime   string           `json:"arrival_time"`
		Seats         ReservationSeats `json:"seats"`
	}

	ReservationSeat struct {
		ReservationID int    `json:"reservation_id,omitempty" db:"reservation_id"`
		CarNumber     int    `json:"car_number,omitempty" db:"car_number"`
		SeatRow       int    `json:"seat_row" db:"seat_row"`
		SeatColumn    string `json:"seat_column" db:"seat_column"`
	}
	ReservationSeats []*ReservationSeat
)

// 隣り合うパターンを見つけたら加算する
func (seats ReservationSeats) GetNeighborSeatsBonus() int {
	m := map[int]int{}
	for _, seat := range seats {
		if _, ok := m[seat.SeatRow]; !ok {
			m[seat.SeatRow] = 0
		}
		if !IsValidTrainSeatColumn(seat.SeatColumn) {
			return 0
		}
		column := TrainSeatColumn(seat.SeatColumn)

		shiftCnt := uint64(column.Int())
		m[seat.SeatRow] |= (1 << shiftCnt)
	}

	var score int
	addScore := func(bits int) {
		var (
			five  = 25
			four  = 20
			three = 15
			two   = 10
		)
		switch bits {
		// 5つ
		case 31:
			score += five
			return
		// 4つ
		case 30, 15:
			score += four
			return
		// 3つ
		case 28, // 11100
			29, // 11101
			14, // 01110
			7,  // 00111
			23: // 10111
			score += three
			return
		}

		// ２つの場合はずらしながら、見つけたら加点
		check := 3 // 00011
		for i := bits; i > 0; i >>= 1 {
			if i&check == check {
				// 隣り合う２つの席を見つけた
				score += two
				i &= ^check
			}
		}
	}

	// 加算処理
	for _, bits := range m {
		addScore(bits)
	}

	return score
}

// 予約確定API
type (
	CommitReservationRequest struct {
		ReservationID int    `json:"reservation_id"`
		CardToken     string `json:"card_token"`
	}
	CommitReservationResponse struct {
		IsOK bool `json:"is_ok"`
	}
)

// 予約キャンセルAPI
type (
	CancelReservationResponse struct {
		IsOK bool `json:"is_ok"`
	}
)
