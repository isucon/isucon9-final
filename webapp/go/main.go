package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

var db *sql.DB

type CarInformation struct {
	Date time.Time
	TrainClass string
	TrainName string
	CarNumber int
	SeatList []TrainSeat
}

type TrainSeat struct {
	Row int
	Column string
	Class string
	IsSmokingSeat bool
	IsOccupied bool
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, World")
}

func distance_fare_handler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT * FROM distance_fare_master")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var distance int
	var fare int
	for rows.Next() {
		err := rows.Scan(&distance, &fare)
		if err != nil {
			panic(err)
		}

		fmt.Fprintf(w, "%d,%d\n", distance, fare)
	}

	err = rows.Err()
	if err != nil {
		panic(err)
	}
}



func train_search_handler(w http.ResponseWriter, r *http.Request) {
	/*
		列車検索
			GET /train/search?use_at=<ISO8601形式の時刻> & from=東京 & to=大阪
	*/
}

func train_seats_handler(w http.ResponseWriter, r *http.Request) {
	/*
		指定した列車の座席列挙
		GET /train/seats?train_class=のぞみ && train_name=96号
	*/
	date, err := time.Parse(time.RFC3339, r.URL.Query().Get("date"))
	if err != nil {
		panic(err)
	}
	train_class := r.URL.Query().Get("train_class")
	train_name := r.URL.Query().Get("train_name")
	car_number, err := strconv.Atoi(r.URL.Query().Get("car_num"))
	if err != nil {
		panic(err)
	}

	rows, err := db.Query("SELECT seat_column,seat_row,seat_class,is_smoking_seat FROM seat_master WHERE train_class=? AND car_number=?", 
		train_class, car_number)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var seat_row int
	var seat_column, seat_class string
	var is_smoking_seat bool
	var seats []TrainSeat
	for rows.Next() {
		err := rows.Scan(&seat_column, &seat_row, &seat_class, &is_smoking_seat)
		if err != nil {
			panic(err)
		}
		var result int
		db.QueryRow("SELECT COUNT(*) FROM seat_reservations WHERE date=? AND train_class=? AND train_name=? AND car_number=? AND seat_row=? AND seat_column=?",
			date,
			train_class, 
			train_name, 
			car_number, 
			seat_row, 
			seat_column).Scan(&result)
		s := TrainSeat{seat_row, seat_column, seat_class, is_smoking_seat, false}
		if result == 1 {
			s.IsOccupied = true
		}
		seats = append(seats, s)
		
		// fmt.Fprintf(w, "%d,%d\n", distance, fare)
	}
	c := CarInformation{date, train_class, train_name, car_number, seats}
	resp, err := json.Marshal(c)
	if err != nil {
		panic(err)
}
	w.Write(resp)


	err = rows.Err()
	if err != nil {
		panic(err)
	}
}

func main() {
	// MySQL関連のお膳立て
	var err error
	db, err = sql.Open("mysql", "isucon:isucon@tcp(127.0.0.1:3306)/isutrain")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// HTTP
	http.HandleFunc("/", handler)
	http.HandleFunc("/train/search", train_search_handler)
	http.HandleFunc("/train/seats", train_seats_handler)

	http.ListenAndServe(":8000", nil)
}
