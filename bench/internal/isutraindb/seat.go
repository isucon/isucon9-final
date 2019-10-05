package isutraindb

import "github.com/chibiegg/isucon9-final/bench/internal/bencherror"

// GetSeatClass は、列車クラスと車両番号から座席クラスを解決します
func GetSeatClass(trainClass string, carNum int) string {
	switch {
	case trainClass == "中間" && carNum == 1:
		return "non-reserved"
	case trainClass == "中間" && carNum == 2:
		return "non-reserved"
	case trainClass == "中間" && carNum == 3:
		return "non-reserved"
	case trainClass == "中間" && carNum == 4:
		return "non-reserved"
	case trainClass == "中間" && carNum == 5:
		return "non-reserved"
	case trainClass == "中間" && carNum == 6:
		return "reserved"
	case trainClass == "中間" && carNum == 7:
		return "reserved"
	case trainClass == "中間" && carNum == 8:
		return "premium"
	case trainClass == "中間" && carNum == 9:
		return "premium"
	case trainClass == "中間" && carNum == 10:
		return "premium"
	case trainClass == "中間" && carNum == 11:
		return "reserved"
	case trainClass == "中間" && carNum == 12:
		return "reserved"
	case trainClass == "中間" && carNum == 13:
		return "reserved"
	case trainClass == "中間" && carNum == 14:
		return "reserved"
	case trainClass == "中間" && carNum == 15:
		return "reserved"
	case trainClass == "中間" && carNum == 16:
		return "reserved"
	case trainClass == "最速" && carNum == 1:
		return "non-reserved"
	case trainClass == "最速" && carNum == 2:
		return "non-reserved"
	case trainClass == "最速" && carNum == 3:
		return "non-reserved"
	case trainClass == "最速" && carNum == 4:
		return "reserved"
	case trainClass == "最速" && carNum == 5:
		return "reserved"
	case trainClass == "最速" && carNum == 6:
		return "reserved"
	case trainClass == "最速" && carNum == 7:
		return "reserved"
	case trainClass == "最速" && carNum == 8:
		return "premium"
	case trainClass == "最速" && carNum == 9:
		return "premium"
	case trainClass == "最速" && carNum == 10:
		return "premium"
	case trainClass == "最速" && carNum == 11:
		return "reserved"
	case trainClass == "最速" && carNum == 12:
		return "reserved"
	case trainClass == "最速" && carNum == 13:
		return "reserved"
	case trainClass == "最速" && carNum == 14:
		return "reserved"
	case trainClass == "最速" && carNum == 15:
		return "reserved"
	case trainClass == "最速" && carNum == 16:
		return "reserved"
	case trainClass == "遅いやつ" && carNum == 1:
		return "non-reserved"
	case trainClass == "遅いやつ" && carNum == 2:
		return "non-reserved"
	case trainClass == "遅いやつ" && carNum == 3:
		return "non-reserved"
	case trainClass == "遅いやつ" && carNum == 4:
		return "non-reserved"
	case trainClass == "遅いやつ" && carNum == 5:
		return "non-reserved"
	case trainClass == "遅いやつ" && carNum == 6:
		return "non-reserved"
	case trainClass == "遅いやつ" && carNum == 7:
		return "non-reserved"
	case trainClass == "遅いやつ" && carNum == 8:
		return "premium"
	case trainClass == "遅いやつ" && carNum == 9:
		return "premium"
	case trainClass == "遅いやつ" && carNum == 10:
		return "premium"
	case trainClass == "遅いやつ" && carNum == 11:
		return "non-reserved"
	case trainClass == "遅いやつ" && carNum == 12:
		return "non-reserved"
	case trainClass == "遅いやつ" && carNum == 13:
		return "non-reserved"
	case trainClass == "遅いやつ" && carNum == 14:
		return "non-reserved"
	case trainClass == "遅いやつ" && carNum == 15:
		return "non-reserved"
	case trainClass == "遅いやつ" && carNum == 16:
		return "reserved"
	default:
		bencherror.SystemErrs.AddError(bencherror.NewSimpleCriticalError("不正なtrainClass=%s, carNum=%d が指定されました", trainClass, carNum))
		return ""
	}
}
