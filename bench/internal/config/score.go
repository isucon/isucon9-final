package config

const (
	ApplicationPenaltyWeight = 500

	// TrivialPenaltyThreshold = 200
	TrivialPenaltyThreshold = 1
	TrivialPenaltyWeight    = 1000
	// TrivialPenaltyWeight    = 5000
	// TrivialPenaltyPerCount  = 100
	TrivialPenaltyPerCount = 3
)

const (
	PremiumSeatExtraScore     = 200
	ReservedSeatExtraScore    = 100
	NonReservedSeatExtraScore = 50
)

func GetFareMultiplier(trainClass, seatClass string) float64 {
	// TODO: fare_masterを元に実装
	return 0
}
