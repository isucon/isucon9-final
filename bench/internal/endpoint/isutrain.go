package endpoint

import "fmt"

const (
	// POST
	Initialize EndpointIdx = iota
	Register
	Login
	Reserve
	ListStations
	SearchTrains
	ListTrainSeats
	ListReservations
)

var isutrainEndpoints = []*Endpoint{
	&Endpoint{path: "/initialize", weight: 1},
	&Endpoint{path: "/register", weight: 1},
	&Endpoint{path: "/login", weight: 1},
	&Endpoint{path: "/reserve", weight: 1},
	&Endpoint{path: "/stations", weight: 1},
	&Endpoint{path: "/train/search", weight: 1},
	&Endpoint{path: "/train/seats", weight: 1},
	&Endpoint{path: "/reservation", weight: 1},
}

const (
	CommitReservation EndpointIdx = iota
	CancelReservation
)

var isutrainDynamicEndpoints = []*Endpoint{
	&Endpoint{path: "/reservation/%s/commit", weight: 1},
	&Endpoint{path: "/reservation/%s/cancel", weight: 1},
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

func IncDynamicPathCounter(idx EndpointIdx) {
	isutrainDynamicEndpoints[idx].inc()
}

func CalcFinalScore() (score int64) {
	for _, endpoint := range isutrainEndpoints {
		score += endpoint.score()
	}
	for _, endpoint := range isutrainDynamicEndpoints {
		score += endpoint.score()
	}
	return
}
