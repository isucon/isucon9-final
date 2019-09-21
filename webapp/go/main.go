package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	// "sync"
)

var (
	banner = `ISUTRAIN API`
)

var dbx *sqlx.DB

// DB定義

type Station struct {
	ID                int     `json:"id" db:"id"`
	Name              string  `json:"name" db:"name"`
	Distance          float64 `json:"-" db:"distance"`
	IsStopExpress     bool    `json:"is_stop_express" db:"is_stop_express"`
	IsStopSemiExpress bool    `json:"is_stop_semi_express" db:"is_stop_semi_express"`
	IsStopLocal       bool    `json:"is_stop_local" db:"is_stop_local"`
}

type DistanceFare struct {
	Distance float64 `json:"distance" db:"distance"`
	Fare     int     `json:"fare" db:"fare"`
}

type Fare struct {
	TrainClass     string    `json:"train_class" db:"train_class"`
	SeatClass      string    `json:"seat_class" db:"seat_class"`
	StartDate      time.Time `json:"start_date" db:"start_date"`
	FareMultiplier float64   `json:"fare_multiplier" db:"fare_multiplier"`
}

type Train struct {
	Date         time.Time `json:"date" db:"date"`
	DepartureAt  string    `json:"departure_at" db:"departure_at"`
	TrainClass   string    `json:"train_class" db:"train_class"`
	TrainName    string    `json:"train_name" db:"train_name"`
	StartStation string    `json:"start_station" db:"start_station"`
	LastStation  string    `json:"last_station" db:"last_station"`
	IsNobori     bool      `json:"is_nobori" db:"is_nobori"`
}

type Seat struct {
	TrainClass    string `json:"train_class" db:"train_class"`
	CarNumber     int    `json:"car_number" db:"car_number"`
	SeatColumn    string `json:"seat_column" db:"seat_column"`
	SeatRow       int    `json:"seat_row" db:"seat_row"`
	SeatClass     string `json:"seat_class" db:"seat_class"`
	IsSmokingSeat bool   `json:"is_smoking_seat" db:"is_smoking_seat"`
}

type Reservation struct {
	ReservationId int       `json:"reservation_id" db:"reservation_id"`
	UserId        int       `json:"user_id" db:"user_id"`
	Date          time.Time `json:"date" db:"date"`
	TrainClass    string    `json:"train_class" db:"train_class"`
	TrainName     string    `json:"train_name" db:"train_name"`
	Departure     string    `json:"departure" db:"departure"`
	Arrival       string    `json:"arrival" db:"arrival"`
	PaymentStatus string    `json:"payment_method" db:"payment_method"`
	Status        string    `json:"status" db:"status"`
	PaymentId     int       `json:"payment_id" db:"payment_id"`
}

type SeatReservation struct {
	ReservationId int    `json:"reservation_id" db:"reservation_id"`
	CarNumber     int    `json:"car_number" db:"car_number"`
	SeatRow       int    `json:"seat_row" db:"seat_row"`
	SeatColumn    string `json:"seat_column" db:"seat_column"`
}

// 未整理

type CarInformation struct {
	Date                time.Time         `json:"date"`
	TrainClass          string            `json:"train_class"`
	TrainName           string            `json:"train_name"`
	CarNumber           int               `json:"car_number"`
	SeatInformationList []SeatInformation `json:"seats"`
}

type SeatInformation struct {
	Row           int    `json:"row"`
	Column        string `json:"column"`
	Class         string `json:"class"`
	IsSmokingSeat bool   `json:"is_smoking_seat"`
	IsOccupied    bool   `json:"is_occupied"`
}

type TrainSearchResponse struct {
	Class            string            `json:"train_class"`
	Name             string            `json:"train_name"`
	Start            string            `json:"start"`
	Last             string            `json:"last"`
	Departure        int               `json:"departure"`
	Destination      int               `json:"destination"`
	DepartureTime    time.Time         `json:"departure_time"`
	ArrivalTime      time.Time         `json:"arrival_time"`
	SeatAvailability map[string]string `json:"seat_availability"`
	Fare             map[string]int    `json:"seat_fare"`
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, World")
}

func distanceFareHandler(w http.ResponseWriter, r *http.Request) {

	distanceFareList := []DistanceFare{}

	query := "SELECT * FROM distance_fare_master"
	err := dbx.Select(&distanceFareList, query)
	if err != nil {
		panic(err)
	}

	for _, distanceFare := range distanceFareList {
		fmt.Fprintf(w, "%d,%d\n", distanceFare.Distance, distanceFare.Fare)
	}

	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	json.NewEncoder(w).Encode(distanceFareList)
}

