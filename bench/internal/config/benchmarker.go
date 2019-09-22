package config

import "time"

const (
	Concurrency = 10

	InitializeTimeout = 20 * time.Second
	APITimeout        = 1 * time.Second
	// BenchmarkTimeout  = 180 * time.Second
	BenchmarkTimeout = 30 * time.Second
)

var Debug bool
