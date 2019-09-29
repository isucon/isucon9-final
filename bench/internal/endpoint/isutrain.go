package endpoint

import (
	"fmt"

	"go.uber.org/zap"
)

const (
	// POST
	Initialize EndpointIdx = iota
	Settings
	Signup
	Login
	Logout
	Reserve
	CommitReservation
	ListStations
	SearchTrains
	ListTrainSeats
	ListReservations
)

var isutrainEndpoints = []*Endpoint{
	&Endpoint{path: "/initialize", weight: 0},
	&Endpoint{path: "/api/settings", weight: 0},
	&Endpoint{path: "/api/auth/signup", weight: 1},
	&Endpoint{path: "/api/auth/login", weight: 1},
	&Endpoint{path: "/api/auth/logout", weight: 1},
	&Endpoint{path: "/api/train/reserve", weight: 1},
	&Endpoint{path: "/api/train/reservation/commit", weight: 1},
	&Endpoint{path: "/api/stations", weight: 1},
	&Endpoint{path: "/api/train/search", weight: 1},
	&Endpoint{path: "/api/train/seats", weight: 1},
	&Endpoint{path: "/api/user/reservations", weight: 1},
}

const (
	CancelReservation EndpointIdx = iota
	ShowReservation
)

var isutrainDynamicEndpoints = []*Endpoint{
	&Endpoint{path: "/api/user/reservations/%d/cancel", weight: 1},
	&Endpoint{path: "/api/user/reservations/%d", weight: 1},
}

func GetPath(idx EndpointIdx) string {
	return isutrainEndpoints[idx].path
}

func GetDynamicPath(idx EndpointIdx, args ...interface{}) string {
	return fmt.Sprintf(isutrainDynamicEndpoints[idx].path, args...)
}

func IncPathCounter(idx EndpointIdx) {
	isutrainEndpoints[idx].inc()
}

func AddExtraScore(idx EndpointIdx, extraScore int64) {
	isutrainEndpoints[idx].addExtraScore(extraScore)
}

func IncDynamicPathCounter(idx EndpointIdx) {
	isutrainDynamicEndpoints[idx].inc()
}

func AddDynamicPathExtraScore(idx EndpointIdx, extraScore int64) {
	isutrainDynamicEndpoints[idx].addExtraScore(extraScore)
}

func CalcFinalScore() (score int64) {
	lgr := zap.S()
	for _, endpoint := range isutrainEndpoints {
		score += endpoint.score()
		lgr.Infof("[%s] score = %d", endpoint.path, endpoint.score())
	}
	for _, endpoint := range isutrainDynamicEndpoints {
		score += endpoint.score()
		lgr.Infof("[%s] score = %d", endpoint.path, endpoint.score())
	}
	return
}

func CalcFinalEndpointCount() (count int64) {
	for _, endpoint := range isutrainEndpoints {
		count += endpoint.count
	}
	for _, endpoint := range isutrainDynamicEndpoints {
		count += endpoint.count
	}
	return
}
