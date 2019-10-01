package isutrain

import (
	"math"
)

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

// TrainSeat は座席です
type TrainSeat struct {
	ReservationID int `json:"reservation_id,omitempty"`
	CarNumber     int `json:"car_number,omitempty"`
	// Row は席位置の列です(ex. １列)
	Row int `json:"row"`
	// Column は席位置の行です (ex. A行)
	Column string `json:"column"`
	// IsSmokingSeat 喫煙所が近くにあるかどうかのフラグです
	IsSmokingSeat bool `json:"is_smoking_seat,omitempty"`
	// IsOccupied は 予約済みであるか否かを示します
	IsOccupied bool `json:"is_occupied,omitempty"`
}

// TrainSeats は座席一覧です
type TrainSeats []*TrainSeat

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

type TrainCar struct {
	CarNumber int    `json:"car_number"`
	SeatClass string `json:"seat_class"`
}

type TrainCars []*TrainCar

// NOTE:  列車検索API  use_at=<RFC3339形式の時刻>&train_class=<>&from=<>&to=<>
// * 流れ
//   * 列車マスタからuse_atに合致するレコードを引っ張る
//     * 各レコードについて、駅マスタから距離を取得し、のぼり下りを考慮して駅名を引っ張る
//     * 引っ張れた駅名をイテレーションし、発駅、着駅を経路に持っているか調べ上げる
//     * 発駅、着駅を含むなら、列車リストに列車を追加
//   * 列車リストを返す(TranSearchResponse, 未定義)
type Train struct {
	// Class は列車種別です
	Class string `json:"train_class"`
	// Name は列車名です
	Name string `json:"train_name"`
	// Start は始点駅IDです
	Start string `json:"start"`
	// EndStation は終点駅IDです
	Last             string            `json:"last"`
	Departure        string            `json:"departure"`
	Arrival          string            `json:"arrival"`
	DepartedAt       string            `json:"departure_time"`
	ArrivedAt        string            `json:"arrival_time"`
	SeatAvailability map[string]string `json:"seat_availability"`
	FareInformation  map[string]int    `json:"seat_fare"`
}

type Trains []*Train

// NOTE: 座席API use_at=<RFC3339形式の時刻>&train_class=<列車クラス>&train_name=<列車名>&car_num=<>
// * 流れ
//   * 座席マスタから列車種別、車両番号に一致するレコードを引っ張る
//     * 隠れコードについて、席予約から条件に合致するレコードの数を取得
//     * １つ見つかったら、予約済み(IsOccupied) フラグを立てる
//     * 席は予約済みかそうでないかにかかわらず結果として追加
//     * CarInformationを返す
type TrainSeatSearchResponse struct {
	Date       string     `json:"date"`
	TrainClass string     `json:"train_class"`
	TrainName  string     `json:"train_name"`
	CarNumber  int        `json:"car_number"`
	Seats      TrainSeats `json:"seats"`
}
