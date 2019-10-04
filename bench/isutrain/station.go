package isutrain

import (
	"errors"

	"github.com/chibiegg/isucon9-final/bench/internal/bencherror"
	"go.uber.org/zap"
)

// ListStationsResponse は /api/stations のレスポンス形式です
type ListStationsResponse []*Station

type Station struct {
	ID                int     `json:"id"`
	Name              string  `json:"name"`
	Distance          float64 `json:"distance"`
	IsStopExpress     bool    `json:"is_stop_express"`
	IsStopSemiExpress bool    `json:"is_stop_semi_express"`
	IsStopLocal       bool    `json:"is_stop_local"`
}

var (
	ErrInvalidStationName = errors.New("駅名が不正です")
)

var stations = []*Station{
	&Station{Name: "東京", IsStopExpress: true, IsStopSemiExpress: true, IsStopLocal: true},
	&Station{Name: "古岡", IsStopExpress: false, IsStopSemiExpress: true, IsStopLocal: true},
	&Station{Name: "絵寒町", IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	&Station{Name: "沙芦公園", IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	&Station{Name: "形顔", IsStopExpress: false, IsStopSemiExpress: true, IsStopLocal: true},
	&Station{Name: "油交", IsStopExpress: true, IsStopSemiExpress: true, IsStopLocal: true},
	&Station{Name: "通墨山", IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	&Station{Name: "初野", IsStopExpress: false, IsStopSemiExpress: true, IsStopLocal: true},
	&Station{Name: "樺威学園", IsStopExpress: false, IsStopSemiExpress: true, IsStopLocal: true},
	&Station{Name: "塩鮫公園", IsStopExpress: false, IsStopSemiExpress: true, IsStopLocal: true},
	&Station{Name: "山田", IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	&Station{Name: "表岡", IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	&Station{Name: "並取", IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	&Station{Name: "細野", IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	&Station{Name: "住郷", IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	&Station{Name: "管英", IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	&Station{Name: "気川", IsStopExpress: false, IsStopSemiExpress: true, IsStopLocal: true},
	&Station{Name: "桐飛", IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	&Station{Name: "樫曲町", IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	&Station{Name: "依酒山", IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	&Station{Name: "堀切町", IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	&Station{Name: "葉千", IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	&Station{Name: "奥山", IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	&Station{Name: "鯉秋寺", IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	&Station{Name: "伍出", IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	&Station{Name: "杏高公園", IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	&Station{Name: "荒川", IsStopExpress: true, IsStopSemiExpress: true, IsStopLocal: true},
	&Station{Name: "磯川", IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	&Station{Name: "茶川", IsStopExpress: false, IsStopSemiExpress: true, IsStopLocal: true},
	&Station{Name: "八実学園", IsStopExpress: false, IsStopSemiExpress: true, IsStopLocal: true},
	&Station{Name: "梓金", IsStopExpress: false, IsStopSemiExpress: true, IsStopLocal: true},
	&Station{Name: "鯉田", IsStopExpress: false, IsStopSemiExpress: true, IsStopLocal: true},
	&Station{Name: "鳴門", IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	&Station{Name: "曲徳町", IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	&Station{Name: "彩岬山", IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	&Station{Name: "根永", IsStopExpress: false, IsStopSemiExpress: true, IsStopLocal: true},
	&Station{Name: "鹿近川", IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	&Station{Name: "結広", IsStopExpress: false, IsStopSemiExpress: true, IsStopLocal: true},
	&Station{Name: "庵金公園", IsStopExpress: false, IsStopSemiExpress: true, IsStopLocal: true},
	&Station{Name: "近岡", IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	&Station{Name: "威香", IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	&Station{Name: "名古屋", IsStopExpress: true, IsStopSemiExpress: true, IsStopLocal: true},
	&Station{Name: "錦太学園", IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	&Station{Name: "和錦台", IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	&Station{Name: "稲冬台", IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	&Station{Name: "松港山", IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	&Station{Name: "甘桜", IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	&Station{Name: "根左海岸", IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	&Station{Name: "島威寺", IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	&Station{Name: "月朱野", IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	&Station{Name: "芋呉川", IsStopExpress: false, IsStopSemiExpress: true, IsStopLocal: true},
	&Station{Name: "木南", IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	&Station{Name: "鳩平ヶ丘", IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	&Station{Name: "維荻学園", IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	&Station{Name: "保池", IsStopExpress: false, IsStopSemiExpress: true, IsStopLocal: true},
	&Station{Name: "九野", IsStopExpress: false, IsStopSemiExpress: true, IsStopLocal: true},
	&Station{Name: "桜田", IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	&Station{Name: "霞苑野", IsStopExpress: false, IsStopSemiExpress: true, IsStopLocal: true},
	&Station{Name: "夷太寺", IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	&Station{Name: "甘野", IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	&Station{Name: "遠山", IsStopExpress: false, IsStopSemiExpress: true, IsStopLocal: true},
	&Station{Name: "銀正", IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	&Station{Name: "末国", IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	&Station{Name: "泉別川", IsStopExpress: false, IsStopSemiExpress: true, IsStopLocal: true},
	&Station{Name: "京都", IsStopExpress: true, IsStopSemiExpress: true, IsStopLocal: true},
	&Station{Name: "桜内", IsStopExpress: false, IsStopSemiExpress: true, IsStopLocal: true},
	&Station{Name: "荻葛ヶ丘", IsStopExpress: false, IsStopSemiExpress: true, IsStopLocal: true},
	&Station{Name: "雨墨", IsStopExpress: false, IsStopSemiExpress: true, IsStopLocal: true},
	&Station{Name: "桂綾寺", IsStopExpress: false, IsStopSemiExpress: true, IsStopLocal: true},
	&Station{Name: "宇治", IsStopExpress: true, IsStopSemiExpress: true, IsStopLocal: true},
	&Station{Name: "塚手海岸", IsStopExpress: false, IsStopSemiExpress: true, IsStopLocal: true},
	&Station{Name: "垣通海岸", IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	&Station{Name: "雨稲ヶ丘", IsStopExpress: false, IsStopSemiExpress: true, IsStopLocal: true},
	&Station{Name: "森果川", IsStopExpress: true, IsStopSemiExpress: true, IsStopLocal: true},
	&Station{Name: "舟田", IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	&Station{Name: "形利", IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	&Station{Name: "午万台", IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	&Station{Name: "早森野", IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	&Station{Name: "桐氷野", IsStopExpress: false, IsStopSemiExpress: true, IsStopLocal: true},
	&Station{Name: "条川", IsStopExpress: true, IsStopSemiExpress: true, IsStopLocal: true},
	&Station{Name: "菊岡", IsStopExpress: false, IsStopSemiExpress: true, IsStopLocal: true},
	&Station{Name: "大阪", IsStopExpress: true, IsStopSemiExpress: true, IsStopLocal: true},
}

func IsValidStations(gotStations []*Station) bool {
	lgr := zap.S()
	if len(stations) != len(gotStations) {
		lgr.Warnf("駅一覧の数が不正: %d", len(gotStations))
		return false
	}
	for i := 0; i < len(stations); i++ {
		var (
			station    = stations[i]
			gotStation = gotStations[i]
		)
		// IDは見ない
		gotStation.ID = station.ID

		// NOTE: 駅一覧では距離を出さないので、距離も見ない

		if *station != *gotStation {
			lgr.Warnf("駅情報が不正: want=%+v, got=%+v", station, gotStation)
			return false
		}
	}
	return true
}

// 下りベースの駅
var sectionMap = map[string]int{
	"東京":   1,
	"古岡":   2,
	"絵寒町":  3,
	"沙芦公園": 4,
	"形顔":   5,
	"油交":   6,
	"通墨山":  7,
	"初野":   8,
	"樺威学園": 9,
	"塩鮫公園": 10,
	"山田":   11,
	"表岡":   12,
	"並取":   13,
	"細野":   14,
	"住郷":   15,
	"管英":   16,
	"気川":   17,
	"桐飛":   18,
	"樫曲町":  19,
	"依酒山":  20,
	"堀切町":  21,
	"葉千":   22,
	"奥山":   23,
	"鯉秋寺":  24,
	"伍出":   25,
	"杏高公園": 26,
	"荒川":   27,
	"磯川":   28,
	"茶川":   29,
	"八実学園": 30,
	"梓金":   31,
	"鯉田":   32,
	"鳴門":   33,
	"曲徳町":  34,
	"彩岬山":  35,
	"根永":   36,
	"鹿近川":  37,
	"結広":   38,
	"庵金公園": 39,
	"近岡":   40,
	"威香":   41,
	"名古屋":  42,
	"錦太学園": 43,
	"和錦台":  44,
	"稲冬台":  45,
	"松港山":  46,
	"甘桜":   47,
	"根左海岸": 48,
	"島威寺":  49,
	"月朱野":  50,
	"芋呉川":  51,
	"木南":   52,
	"鳩平ヶ丘": 53,
	"維荻学園": 54,
	"保池":   55,
	"九野":   56,
	"桜田":   57,
	"霞苑野":  58,
	"夷太寺":  59,
	"甘野":   60,
	"遠山":   61,
	"銀正":   62,
	"末国":   63,
	"泉別川":  64,
	"京都":   65,
	"桜内":   66,
	"荻葛ヶ丘": 67,
	"雨墨":   68,
	"桂綾寺":  69,
	"宇治":   70,
	"塚手海岸": 71,
	"垣通海岸": 72,
	"雨稲ヶ丘": 73,
	"森果川":  74,
	"舟田":   75,
	"形利":   76,
	"午万台":  77,
	"早森野":  78,
	"桐氷野":  79,
	"条川":   80,
	"菊岡":   81,
	"大阪":   82,
}

// aの区間に対し、bの区間が被って入るかチェック (下りベース)
func isKudariOverwrap(aOrigin, aDestination string, bOrigin, bDestination string) (bool, error) {
	var (
		aDepartureNum, ok1 = sectionMap[aOrigin]
		aArrivalNum, ok2   = sectionMap[aDestination]
		bDepartureNum, ok3 = sectionMap[bOrigin]
		bArrivalNum, ok4   = sectionMap[bDestination]
	)
	if !ok1 {
		return false, bencherror.NewSimpleCriticalError("不正な駅 %s が isKudariOverwrapのaOriginに指定されました", aOrigin)
	}
	if !ok2 {
		return false, bencherror.NewSimpleCriticalError("不正な駅 %s が isKudariOverwrapのaDestinationに指定されました", aDestination)
	}
	if !ok3 {
		return false, bencherror.NewSimpleCriticalError("不正な駅 %s が isKudariOverwrapのbOriginに指定されました", bOrigin)
	}
	if !ok4 {
		return false, bencherror.NewSimpleCriticalError("不正な駅 %s が isKudariOverwrapのbDestinationに指定されました", bDestination)
	}

	if bDepartureNum < aDepartureNum && bArrivalNum <= aDepartureNum {
		return false, nil
	}
	if bDepartureNum >= aArrivalNum && bArrivalNum > aArrivalNum {
		return false, nil
	}

	return true, nil
}

// 上り経路か否か
func isKudari(origin, destination string) (bool, error) {
	var (
		originNum, ok1      = sectionMap[origin]
		destinationNum, ok2 = sectionMap[destination]
	)
	if !ok1 {
		return false, bencherror.NewSimpleCriticalError("不正な駅 %s が isKudariのOriginに指定されました", origin)
	}
	if !ok2 {
		return false, bencherror.NewSimpleCriticalError("不正な駅 %s が isKudariのDestinationに指定されました", destination)
	}

	return destinationNum > originNum, nil
}

func IsValidStation(station string) bool {
	if _, ok := sectionMap[station]; !ok {
		return false
	}
	return true
}
