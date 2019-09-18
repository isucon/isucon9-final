package scenario

import (
	"context"
	"errors"
	"testing"

	"github.com/chibiegg/isucon9-final/bench/internal/bencherror"
	"github.com/chibiegg/isucon9-final/bench/internal/config"
	"github.com/chibiegg/isucon9-final/bench/isutrain"
	"github.com/chibiegg/isucon9-final/bench/mock"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestScenario(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	mock.Register()

	initClient, err := isutrain.NewClientForInitialize("http://localhost")
	assert.NoError(t, err)
	initClient.ReplaceMockTransport()
	initClient.Initialize(context.Background())

	scenario, err := NewBasicScenario("http://localhost")
	scenario.Client.ReplaceMockTransport()
	assert.NoError(t, err)

	assert.NoError(t, scenario.Run(context.TODO()))
}

func TestInitializeBenchError(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	m := mock.Register()
	m.Inject(func(path string) error {
		if path == config.InitializePath {
			return errors.New("")
		}
		return nil
	})

	initClient, err := isutrain.NewClientForInitialize("http://localhost")
	assert.NoError(t, err)
	initClient.ReplaceMockTransport()
	initClient.Initialize(context.Background())

	assert.True(t, bencherror.InitializeErrs.IsError())
}

func TestScenarioBenchError(t *testing.T) {

}

func TestHTTPStatusCodeError(t *testing.T) {

}
