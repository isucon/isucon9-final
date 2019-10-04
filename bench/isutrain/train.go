package isutrain

import (
	"math"
)

// /api/train/search のレスポンス形式
type (
	Train struct {
		Class            string            `json:"train_class"`
		Name             string            `json:"train_name"`
		Start            string            `json:"start"`
		Last             string            `json:"last"`
		Departure        string            `json:"departure"`
		Arrival          string            `json:"arrival"`
		DepartedAt       string            `json:"departure_time"`
		ArrivedAt        string            `json:"arrival_time"`
		SeatAvailability map[string]string `json:"seat_availability"`
		FareInformation  map[string]int    `json:"seat_fare"`
	}

	SearchTrainsResponse []*Train
)

// /api/train/seats のレスポンス形式
type (
	SearchTrainSeatsResponse struct {
		Date       string     `json:"date"`
		TrainClass string     `json:"train_class"`
		TrainName  string     `json:"train_name"`
		CarNumber  int        `json:"car_number"`
		Seats      TrainSeats `json:"seats"`
		Cars       TrainCars  `json:"cars"`
	}

	TrainSeats []*TrainSeat
	TrainSeat  struct {
		Row           int    `json:"row"`
		Column        string `json:"column"`
		Class         string `json:"class"`
		IsSmokingSeat bool   `json:"is_smoking_seat,omitempty"`
		IsOccupied    bool   `json:"is_occupied,omitempty"`
	}

	TrainCars []*TrainCar
	TrainCar  struct {
		CarNumber int    `json:"car_number"`
		SeatClass string `json:"seat_class"`
	}
)

func (seats TrainSeats) IsSame(gotSeats TrainSeats) bool {
	if len(seats) != len(gotSeats) {
		return false
	}

	for i := 0; i < len(seats); i++ {
		var (
			seat    = seats[i]
			gotSeat = gotSeats[i]
		)
		if *seat != *gotSeat {
			return false
		}
	}

	return true
}

func (cars TrainCars) IsSame(gotCars TrainCars) bool {
	if len(cars) != len(gotCars) {
		return false
	}

	for i := 0; i < len(cars); i++ {
		var (
			car    = cars[i]
			gotCar = gotCars[i]
		)
		if *car != *gotCar {
			return false
		}
	}

	return true
}

func IsValidTrainSeatColumn(seatColumn string) bool {
	switch TrainSeatColumn(seatColumn) {
	case ColumnA, ColumnB, ColumnC, ColumnD, ColumnE:
		return true
	default:
		return false
	}
}

type TrainSeatColumn string

const (
	ColumnA TrainSeatColumn = "A"
	ColumnB                 = "B"
	ColumnC                 = "C"
	ColumnD                 = "D"
	ColumnE                 = "E"
)

func (c TrainSeatColumn) Int() int {
	switch c {
	case ColumnA:
		return 0
	case ColumnB:
		return 1
	case ColumnC:
		return 2
	case ColumnD:
		return 3
	case ColumnE:
		return 4
	default:
		return 100
	}
}

func (c TrainSeatColumn) IsNeighbor(c2 TrainSeatColumn) bool {
	return math.Abs(float64(c.Int()-c2.Int())) == 1.0
}

type SeatAvailability string

const (
	SaPremium       SeatAvailability = "premium"
	SaPremiumSmoke  SeatAvailability = "premium_smoke"
	SaReserved      SeatAvailability = "reserved"
	SaReservedSmoke SeatAvailability = "reserved_smoke"
	SaNonReserved   SeatAvailability = "non_reserved"
)

func (sa SeatAvailability) String() string {
	return string(sa)
}

func (sa SeatAvailability) Value() string {
	switch sa {
	case SaPremium, SaReservedSmoke, SaNonReserved:
		return "○"
	case SaPremiumSmoke:
		return "×"
	case SaReserved:
		return "△"
	default:
		return ""
	}
}

type FareInformation string

const (
	FiPremium       FareInformation = "premium"
	FiPremiumSmoke  FareInformation = "premium_smoke"
	FiReserved      FareInformation = "reserved"
	FiReservedSmoke FareInformation = "reserved_smoke"
	FiNonReserved   FareInformation = "non_reserved"
)

func (fi FareInformation) String() string {
	return string(fi)
}

func (fi FareInformation) Value() int {
	switch fi {
	case FiPremium:
		return 24000
	case FiPremiumSmoke:
		return 24500
	case FiReserved:
		return 19000
	case FiReservedSmoke:
		return 19500
	case FiNonReserved:
		return 15000
	default:
		return -1
	}
}

func IsValidTrainClass(trainClass string) bool {
	switch trainClass {
	case "遅いやつ", "中間", "最速":
		return true
	default:
		return false
	}
}

func IsValidSeatClass(seatClass string) bool {
	switch seatClass {
	case "premium", "reserved", "non-reserved":
		return true
	default:
		return false
	}
}

func IsValidCarNumber(carNum int) bool {
	return carNum >= 1 && carNum < 17
}
