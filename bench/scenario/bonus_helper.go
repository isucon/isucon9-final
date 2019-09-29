package scenario

import (
	"errors"

	"github.com/chibiegg/isucon9-final/bench/isutrain"
)

var (
	ErrDuplicateSeat = errors.New("座席検索結果に重複している席が存在します")
)

func CalcNeighborSeatsBonus(seats isutrain.TrainSeats) float64 {
	// 同じ席とか間違えて返ってきてたら減点しなくちゃならない
	var maxLen int

	m := map[int][]string{}
	for _, seat := range seats {
		// FIXME: 重複座席検出
		if _, ok := m[seat.Row]; !ok {
			m[seat.Row] = []string{}
		}
		m[seat.Row] = append(m[seat.Row], seat.Column)

		if len(m[seat.Row]) > maxLen {
			maxLen = len(m[seat.Row])
		}
	}

	switch maxLen {
	case 1:
		return 0
	case 2:
		return 0.2
	case 3:
		return 0.4
	case 4:
		return 0.9
	case 5:
		return 1.0
	default:
		return 0
	}
}
