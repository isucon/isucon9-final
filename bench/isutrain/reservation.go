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

type ReservationStatus string

const (
	Pending ReservationStatus = "pending"
	Ok      ReservationStatus = "ok"
)

// Reservation は予約です
type Reservation struct {
	// ID は採番された予約IDです
	ID int `json:"id"`
	// PaymentMethod は決済種別です
	PaymentMethod string `json:"payment_method"`
	// Status は予約の決済状況です
	Status string `json:"status"`
	// ReserveAt は予約時刻です
	ReserveAt time.Time `json:"reserve_at"`
}

type Reservations []*Reservation