func getDistanceFare(origToDestDistance float64) int {

	distanceFareList := []DistanceFare{}

	query := "SELECT distance,fare FROM distance_fare_master ORDER BY distance"
	err := dbx.Select(&distanceFareList, query)
	if err != nil {
		panic(err)
	}

	lastDistance := 0.0
	lastFare := 0
	for _, distanceFare := range distanceFareList {

		fmt.Println(origToDestDistance, distanceFare.Distance, distanceFare.Fare)
		if float64(lastDistance) < origToDestDistance && origToDestDistance < float64(distanceFare.Distance) {
			break
		}
		lastDistance = distanceFare.Distance
		lastFare = distanceFare.Fare
	}

	return lastFare
}

func fareCalc(date time.Time, depStation int, destStation int, trainClass, seatClass string) int {
	//
	// 料金計算メモ
	// 距離運賃(円) * 期間倍率(繁忙期なら2倍等) * 車両クラス倍率(急行・各停等) * 座席クラス倍率(プレミアム・指定席・自由席)
	//

	var err error
	var fromStation, toStation Station

	query := "SELECT * FROM station_master WHERE id=?"

	// From
	err = dbx.Get(&fromStation, query, depStation)
	if err == sql.ErrNoRows {
		panic(err)
	}
	if err != nil {
		panic(err)
	}

	// To
	err = dbx.Get(&fromStation, query, destStation)
	if err == sql.ErrNoRows {
		panic(err)
	}
	if err != nil {
		log.Print(err)
		panic(err)
	}

	fmt.Println("distance", math.Abs(toStation.Distance-fromStation.Distance))
	distFare := getDistanceFare(math.Abs(toStation.Distance - fromStation.Distance))
	fmt.Println("distFare", distFare)

	// 期間・車両・座席クラス倍率
	fareList := []Fare{}
	query = "SELECT * FROM fare_master WHERE train_class=? AND seat_class=? ORDER BY start_date"
	err = dbx.Select(&fareList, query, trainClass, seatClass)
	if err != nil {
		panic(err)
	}

	var selectedFare Fare

	for _, fare := range fareList {
		if err != nil {
			panic(err)
		}

		// TODO: start_dateをちゃんと見る必要がある
		fmt.Println(fare.StartDate, fare.FareMultiplier)
		selectedFare = fare
	}

	fmt.Println("%%%%%%%%%%%%%%%%%%%")

	// TODO: 端数の扱い考える
	// TODO: start_dateをちゃんと見る必要がある
	// TODO: 距離見てる...？
	return int(float64(distFare) * selectedFare.FareMultiplier)
}

func getStationsHandler(w http.ResponseWriter, r *http.Request) {
	/*
		駅一覧
			GET /api/stations

		return []Station{}
	*/

	stations := []Station{}

	query := "SELECT * FROM station_master ORDER BY id"
	err := dbx.Select(&stations, query)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	json.NewEncoder(w).Encode(stations)
}

func (train Train) getAvailableSeats(fromStation Station, toStation Station, seatClass string, isSmokingSeat bool) ([]Seat, error) {
	var err error

	// 全ての座席を
	query := "SELECT * FROM seat_master WHERE train_class=? AND seat_class=? AND is_smoking_seat=?"

	seatList := []Seat{}
	err = dbx.Select(&seatList, query, train.TrainClass, seatClass, isSmokingSeat)
	if err != nil {
		panic(err)
	}

	availableSeatMap := map[string]Seat{}
	for _, seat := range seatList {
		availableSeatMap[fmt.Sprintf("%d_%d_%s", seat.CarNumber, seat.SeatRow, seat.SeatColumn)] = seat
	}

	query = `
	SELECT sr.reservation_id, sr.car_number, sr.seat_row, sr.seat_column
	FROM seat_reservations sr, reservations r, seat_master s, station_master std, station_master sta
	WHERE
		r.reservation_id=sr.reservation_id AND
		s.train_class=r.train_class AND
		s.car_number=sr.car_number AND
		s.seat_column=sr.seat_column AND
		s.seat_row=sr.seat_row AND
		std.name=r.departure AND
		sta.name=r.arrival
	`

	if train.IsNobori {
		query += "AND ((sta.id < ? AND ? <= std.id) OR (sta.id < ? AND ? <= std.id))"
	} else {
		query += "AND ((std.id <= ? AND ? < sta.id) OR (std.id <= ? AND ? < sta.id))"
	}

	seatReservationList := []SeatReservation{}
	err = dbx.Select(&seatReservationList, query, fromStation.ID, fromStation.ID, toStation.ID, toStation.ID)
	if err != nil {
		panic(err)
	}

	for _, seatReservation := range seatReservationList {
		key := fmt.Sprintf("%d_%d_%s", seatReservation.CarNumber, seatReservation.SeatRow, seatReservation.SeatColumn)
		delete(availableSeatMap, key)
	}

	ret := []Seat{}
	for _, seat := range availableSeatMap {
		ret = append(ret, seat)
	}
	return ret, nil
}

