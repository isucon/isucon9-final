package isutrain

// Train は新幹線列車です
type Train struct {
	// Class は列車種別です
	Class string `json:"class"`
	// Name は列車名です
	Name string `json:"name"`
	// Start は始点駅IDです
	Start int `json:"start"`
	// EndStation は終点駅IDです
	Last int `json:"last"`
}

// Trains は列車一覧です
type Trains []*Train

// TrainSeat は座席です
type TrainSeat struct {
	// Row は席位置の列です(ex. １列)
	Row int `json:"row"`
	// Column は席位置の行です (ex. A行)
	Column string `json:"column"`
	// Class は座席種別です
	Class string `json:"class"`
	// IsSmokingSeat 喫煙所が近くにあるかどうかのフラグです
	IsSmokingSeat bool `json:"is_smoking_seat"`
	// IsOccupied は 予約済みであるか否かを示します
	IsOccupied bool `json:"is_occupied"`
}

// TrainSeats は座席一覧です
// TODO: 列でざっくり指定、行でざっくり指定
type TrainSeats []*TrainSeat

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

// NOTE:  列車検索API  use_at=<RFC3339形式の時刻>&train_class=<>&from=<>&to=<>
// * 流れ
//   * 列車マスタからuse_atに合致するレコードを引っ張る
//     * 各レコードについて、駅マスタから距離を取得し、のぼり下りを考慮して駅名を引っ張る
//     * 引っ張れた駅名をイテレーションし、発駅、着駅を経路に持っているか調べ上げる
//     * 発駅、着駅を含むなら、列車リストに列車を追加
//   * 列車リストを返す(TranSearchResponse, 未定義)
/*
type TrainSearchResponse struct {
	Train
	Departure     string    `json:"departure"`
	Destination   string    `json:"destination"`
	DepartureTime time.Time `json:"departure_time"`
	ArrivalTime   time.Time `json:"arrival_time"`
	SeatAvailability map[string]string `json:"seat_availability"`
	Fare map[string]int `json:"seat_fare"`
}
*/

// NOTE: 座席API use_at=<RFC3339形式の時刻>&train_class=<列車クラス>&train_name=<列車名>&car_num=<>
// * 流れ
//   * 座席マスタから列車種別、車両番号に一致するレコードを引っ張る
//     * 隠れコードについて、席予約から条件に合致するレコードの数を取得
//     * １つ見つかったら、予約済み(IsOccupied) フラグを立てる
//     * 席は予約済みかそうでないかにかかわらず結果として追加
//     * CarInformationを返す
