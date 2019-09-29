package isutraindb

import (
	"errors"
	"math"
)

var (
	ErrStationNotFound = errors.New("指定された駅は見つかりませんでした")
	ErrInvalidDistance = errors.New("距離が不正です")
)

var distanceMap = map[string]float64{
	"東京":   0.000000,
	"古岡":   12.745608,
	"絵寒町":  32.107649,
	"沙芦公園": 45.037138,
	"形顔":   52.773422,
	"油交":   60.930427,
	"通墨山":  72.915666,
	"初野":   80.517696,
	"樺威学園": 96.053004,
	"塩鮫公園": 112.665386,
	"山田":   119.444708,
	"表岡":   131.462232,
	"並取":   149.826976,
	"細野":   166.909255,
	"住郷":   182.323457,
	"管英":   188.887999,
	"気川":   207.599747,
	"桐飛":   217.900353,
	"樫曲町":  229.697609,
	"依酒山":  244.770170,
	"堀切町":  251.948590,
	"葉千":   269.009280,
	"奥山":   275.384825,
	"鯉秋寺":  284.952294,
	"伍出":   291.499545,
	"杏高公園": 310.086023,
	"荒川":   325.553902,
	"磯川":   334.561908,
	"茶川":   343.842013,
	"八実学園": 355.192588,
	"梓金":   374.584703,
	"鯉田":   381.847874,
	"鳴門":   393.244289,
	"曲徳町":  411.802367,
	"彩岬山":  420.375925,
	"根永":   428.829478,
	"鹿近川":  445.676144,
	"結広":   457.246917,
	"庵金公園": 474.044387,
	"近岡":   487.270404,
	"威香":   504.163580,
	"名古屋":  519.612391,
	"錦太学園": 531.408202,
	"和錦台":  548.584849,
	"稲冬台":  554.215596,
	"松港山":  572.885503,
	"甘桜":   584.344724,
	"根左海岸": 603.713433,
	"島威寺":  614.711098,
	"月朱野":  633.406177,
	"芋呉川":  640.097895,
	"木南":   657.573946,
	"鳩平ヶ丘": 677.211495,
	"維荻学園": 689.581633,
	"保池":   696.405431,
	"九野":   711.087956,
	"桜田":   728.268005,
	"霞苑野":  735.983348,
	"夷太寺":  744.581560,
	"甘野":   751.340202,
	"遠山":   770.125141,
	"銀正":   788.163214,
	"末国":   799.939778,
	"泉別川":  807.476895,
	"京都":   819.772794,
	"桜内":   833.349255,
	"荻葛ヶ丘": 839.298450,
	"雨墨":   853.080719,
	"桂綾寺":  863.842723,
	"宇治":   869.266132,
	"塚手海岸": 878.247393,
	"垣通海岸": 893.724394,
	"雨稲ヶ丘": 900.098745,
	"森果川":  909.518544,
	"舟田":   919.249073,
	"形利":   938.540025,
	"午万台":  954.151248,
	"早森野":  966.498192,
	"桐氷野":  975.568259,
	"条川":   990.339004,
	"菊岡":   1005.597665,
	"大阪":   1024.983484,
}

// GetDistance は２駅間の距離を取得します
func getDistance(from, to string) (float64, error) {

	fromPos, ok := distanceMap[from]
	if !ok {
		return -1, ErrStationNotFound
	}
	toPos, ok := distanceMap[to]
	if !ok {
		return -1, ErrStationNotFound
	}

	return math.Abs(toPos - fromPos), nil
}

// GetDistanceFare は距離運賃を取得します
func GetDistanceFare(from, to string) (int, error) {

	distance, err := getDistance(from, to)
	if err != nil {
		return -1, err
	}

	switch {
	case distance > 0 && distance < 50:
		return 2500, nil
	case distance > 50 && distance < 75:
		return 3000, nil
	case distance > 75 && distance < 100:
		return 3700, nil
	case distance > 100 && distance < 150:
		return 4500, nil
	case distance > 150 && distance < 200:
		return 5200, nil
	case distance > 200 && distance < 300:
		return 6000, nil
	case distance > 300 && distance < 400:
		return 7200, nil
	case distance > 400 && distance < 500:
		return 8300, nil
	case distance > 500 && distance < 1000:
		return 12000, nil
	case distance > 1000:
		return 20000, nil
	default:
		return -1, ErrInvalidDistance
	}
}

// StopInfo は駅に停車するフラグ情報です
type StopInfo struct {
	IsStopExpress     bool
	IsStopSemiExpress bool
	IsStopLocal       bool
}

