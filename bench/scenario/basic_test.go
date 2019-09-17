package scenario

import (
	"context"
	"log"
	"testing"

	"github.com/chibiegg/isucon9-final/bench/isutrain"
	"github.com/chibiegg/isucon9-final/bench/mock"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestScenario(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	m := mock.Register()
	log.Println(m)

	initClient, err := isutrain.NewClientForInitialize("http://localhost")
	assert.NoError(t, err)
	initClient.ReplaceMockTransport()
	initClient.Initialize(context.Background())

	scenario, err := NewBasicScenario("http://localhost")
	scenario.client.ReplaceMockTransport()
	assert.NoError(t, err)

	assert.NoError(t, scenario.Run(context.TODO()))
}

func TestHTTPStatusCodeError(t *testing.T) {

}
