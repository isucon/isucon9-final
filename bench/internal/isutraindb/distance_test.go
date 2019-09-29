package isutraindb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetDistanceFare(t *testing.T) {
	for station1 := range distanceMap {
		for station2 := range distanceMap {
			if station1 == station2 {
				continue
			}

			distanceFare, err := GetDistanceFare(station1, station2)
			assert.NoError(t, err)
			assert.Condition(t, func() bool {
				return distanceFare > 0
			})
		}
	}
}

func TestGetDistance(t *testing.T) {
	for station1 := range distanceMap {
		for station2 := range distanceMap {
			if station1 == station2 {
				continue
			}
			distance, err := getDistance(station1, station2)
			assert.NoError(t, err)
			assert.NotZero(t, distance)
		}
	}
}
