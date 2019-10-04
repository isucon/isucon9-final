package config

import "time"

const (
	InitializeTimeout = 20 * time.Second
	APITimeout        = 5 * time.Second
	BenchmarkTimeout  = 60 * time.Second
)

const (
	WorkloadMultiplier = 1
)

const (
	AttackSearchTrainTimeout    = 20 * time.Second
	AttackListTrainSeatsTimeout = 20 * time.Second
)

var Debug bool
var SlackWebhookURL string
var Language = "unknown"