var stopInfoMap = map[string]*StopInfo{
	"東京":   &StopInfo{IsStopExpress: true, IsStopSemiExpress: true, IsStopLocal: true},
	"古岡":   &StopInfo{IsStopExpress: false, IsStopSemiExpress: true, IsStopLocal: true},
	"絵寒町":  &StopInfo{IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	"沙芦公園": &StopInfo{IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	"形顔":   &StopInfo{IsStopExpress: false, IsStopSemiExpress: true, IsStopLocal: true},
	"油交":   &StopInfo{IsStopExpress: true, IsStopSemiExpress: true, IsStopLocal: true},
	"通墨山":  &StopInfo{IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	"初野":   &StopInfo{IsStopExpress: false, IsStopSemiExpress: true, IsStopLocal: true},
	"樺威学園": &StopInfo{IsStopExpress: false, IsStopSemiExpress: true, IsStopLocal: true},
	"塩鮫公園": &StopInfo{IsStopExpress: false, IsStopSemiExpress: true, IsStopLocal: true},
	"山田":   &StopInfo{IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	"表岡":   &StopInfo{IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	"並取":   &StopInfo{IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	"細野":   &StopInfo{IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	"住郷":   &StopInfo{IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	"管英":   &StopInfo{IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	"気川":   &StopInfo{IsStopExpress: false, IsStopSemiExpress: true, IsStopLocal: true},
	"桐飛":   &StopInfo{IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	"樫曲町":  &StopInfo{IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	"依酒山":  &StopInfo{IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	"堀切町":  &StopInfo{IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	"葉千":   &StopInfo{IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	"奥山":   &StopInfo{IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	"鯉秋寺":  &StopInfo{IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	"伍出":   &StopInfo{IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	"杏高公園": &StopInfo{IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	"荒川":   &StopInfo{IsStopExpress: true, IsStopSemiExpress: true, IsStopLocal: true},
	"磯川":   &StopInfo{IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	"茶川":   &StopInfo{IsStopExpress: false, IsStopSemiExpress: true, IsStopLocal: true},
	"八実学園": &StopInfo{IsStopExpress: false, IsStopSemiExpress: true, IsStopLocal: true},
	"梓金":   &StopInfo{IsStopExpress: false, IsStopSemiExpress: true, IsStopLocal: true},
	"鯉田":   &StopInfo{IsStopExpress: false, IsStopSemiExpress: true, IsStopLocal: true},
	"鳴門":   &StopInfo{IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	"曲徳町":  &StopInfo{IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	"彩岬山":  &StopInfo{IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	"根永":   &StopInfo{IsStopExpress: false, IsStopSemiExpress: true, IsStopLocal: true},
	"鹿近川":  &StopInfo{IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	"結広":   &StopInfo{IsStopExpress: false, IsStopSemiExpress: true, IsStopLocal: true},
	"庵金公園": &StopInfo{IsStopExpress: false, IsStopSemiExpress: true, IsStopLocal: true},
	"近岡":   &StopInfo{IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	"威香":   &StopInfo{IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	"名古屋":  &StopInfo{IsStopExpress: true, IsStopSemiExpress: true, IsStopLocal: true},
	"錦太学園": &StopInfo{IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	"和錦台":  &StopInfo{IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	"稲冬台":  &StopInfo{IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	"松港山":  &StopInfo{IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	"甘桜":   &StopInfo{IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	"根左海岸": &StopInfo{IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	"島威寺":  &StopInfo{IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	"月朱野":  &StopInfo{IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	"芋呉川":  &StopInfo{IsStopExpress: false, IsStopSemiExpress: true, IsStopLocal: true},
	"木南":   &StopInfo{IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	"鳩平ヶ丘": &StopInfo{IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	"維荻学園": &StopInfo{IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	"保池":   &StopInfo{IsStopExpress: false, IsStopSemiExpress: true, IsStopLocal: true},
	"九野":   &StopInfo{IsStopExpress: false, IsStopSemiExpress: true, IsStopLocal: true},
	"桜田":   &StopInfo{IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	"霞苑野":  &StopInfo{IsStopExpress: false, IsStopSemiExpress: true, IsStopLocal: true},
	"夷太寺":  &StopInfo{IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	"甘野":   &StopInfo{IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	"遠山":   &StopInfo{IsStopExpress: false, IsStopSemiExpress: true, IsStopLocal: true},
	"銀正":   &StopInfo{IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	"末国":   &StopInfo{IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	"泉別川":  &StopInfo{IsStopExpress: false, IsStopSemiExpress: true, IsStopLocal: true},
	"京都":   &StopInfo{IsStopExpress: true, IsStopSemiExpress: true, IsStopLocal: true},
	"桜内":   &StopInfo{IsStopExpress: false, IsStopSemiExpress: true, IsStopLocal: true},
	"荻葛ヶ丘": &StopInfo{IsStopExpress: false, IsStopSemiExpress: true, IsStopLocal: true},
	"雨墨":   &StopInfo{IsStopExpress: false, IsStopSemiExpress: true, IsStopLocal: true},
	"桂綾寺":  &StopInfo{IsStopExpress: false, IsStopSemiExpress: true, IsStopLocal: true},
	"宇治":   &StopInfo{IsStopExpress: true, IsStopSemiExpress: true, IsStopLocal: true},
	"塚手海岸": &StopInfo{IsStopExpress: false, IsStopSemiExpress: true, IsStopLocal: true},
	"垣通海岸": &StopInfo{IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	"雨稲ヶ丘": &StopInfo{IsStopExpress: false, IsStopSemiExpress: true, IsStopLocal: true},
	"森果川":  &StopInfo{IsStopExpress: true, IsStopSemiExpress: true, IsStopLocal: true},
	"舟田":   &StopInfo{IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	"形利":   &StopInfo{IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	"午万台":  &StopInfo{IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	"早森野":  &StopInfo{IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: true},
	"桐氷野":  &StopInfo{IsStopExpress: false, IsStopSemiExpress: true, IsStopLocal: true},
	"条川":   &StopInfo{IsStopExpress: true, IsStopSemiExpress: true, IsStopLocal: true},
	"菊岡":   &StopInfo{IsStopExpress: false, IsStopSemiExpress: true, IsStopLocal: true},
	"大阪":   &StopInfo{IsStopExpress: true, IsStopSemiExpress: true, IsStopLocal: true},
}

// GetStopInfo は、駅に停車するフラグ情報を返します
func GetStopInfo(station string) (isStopExpress, isStopSemiExpress, isStopLocal bool, err error) {
	stopInfo, ok := stopInfoMap[station]
	if !ok {
		return false, false, false, ErrStationNotFound
	}

	return stopInfo.IsStopExpress, stopInfo.IsStopSemiExpress, stopInfo.IsStopLocal, nil
}