func trainSearchHandler(w http.ResponseWriter, r *http.Request) {
	/*
		列車検索
			GET /train/search?use_at=<ISO8601形式の時刻> & from=東京 & to=大阪

		return
			料金
			空席情報
			発駅と着駅の到着時刻

	*/

	jst := time.FixedZone("Asia/Tokyo", 9*60*60)
	date, err := time.Parse(time.RFC3339, r.URL.Query().Get("use_at"))
	if err != nil {
		panic(err)
	}
	date = date.In(jst)

	trainClass := r.URL.Query().Get("train_class")
	from_id, _ := strconv.Atoi(r.URL.Query().Get("from"))
	to_id, _ := strconv.Atoi(r.URL.Query().Get("to"))

	var fromStation, toStation Station
	query := "SELECT * FROM station_master WHERE id=?"

	// From
	err = dbx.Get(&fromStation, query, from_id)
	if err == sql.ErrNoRows {
		panic(err)
	}
	if err != nil {
		panic(err)
	}

	// To
	err = dbx.Get(&toStation, query, to_id)
	if err == sql.ErrNoRows {
		panic(err)
	}
	if err != nil {
		log.Print(err)
		panic(err)
	}

	isNobori := false
	if fromStation.Distance > toStation.Distance {
		isNobori = true
	}

	query = "SELECT * FROM station_master ORDER BY distance"
	if isNobori {
		// 上りだったら駅リストを逆にする
		query += " DESC"
	}

	trainList := []Train{}
	if trainClass == "" {
		query := "SELECT * FROM train_master WHERE date=? AND is_nobori=?"
		err = dbx.Select(&trainList, query, date.Format("2006/01/02"), isNobori)
	} else {
		query := "SELECT * FROM train_master WHERE date=? AND is_nobori=? AND train_class=?"
		err = dbx.Select(&trainList, query, date.Format("2006/01/02"), isNobori, trainClass)
	}
	if err != nil {
		panic(err)
	}

	stations := []Station{}
	err = dbx.Select(&stations, query)
	if err != nil {
		panic(err)
	}

	fmt.Println("From", fromStation)
	fmt.Println("To", toStation)

	trainSearchResponseList := []TrainSearchResponse{}

	for _, train := range trainList {
		isSeekedToFirstStation := false
		isContainsOriginStation := false
		isContainsDestStation := false
		i := 0

		for _, station := range stations {

			if !isSeekedToFirstStation {
				// 駅リストを列車の発駅まで読み飛ばして頭出しをする
				// 列車の発駅以前は止まらないので無視して良い
				if station.Name == train.StartStation {
					isSeekedToFirstStation = true
				} else {
					continue
				}
			}

			if station.ID == from_id {
				// 発駅を経路中に持つ編成の場合フラグを立てる
				isContainsOriginStation = true
			}
			if station.ID == to_id {
				if isContainsOriginStation {
					// 発駅と着駅を経路中に持つ編成の場合
					isContainsDestStation = true
					break
				} else {
					// 出発駅より先に終点が見つかったとき
					fmt.Println("なんかおかしい")
					break
				}
			}
			if station.Name == train.LastStation {
				// 駅が見つからないまま当該編成の終点に着いてしまったとき
				break
			}
			i++
		}

		if isContainsOriginStation && isContainsDestStation {
			// 列車情報

			// TODO: 所要時間計算
			// TODO: ここの値はダミーなのでちゃんと計算して突っ込む
			departureAt := time.Now()
			// TODO: ここの値はダミーなのでちゃんと計算して突っ込む
			arrivalAt := time.Now()



			premium_avail_seats, err := train.getAvailableSeats(fromStation, toStation, "premium", false)
			if err != nil {
				panic(nil)
			}
			premium_smoke_avail_seats, err := train.getAvailableSeats(fromStation, toStation, "premium", true)
			if err != nil {
				panic(nil)
			}

			reserved_avail_seats, err := train.getAvailableSeats(fromStation, toStation, "reserved", false)
			if err != nil {
				panic(nil)
			}
			reserved_smoke_avail_seats, err := train.getAvailableSeats(fromStation, toStation, "reserved", true)
			if err != nil {
				panic(nil)
			}

			premium_avail := "○"
			if len(premium_avail_seats) == 0 {
				premium_avail = "×"
			} else if len(premium_avail_seats) < 10 {
				premium_avail = "△"
			}

			premium_smoke_avail := "○"
			if len(premium_smoke_avail_seats) == 0 {
				premium_smoke_avail = "×"
			} else if len(premium_smoke_avail_seats) < 10 {
				premium_smoke_avail = "△"
			}

			reserved_avail := "○"
			if len(reserved_avail_seats) == 0 {
				reserved_avail = "×"
			} else if len(reserved_avail_seats) < 10 {
				reserved_avail = "△"
			}

			reserved_smoke_avail := "○"
			if len(reserved_smoke_avail_seats) == 0 {
				reserved_smoke_avail = "×"
			} else if len(reserved_smoke_avail_seats) < 10 {
				reserved_smoke_avail = "△"
			}

			// TODO: 空席情報
			seatAvailability := map[string]string{
				"premium":        premium_avail,
				"premium_smoke":  premium_smoke_avail,
				"reserved":       reserved_avail,
				"reserved_smoke": reserved_smoke_avail,
				"non_reserved":   "○",
			}

			// 料金計算
			fareInformation := map[string]int{
				"premium":        fareCalc(date, from_id, to_id, train.TrainClass, "premium"),
				"premium_smoke":  fareCalc(date, from_id, to_id, train.TrainClass, "premium_smoke"),
				"reserved":       fareCalc(date, from_id, to_id, train.TrainClass, "reserved"),
				"reserved_smoke": fareCalc(date, from_id, to_id, train.TrainClass, "reserved_smoke"),
				"non_reserved":   fareCalc(date, from_id, to_id, train.TrainClass, "non_reserved"),
			}

			trainSearchResponseList = append(trainSearchResponseList, TrainSearchResponse{
				train.TrainClass, train.TrainName, train.StartStation, train.LastStation,
				from_id, to_id, departureAt, arrivalAt, seatAvailability, fareInformation,
			})
		}
	}
	resp, err := json.Marshal(trainSearchResponseList)
	if err != nil {
		panic(err)
	}
	w.Write(resp)

}

