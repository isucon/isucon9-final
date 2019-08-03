package isutrain

import (
	"context"
	"log"
	"sync"
	"testing"

	"github.com/jarcoal/httpmock"
)

func TestBasicScenario(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	RegisterMockEndpoints()

	scenario := NewBasicScenario("http://localhost")
	go scenario.Run(context.Background())

	var (
		sum int
		wg  sync.WaitGroup
	)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for benchResult := range scenario.BenchResultStream() {
			sum += benchResult.Type.Score()
		}
	}()

	wg.Wait()
	log.Println(sum)
}
