package scenario

import (
	"context"
	"sync"
	"testing"

	"github.com/chibiegg/isucon9-final/bench/internal/score"
	"github.com/chibiegg/isucon9-final/bench/isutrain"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestBasicScenario(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	isutrain.RegisterMockEndpoints()

	scenario := NewBasicScenario("http://localhost")
	go scenario.Run(context.Background())

	var (
		scores = []int{
			score.RegisterType.Score(),
			score.LoginType.Score(),
			score.SearchTrainsType.Score(),
			score.ListTrainSeatsType.Score(),
			score.ReserveType.Score(),
			score.CommitReservationType.Score(),
			score.ListReservationsType.Score(),
		}
		got, want int
		wg        sync.WaitGroup
	)
	wg.Add(1)
	go func() {
		defer wg.Done()

		var idx int
		for requestResult := range scenario.RequestResultStream() {
			got += requestResult.Type.Score()
			want += scores[idx]

			idx++
		}
	}()

	wg.Wait()
	assert.Equal(t, want, got)
}
