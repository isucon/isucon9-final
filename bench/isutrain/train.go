package isutrain

// Train は新幹線列車です
type Train struct {
	// Class は列車種別です
	Class string `json:"class"`
	// Name は列車名です
	Name string `json:"name"`
	// StartStationID は始点駅IDです
	StartStationID int `json:"start_station_id"`
	// EndStation は終点駅IDです
	EndStationID int `json:"end_station_id"`
}

// Trains は列車一覧です
type Trains []*Train

// TODO: SeatClassのconst

// Seat は座席です
type Seat struct {
	// Class は座席種別です
	Class string `json:"class"`
	// Row は席位置の列です(ex. １列)
	Row int `json:"row"`
	// Column は席位置の行です(ex. A行)
	Column string `json:"column"`
	// IsSmokingSeat 喫煙所が近くにあるかどうかのフラグです
	IsSmokingSeat bool `json:"is_smoking_seat"`
}

// Seats は座席一覧です
// TODO: 列でざっくり指定、行でざっくり指定
type Seats []*Seat
