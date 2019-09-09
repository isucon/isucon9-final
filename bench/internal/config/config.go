package config

import "time"

const (
	Concurrency = 10

	BenchmarkerLevelInterval = 5 * time.Second

	InitializeTimeout = 20 * time.Second
	APITimeout        = 10 * time.Second
	BenchmarkTimeout  = 180 * time.Second
)