func trainSeatsHandler(w http.ResponseWriter, r *http.Request) {
	/*
		指定した列車の座席列挙
		GET /train/seats?train_class=のぞみ && train_name=96号
	*/

	jst := time.FixedZone("Asia/Tokyo", 9*60*60)
	date, err := time.Parse(time.RFC3339, r.URL.Query().Get("use_at"))
	if err != nil {
		panic(err)
	}
	date = date.In(jst)

	trainClass := r.URL.Query().Get("train_class")
	trainName := r.URL.Query().Get("train_name")
	carNumber, err := strconv.Atoi(r.URL.Query().Get("car_num"))
	if err != nil {
		panic(err)
	}

	seatList := []Seat{}

	query := "SELECT seat_column,seat_row,seat_class,is_smoking_seat FROM seat_master WHERE train_class=? AND car_number=?"
	err = dbx.Select(&seatList, query, trainClass, carNumber)
	if err != nil {
		panic(err)
	}

	var seatInformationList []SeatInformation
	for _, seat := range seatList {
		var result int

		query := "SELECT COUNT(*) FROM seat_reservations WHERE date=? AND train_class=? AND train_name=? AND car_number=? AND seat_row=? AND seat_column=?"
		err = dbx.Select(
			&result, query,
			date,
			seat.TrainClass,
			trainName,
			seat.CarNumber,
			seat.SeatRow,
			seat.SeatColumn,
		)
		if err != nil {
			panic(err)
		}

		s := SeatInformation{seat.SeatRow, seat.SeatColumn, seat.SeatClass, seat.IsSmokingSeat, false}
		if result == 1 {
			s.IsOccupied = true
		}
		seatInformationList = append(seatInformationList, s)

		// fmt.Fprintf(w, "%d,%d\n", distance, fare)
	}
	c := CarInformation{date, trainClass, trainName, carNumber, seatInformationList}
	resp, err := json.Marshal(c)
	if err != nil {
		panic(err)
	}
	w.Write(resp)
}

func main() {
	// MySQL関連のお膳立て
	var err error

	host := os.Getenv("MYSQL_HOSTNAME")
	if host == "" {
		host = "127.0.0.1"
	}
	port := os.Getenv("MYSQL_PORT")
	if port == "" {
		port = "3306"
	}
	_, err = strconv.Atoi(port)
	if err != nil {
		port = "3306"
	}
	user := os.Getenv("MYSQL_USER")
	if user == "" {
		user = "isutrain"
	}
	dbname := os.Getenv("MYSQL_DATABASE")
	if dbname == "" {
		dbname = "isutrain"
	}
	password := os.Getenv("MYSQL_PASSWORD")
	if password == "" {
		password = "isutrain"
	}

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true&loc=Local",
		user,
		password,
		host,
		port,
		dbname,
	)

	dbx, err = sqlx.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("failed to connect to DB: %s.", err.Error())
	}
	defer dbx.Close()

	// HTTP
	http.HandleFunc("/api/stations", getStationsHandler)
	http.HandleFunc("/api/train/search", trainSearchHandler)
	http.HandleFunc("/api/train/seats", trainSeatsHandler)

	fmt.Println(banner)
	http.ListenAndServe(":8000", nil)
}
