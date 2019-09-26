package util

import (
	"context"
	"time"
)

// FormatISO8601 はISO8601形式で時刻をフォーマットします
func FormatISO8601(t time.Time) string {
	return t.Format("2006-01-02T15:04:05+09:00")
}

// ParseISO8601 はISO8601形式で時刻をパースします
func ParseISO8601(t string) (time.Time, error) {
	return time.Parse("2006-01-02T15:04:05+09:00", t)
}

// MeasureScenario はシナリオの経過時間を計測します
func MeasureScenario(ctx context.Context, scenarioFn func(ctx context.Context) error) (time.Duration, error) {
	t := time.Now()
	err := scenarioFn(ctx)
	return time.Since(t), err
}
