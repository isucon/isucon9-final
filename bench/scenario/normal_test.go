package scenario

import (
	"context"
	"fmt"
	"testing"

	"github.com/chibiegg/isucon9-final/bench/internal/bencherror"
	"github.com/chibiegg/isucon9-final/bench/internal/config"
	"github.com/chibiegg/isucon9-final/bench/internal/endpoint"
	"github.com/chibiegg/isucon9-final/bench/isutrain"
	"github.com/chibiegg/isucon9-final/bench/mock"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestScenario(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	mock.Register()

	initClient, err := isutrain.NewClientForInitialize()
	assert.NoError(t, err)
	initClient.ReplaceMockTransport()
	initClient.Initialize(context.Background())

	config.Debug = true
	assert.NoError(t, NormalScenario(context.Background()))
}

func TestInitializeBenchError(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	m, err := mock.Register()
	assert.NoError(t, err)
	m.Inject(func(path string) error {
		if path == endpoint.GetPath(endpoint.Initialize) {
			return fmt.Errorf("POST /initialize: テスト用のエラーです")
		}
		return nil
	})

	initClient, err := isutrain.NewClientForInitialize()
	assert.NoError(t, err)
	initClient.ReplaceMockTransport()
	initClient.Initialize(context.Background())

	assert.True(t, bencherror.InitializeErrs.IsError())
}

func TestScenarioBenchError(t *testing.T) {

}

func TestHTTPStatusCodeError(t *testing.T) {

}
