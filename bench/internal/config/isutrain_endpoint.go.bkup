package config

import "fmt"

const (
	// GET
	IsutrainListStationsPath     = "/stations"
	IsutrainSearchTrainsPath     = "/train/search"
	IsutrainListTrainSeatsPath   = "/train/seats"
	IsutrainListReservationsPath = "/reservation"

	// POST
	IsutrainInitializePath = "/initialize"
	IsutrainRegisterPath   = "/register"
	IsutrainLoginPath      = "/login"
	IsutrainReservePath    = "/reserve"
)

var (
	// POST
	IsutrainBuildCommitReservationPath = func(id string) string {
		return fmt.Sprintf("/reservation/%s/commit", id)
	}

	// DELETE
	IsutrainBuildCancelReservationPath = func(id string) string {
		return fmt.Sprintf("/reservation/%s/cancel", id)
	}
)

var (
	// Mock
	IsutrainMockCommitReservationPath = `=~^/reservation/(\d+)/commit\z`
	IsutrainMockCancelReservationPath = `=~^/reservation/(\d+)/cancel\z`
)
