package main

import (
	"fmt"
)

func getUsableTrainClassList(fromStation Station, toStation Station) []string {
	usable := map[string]string{}

	for key, value := range TrainClassMap {
		usable[key] = value
	}

	if !fromStation.IsStopExpress {
		delete(usable, "express")
	}
	if !fromStation.IsStopSemiExpress {
		delete(usable, "semi_express")
	}
	if !fromStation.IsStopLocal {
		delete(usable, "local")
	}

	if !toStation.IsStopExpress {
		delete(usable, "express")
	}
	if !toStation.IsStopSemiExpress {
		delete(usable, "semi_express")
	}
	if !toStation.IsStopLocal {
		delete(usable, "local")
	}

	ret := []string{}
	for _, v := range usable {
		ret = append(ret, v)
	}

	return ret
}

func (train Train) getAvailableSeats(fromStation Station, toStation Station, seatClass string, isSmokingSeat bool) ([]Seat, error) {
	// 指定種別の空き座席を返す
	var err error

	// 全ての座席を取得する
	query := "SELECT * FROM seat_master WHERE train_class=? AND seat_class=? AND is_smoking_seat=?"

	seatList := []Seat{}
	err = dbx.Select(&seatList, query, train.TrainClass, seatClass, isSmokingSeat)
	if err != nil {
		return nil, err
	}

	// 指定種別の予約を検索
	reservationList := []Reservation{}
	query = "SELECT * FROM reservations WHERE train_class=?"
	err = dbx.Select(&reservationList, query, train.TrainClass)
	fmt.Println(reservationList, train.TrainClass)
	if err != nil {
		return nil, err
	}

// func checkSeatClass(trainClass string, carNumber, seatRow int, seatColumn, seatClass string, isSmokingSeat bool) (bool, error) {

	// 予約したい区間と関係のない予約を除外する
	var applicableReservationList []Reservation
	for _, reservation := range reservationList {
		departureStationID, err := getStationID(reservation.Departure)
		if err != nil {
			return nil, err
		}
		arrivalStationID, err := getStationID(reservation.Arrival)
		if err != nil {
			return nil, err
		}

		fmt.Println(reservation, seatClass, isSmokingSeat, train.IsNobori)
		if train.IsNobori {
			if (arrivalStationID < fromStation.ID && fromStation.ID <= departureStationID) || (arrivalStationID < toStation.ID && toStation.ID <= departureStationID) || (fromStation.ID < arrivalStationID && departureStationID < toStation.ID) {
				ok, err := checkSeatClass(reservation, seatClass, isSmokingSeat)
				if err != nil {
					return nil, err
				}
				if ok {
					applicableReservationList = append(applicableReservationList, reservation)
				}
			}
		} else {
			if (departureStationID <= fromStation.ID && fromStation.ID < arrivalStationID) || (departureStationID <= toStation.ID && toStation.ID < arrivalStationID) || (arrivalStationID < fromStation.ID && toStation.ID < departureStationID) {
				ok, err := checkSeatClass(reservation, seatClass, isSmokingSeat)
				if err != nil {
					return nil, err
				}
				if ok {
					applicableReservationList = append(applicableReservationList, reservation)
				}
			}
		}
	}

	// シート種別





	availableSeatMap := map[string]Seat{}
	for _, seat := range seatList {
		availableSeatMap[fmt.Sprintf("%d_%d_%s", seat.CarNumber, seat.SeatRow, seat.SeatColumn)] = seat
	}

	// すでに取られている予約を取得する

	// query = `
	// SELECT sr.reservation_id, sr.car_number, sr.seat_row, sr.seat_column
	// FROM 
	// seat_reservations sr, 
	// reservations r, 
	// seat_master s, 
	// station_master std, 
	// station_master sta
	// WHERE
		// r.reservation_id=sr.reservation_id AND
		// s.train_class=r.train_class AND
		// s.car_number=sr.car_number AND
		// s.seat_column=sr.seat_column AND
		// s.seat_row=sr.seat_row AND
		// std.name=r.departure AND
		// sta.name=r.arrival
	// `
	// if train.IsNobori {
	// 		query += 
	// 		"AND ((sta.id < fromStation.ID AND fromStation.ID <= std.id) OR 
	// 		(sta.id < toStation.ID AND toStation.ID <= std.id) OR 
	// 		(fromStation.ID < sta.id AND std.id < toStation.ID))"
	// } else {
	// 		query += "AND ((std.id <= fromStation.ID AND fromStation.ID < sta.id) OR 
	// 		(std.id <= toStation.ID AND toStation.ID < sta.id) OR 
	// 		(sta.id < fromStation.ID AND toStation.ID < std.id))"
	// }
	
	if len(applicableReservationList) > 0 {
		query = "SELECT * FROM seat_reservations"
		for i, v := range applicableReservationList {
			if i == 0 {
				query += " WHERE "
			} else {
				query += " OR "
			}
			query += fmt.Sprintf("reservation_id=%d", v.ReservationId)
		}
		fmt.Println(query)
		fmt.Printf("%#v\n", applicableReservationList)
		seatReservationList := []SeatReservation{}
		err = dbx.Select(&seatReservationList, query)
		if err != nil {
			panic(err)
			return nil, err
		}

		
		for _, seatReservation := range seatReservationList {
			key := fmt.Sprintf("%d_%d_%s", seatReservation.CarNumber, seatReservation.SeatRow, seatReservation.SeatColumn)
			delete(availableSeatMap, key)
		}
	}

	// ok := []SeatReservation{}
	// for _, v := seatReservationList {
	// 	if train.IsNobori {

	// 	}
	// }




	ret := []Seat{}
	for _, seat := range availableSeatMap {
		ret = append(ret, seat)
	}
	return ret, nil
}

func checkSeatClass(reservation Reservation, seatClass string, isSmokingSeat bool) (bool, error) {
	var result struct {
		SeatClass string `db:"seat_class"`
		isSmokingSeat bool `db:"is_smoking_seat"`
	}

	err := dbx.Get(&result, "SELECT seat_class,is_smoking_seat FROM seat_master JOIN seat_reservations ON seat_master.car_number=seat_reservations.car_number AND seat_master.seat_column=seat_reservations.seat_column AND seat_master.seat_row=seat_reservations.seat_row WHERE reservation_id=? AND train_class=?", reservation.ReservationId, reservation.TrainClass)
	if err != nil {
		panic(err)
		return false, err
	}

	return (result.SeatClass == seatClass && result.isSmokingSeat == isSmokingSeat), nil

}

func getStationID(name string) (int, error) {
	var result int
	err := dbx.Get(&result, "SELECT id FROM station_master WHERE name=?", name)
	if err != nil {
		return 0, err
	}
	return result, nil
}
