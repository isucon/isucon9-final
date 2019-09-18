package assets

import (
	"context"
	"crypto/sha256"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/chibiegg/isucon9-final/bench/isutrain"
	"github.com/chibiegg/isucon9-final/bench/mock"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	assets, err := Load("../assets/testdata")
	assert.NoError(t, err)
	assert.Len(t, assets, 6)

	for _, asset := range assets {
		path := filepath.Join("testdata", asset.Path)
		_, err := os.Stat(path)
		assert.NoError(t, err)
		assert.False(t, os.IsNotExist(err))

		b, err := ioutil.ReadFile(path)
		assert.NoError(t, err)

		hash := sha256.Sum256(b)
		assert.Equal(t, hash, asset.Hash)
	}
}

func TestMock(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	mock.Register()

	client, err := isutrain.NewClient("http://localhost")
	assert.NoError(t, err)
	client.ReplaceMockTransport()

	assets, err := Load("testdata")
	for _, asset := range assets {
		b, err := client.DownloadAsset(context.TODO(), asset.Path)
		assert.NoError(t, err)

		hash := sha256.Sum256(b)
		assert.Equal(t, asset.Hash, hash)
	}
}
