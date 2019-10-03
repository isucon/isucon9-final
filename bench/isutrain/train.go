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

func (seats TrainSeats) GetNeighborSeatsMultiplier() float64 {
	m := map[int][]TrainSeatColumn{}
	for _, seat := range seats {
		if _, ok := m[seat.Row]; !ok {
			m[seat.Row] = []TrainSeatColumn{}
		}
		m[seat.Row] = append(m[seat.Row], TrainSeatColumn(seat.Column))
	}

	var max float64
	for _, columns := range m {
		var neighborCount int
		if len(columns) > 1 {
			for i := 1; i < len(columns); i++ {
				if columns[i-1].IsNeighbor(columns[i]) {
					neighborCount++
				}
			}
		}
		max = math.Max(max, float64(neighborCount))
	}

	switch int(max) + 1 {
	case 1:
		return 1
	case 2:
		return 1.2
	case 3:
		return 1.4
	case 4:
		return 1.9
	case 5:
		return 2.0
	default:
		return 1
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
