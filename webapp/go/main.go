package main

import (
	"database/sql"
	"fmt"
	"net/http"
)

var db sql.DB

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, World")
}

func fares_handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, World")
	rows, err := db.Query("SELECT * FROM fares_master")
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
	db, err := sql.Open("mysql", "isucon:isucon@tcp(127.0.0.1:3306)/isutrain")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// HTTP
	http.HandleFunc("/", handler)
	http.HandleFunc("/fares", fares_handler)

	http.ListenAndServe(":8000", nil)
}
