package consts

// PenaltyWeight は、application error １つにつき課せられるペナルティの重さです
const (
	ApplicationPenaltyWeight = 500

	TrivialPenaltyThreshold = 200
	TrivialPenaltyWeight    = 5000
	TrivialPenaltyPerCount  = 100
)
