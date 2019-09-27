package isutraindb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetSeatClass(t *testing.T) {
	trainClasses := []string{"最速", "中間", "遅いやつ"}

	seatClassCounter := map[string]int{
		"non-reserved": 0,
		"reserved":     0,
		"premium":      0,
	}
	for carNum := 1; carNum < 17; carNum++ {
		for _, trainClass := range trainClasses {
			seatClass := GetSeatClass(trainClass, carNum)
			assert.Condition(t, func() bool {
				return len(seatClass) > 0
			})
			seatClassCounter[seatClass]++
		}
	}

	assert.Equal(t, 20, seatClassCounter["non-reserved"])
	assert.Equal(t, 19, seatClassCounter["reserved"])
	assert.Equal(t, 9, seatClassCounter["premium"])
}
