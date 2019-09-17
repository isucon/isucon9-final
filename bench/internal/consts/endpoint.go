package consts

import "fmt"

const (
	// GET
	ListStationsPath     = "/stations"
	SearchTrainsPath     = "/train/search"
	ListTrainSeatsPath   = "/train/seats"
	ListReservationsPath = "/reservation"

	// POST
	InitializePath = "/initialize"
	RegisterPath   = "/register"
	LoginPath      = "/login"
	ReservePath    = "/reserve"
)

var (
	// POST
	BuildCommitReservationPath = func(id string) string {
		return fmt.Sprintf("/reservation/%s/commit", id)
	}

	// DELETE
	BuildCancelReservationPath = func(id string) string {
		return fmt.Sprintf("/reservation/%s/cancel", id)
	}
)

var (
	// Mock
	MockCommitReservationPath = `=~^/reservation/(\d+)/commit\z`
	MockCancelReservationPath = `=~^/reservation/(\d+)/cancel\z`
)
