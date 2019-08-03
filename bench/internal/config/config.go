package config

import "time"

const (
	Concurrency = 10

	MinErrCountThreshold = 20
	MaxErrCountThreshold = 50

	BenchmarkerLevelInterval = 5 * time.Second

	BenchmarkTimeout = 60 * time.Second

	InitializeTimeout  = 30 * time.Second
	IsutrainAPITimeout = 15 * time.Second
)
