package cache

import (
	"errors"
)

var (
	ErrInvalidStationName = errors.New("駅名が不正です")
)

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
	if !ok1 || !ok2 || !ok3 || !ok4 {
		return false, ErrInvalidStationName
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
	if !ok1 || !ok2 {
		return false, ErrInvalidStationName
	}

	return destinationNum > originNum, nil
}
