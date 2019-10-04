package config

// 参考値
// TrivialPenaltyThreshold = 200
// TrivialPenaltyWeight    = 5000
// TrivialPenaltyPerCount  = 100

const (
	// ApplicationPenaltyWeight はwebappのエラー１つにつき課せられるペナルティの重み
	ApplicationPenaltyWeight = 5

	// NOTE: timeoutや golang netパッケージのTemporaryエラー(netエラーとする) に関する設定

	// TrivialPenaltyThreshold は netエラーにペナルティが課せられる閾値
	TrivialPenaltyThreshold = 1
	// TrivialPenaltyWeight は netエラーに課せられるペナルティの重み
	TrivialPenaltyWeight = 5
	// TrivialPenaltyPerCount は netエラー幾つにつきペナルティが課せられるかの除数
	TrivialPenaltyPerCount = 1
)

const (
	// ReservedSeatExtraScore は 指定席の予約成功に加えられるボーナス
	ReservedSeatExtraScore = 10
)
