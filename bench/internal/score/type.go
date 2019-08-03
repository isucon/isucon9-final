package score

import "fmt"

type Type int

const (
	_ Type = iota

	// GET
	SearchTrainsType
	ListTrainSeatsType
	ListReservationsType

	// POST
	RegisterType
	LoginType
	ReserveType
	CommitReservationType

	// DELETE
	CancelReservationType
)

func (t Type) String() string {
	switch t {
	case SearchTrainsType:
		return "SearchTrains"
	case ListTrainSeatsType:
		return "ListTrainSeats"
	case ListReservationsType:
		return "ListReservations"
	case RegisterType:
		return "Register"
	case LoginType:
		return "Login"
	case ReserveType:
		return "Reserve"
	case CommitReservationType:
		return "CommitReservation"
	case CancelReservationType:
		return "CancelReservation"
	default:
		return fmt.Sprintf("Unknown %d", t)
	}
}

func (t Type) Score() int {
	switch t {
	case SearchTrainsType:
		return 1
	case ListTrainSeatsType:
		return 1
	case ListReservationsType:
		return 1
	case RegisterType:
		return 1
	case LoginType:
		return 1
	case ReserveType:
		return 1
	case CommitReservationType:
		return 1
	case CancelReservationType:
		return 1
	default:
		return 0
	}
}
