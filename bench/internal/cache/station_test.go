package cache

import (
	"testing"

	"gopkg.in/go-playground/assert.v1"
)

func TestStation(t *testing.T) {
	tests := []struct {
		aOrigin        string
		aDestination   string
		bOrigin        string
		bDestination   string
		wantIsOverwrap bool
		wantErr        error
	}{
		{
			aOrigin:        "東京",
			aDestination:   "大阪",
			bOrigin:        "桜内",
			bDestination:   "舟田",
			wantIsOverwrap: true,
			wantErr:        nil,
		},
		{
			aOrigin:        "初野",
			aDestination:   "気川",
			bOrigin:        "東京",
			bDestination:   "山田",
			wantIsOverwrap: true,
			wantErr:        nil,
		},
		{
			aOrigin:        "初野",
			aDestination:   "気川",
			bOrigin:        "山田",
			bDestination:   "葉千",
			wantIsOverwrap: true,
			wantErr:        nil,
		},
		{
			aOrigin:        "東京",
			aDestination:   "油交",
			bOrigin:        "初野",
			bDestination:   "山田",
			wantIsOverwrap: false,
			wantErr:        nil,
		},
	}

	for _, tt := range tests {
		result, err := isOverwrap(tt.aOrigin, tt.aDestination, tt.bOrigin, tt.bDestination)
		assert.Equal(t, tt.wantErr, err)
		assert.Equal(t, tt.wantIsOverwrap, result)
	}
}
