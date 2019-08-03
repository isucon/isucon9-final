package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"fmt"
	"net/http"
)

var db *sql.DB

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
	http.HandleFunc("/distance_fare", distance_fare_handler)

	http.ListenAndServe(":8000", nil)
}
