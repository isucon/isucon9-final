package isutrain

import "time"

// TODO: Statusのconst

// PaymentMethod は決済方法です
type PaymentMethod string

const (
	// CreditCard はクレカ決済
	CreditCard PaymentMethod = "creditcard"
	// TicketVendingMachine は券売機決済
	TicketVendingMachine PaymentMethod = "ticket-vending-machine"
)

// ReservationStatus は予約状態です
type ReservationStatus string

const (
	// Pending は予約が確定していない状態です
	Pending ReservationStatus = "pending"
	// Ok は予約が確定した状態です
	Ok ReservationStatus = "ok"
)

// SeatReservation は座席予約です
type SeatReservation struct {
	// ID は採番された予約IDです
	ID int `json:"id"`
	// PaymentMethod は決済種別です
	PaymentMethod string `json:"payment_method"`
	// Status は予約の決済状況です
	Status string `json:"status"`
	// ReserveAt は予約時刻です
	ReserveAt time.Time `json:"reserve_at"`
}

// SeatReservations は座席一覧です
type SeatReservations []*SeatReservation
