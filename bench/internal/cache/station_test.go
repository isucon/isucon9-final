package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKudariOverwrap(t *testing.T) {
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
		{
			aOrigin:        "古岡",
			aDestination:   "荒川",
			bOrigin:        "荒川",
			bDestination:   "鳴門",
			wantIsOverwrap: false,
			wantErr:        nil,
		},
		{
			aOrigin:        "古岡",
			aDestination:   "荒川",
			bOrigin:        "古岡",
			bDestination:   "荒川",
			wantIsOverwrap: true,
			wantErr:        nil,
		},
		{
			aOrigin:        "古岡",
			aDestination:   "荒川",
			bOrigin:        "山田",
			bDestination:   "鳴門",
			wantIsOverwrap: true,
			wantErr:        nil,
		},
	}

	for _, tt := range tests {
		result, err := isKudariOverwrap(tt.aOrigin, tt.aDestination, tt.bOrigin, tt.bDestination)
		assert.Equal(t, tt.wantErr, err)
		assert.Equal(t, tt.wantIsOverwrap, result)
	}
}

func TestNoboriOverwrap(t *testing.T) {

}

func TestKudari(t *testing.T) {
	tests := []struct {
		Departure    string
		Arrival      string
		wantIsKudari bool
		wantErr      error
	}{
		{
			Departure:    "東京",
			Arrival:      "大阪",
			wantIsKudari: true,
			wantErr:      nil,
		},
		{
			Departure:    "大阪",
			Arrival:      "東京",
			wantIsKudari: false,
			wantErr:      nil,
		},
		{
			Departure:    "初野",
			Arrival:      "山田",
			wantIsKudari: true,
			wantErr:      nil,
		},
		{
			Departure:    "山田",
			Arrival:      "初野",
			wantIsKudari: false,
			wantErr:      nil,
		},
	}

	for _, tt := range tests {
		kudari, err := isKudari(tt.Departure, tt.Arrival)
		assert.Equal(t, tt.wantErr, err)
		assert.Equal(t, tt.wantIsKudari, kudari)
	}
}
