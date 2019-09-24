package xrandom

// FIXME: 予約リクエスト生成
// FIXME:

const (
	StopExpress int = iota << 1
	StopSemiExpress
	StopLocal
)

// station
var (
	// 数直線上の距離がそのままコスト(大きいほど負荷が高くなる)
	stations = []string{
		"東京",
		"古岡",
		"絵寒町",
		"沙芦公園",
		"形顔",
		"油交",
		"通墨山",
		"初野",
		"樺威学園",
		"塩鮫公園",
		"山田",
		"表岡",
		"並取",
		"細野",
		"住郷",
		"管英",
		"気川",
		"桐飛",
		"樫曲町",
		"依酒山",
		"堀切町",
		"葉千",
		"奥山",
		"鯉秋寺",
		"伍出",
		"杏高公園",
		"荒川",
		"磯川",
		"茶川",
		"八実学園",
		"梓金",
		"鯉田",
		"鳴門",
		"曲徳町",
		"彩岬山",
		"根永",
		"鹿近川",
		"結広",
		"庵金公園",
		"近岡",
		"威香",
		"名古屋",
		"錦太学園",
		"和錦台",
		"稲冬台",
		"松港山",
		"甘桜",
		"根左海岸",
		"島威寺",
		"月朱野",
		"芋呉川",
		"木南",
		"鳩平ヶ丘",
		"維荻学園",
		"保池",
		"九野",
		"桜田",
		"霞苑野",
		"夷太寺",
		"甘野",
		"遠山",
		"銀正",
		"末国",
		"泉別川",
		"京都",
		"桜内",
		"荻葛ヶ丘",
		"雨墨",
		"桂綾寺",
		"宇治",
		"塚手海岸",
		"垣通海岸",
		"雨稲ヶ丘",
		"森果川",
		"舟田",
		"形利",
		"午万台",
		"早森野",
		"桐氷野",
		"条川",
		"菊岡",
		"大阪",
	}
)

// train

var (
	trainClasses = []string{
		"遅いやつ",
		"中間",
		"最速",
	}
)

type User struct {
	Email    string
	Password string
}

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
		return ""
	}
}
