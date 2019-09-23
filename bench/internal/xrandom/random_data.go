package xrandom

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/chibiegg/isucon9-final/bench/internal/util"
)

// FIXME: 予約リクエスト生成
// FIXME:

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

func init() {
	// ユーザのメールアドレスやパスワードはチューニングポイントでないので、起動時にシャッフルして使う
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(users), func(i, j int) {
		users[i], users[j] = users[j], users[i]
	})
}

type User struct {
	Email    string
	Password string
}

func GenRandomUser() (*User, error) {
	emailRandomStr, err := util.SecureRandomStr(10)
	if err != nil {
		return nil, err
	}
	passwdRandomStr, err := util.SecureRandomStr(20)
	if err != nil {
		return nil, err
	}
	return &User{
		Email:    fmt.Sprintf("%s@example.com", emailRandomStr),
		Password: passwdRandomStr,
	}, nil
}

var (
	users = []*User{
		&User{Email: "sugiyamaakira@example.com", Password: "7@FN+GXf_8"},
		&User{Email: "mituru51@example.com", Password: "$wkGGAlRj9"},
		&User{Email: "qyoshida@example.com", Password: "@93DZHSp*n"},
		&User{Email: "naoki98@example.com", Password: "ou+3HWke^3"},
		&User{Email: "gmiyake@example.com", Password: "CKW3UVwh$t"},
		&User{Email: "aoyamatakuma@example.com", Password: "@iF+mTFH82"},
		&User{Email: "ukijima@example.com", Password: "_3u&JEm1uI"},
		&User{Email: "akirakijima@example.com", Password: "Vu9R%DbQa("},
		&User{Email: "tkimura@example.com", Password: "V%0OePdO_C"},
		&User{Email: "miki42@example.com", Password: "(OW0LlXrj4"},
		&User{Email: "harukakato@example.com", Password: "xfIidsQW_6"},
		&User{Email: "chiyoaoyama@example.com", Password: "v5giauKF#3"},
		&User{Email: "akemimurayama@example.com", Password: "F&U8DESns@"},
		&User{Email: "minorumurayama@example.com", Password: "!U59SXKg^$"},
		&User{Email: "sasadayuta@example.com", Password: "%B8_XNiE(U"},
		&User{Email: "osamuwakamatsu@example.com", Password: "4MtlEzWn$A"},
		&User{Email: "akira93@example.com", Password: "4HkBySp__v"},
		&User{Email: "yoichisasada@example.com", Password: "iIh+Bddv_4"},
		&User{Email: "tsubasasuzuki@example.com", Password: "5Weng+my$Q"},
		&User{Email: "kijimachiyo@example.com", Password: "@jZ3BbAy$U"},
		&User{Email: "shirokawa@example.com", Password: "%fF$KXgrr4"},
		&User{Email: "takahashitsubasa@example.com", Password: "^J0Z1SeF25"},
		&User{Email: "hirokikato@example.com", Password: "(!+7cThbFF"},
		&User{Email: "yoichinishinosono@example.com", Password: "(iueJ^cn30"},
		&User{Email: "tarouno@example.com", Password: "*h58hPtvW&"},
		&User{Email: "yamagishinaotohideki@example.com", Password: "1ht%Zl^u+&"},
		&User{Email: "kumiko46@example.com", Password: "1gk_SYzm&D"},
		&User{Email: "yasuhirokimura@example.com", Password: "Hqs6e+J*X$"},
		&User{Email: "yumiko32@example.com", Password: "TrGE9GKc^("},
		&User{Email: "enakatsugawa@example.com", Password: "*422BNlM76"},
		&User{Email: "hamadayuta@example.com", Password: "*pKyki$Dw1"},
		&User{Email: "uuno@example.com", Password: "w^j$9#YjNr"},
		&User{Email: "yumiko74@example.com", Password: "L4sw29Vu%V"},
		&User{Email: "tharada@example.com", Password: "eDaey6Sh!_"},
		&User{Email: "tsubasakondo@example.com", Password: "98$PG9Ot!X"},
		&User{Email: "btanaka@example.com", Password: "_f41zZQrSX"},
		&User{Email: "tomomiito@example.com", Password: "!6RUWQq8_0"},
		&User{Email: "tsubasa18@example.com", Password: "guvt%1Ack%"},
		&User{Email: "itsuchiya@example.com", Password: "%5ZjdjdesI"},
		&User{Email: "kyosuke85@example.com", Password: "s_4H%RaB&T"},
		&User{Email: "sogaki@example.com", Password: "tWlozVL6_3"},
		&User{Email: "rikatsuchiya@example.com", Password: "KN!oB9FlR$"},
		&User{Email: "akemi00@example.com", Password: "ou_ZkQ9Y*4"},
		&User{Email: "momokoishida@example.com", Password: "#@l(GyQeo3"},
		&User{Email: "nishinosonotomoya@example.com", Password: "6^7yNx!gt&"},
		&User{Email: "maaya87@example.com", Password: "&gPTL^Wr6F"},
		&User{Email: "akemiyoshimoto@example.com", Password: "n5@%7BfJ$5"},
		&User{Email: "kanonaotohideki@example.com", Password: "%4#JRis71&"},
		&User{Email: "saitoakemi@example.com", Password: "!7+4vuDl^4"},
		&User{Email: "yuta29@example.com", Password: "sR+1z@Uxo)"},
		&User{Email: "rika62@example.com", Password: "06j7Fwq6+Z"},
		&User{Email: "manabuito@example.com", Password: "Nyl+GdvK_2"},
		&User{Email: "asuka57@example.com", Password: "o+7Q8jCjG^"},
		&User{Email: "takuma86@example.com", Password: "s_@6OgZs4Y"},
		&User{Email: "matsumotokyosuke@example.com", Password: "L_5LKy3qll"},
		&User{Email: "naotohidekikato@example.com", Password: "_uTbYv14q9"},
		&User{Email: "yoshimotorei@example.com", Password: "ccd6(MVn+l"},
		&User{Email: "yasuhiro92@example.com", Password: "(z5RCx$dz!"},
		&User{Email: "qmurayama@example.com", Password: "8_1OLzwt+6"},
		&User{Email: "wyamagishi@example.com", Password: "4a36sJ$6I@"},
		&User{Email: "kanorei@example.com", Password: "l*u19qZh*j"},
		&User{Email: "akiraekoda@example.com", Password: "!i2JOz4arm"},
		&User{Email: "ttsuda@example.com", Password: "XtlZ@5Uiq^"},
		&User{Email: "minoru39@example.com", Password: "D(4P2w!I&r"},
		&User{Email: "ogakinaoko@example.com", Password: "O!h4CQ)iJc"},
		&User{Email: "tsudasayuri@example.com", Password: "!3R$1YdHSK"},
		&User{Email: "aotayosuke@example.com", Password: "@*0LjyL5%X"},
		&User{Email: "taichinagisa@example.com", Password: "Qy%l6X9qP%"},
		&User{Email: "jmurayama@example.com", Password: "V8lKpJsC*6"},
		&User{Email: "tomoyatakahashi@example.com", Password: "zfX6B7wA*s"},
		&User{Email: "ryohei69@example.com", Password: "!GUxP6UUy1"},
		&User{Email: "rsasada@example.com", Password: "(eHIyGon@1"},
		&User{Email: "kyosukenakamura@example.com", Password: "31MD^6fC@m"},
		&User{Email: "nakajimarika@example.com", Password: "O$Ks_Qrg!9"},
		&User{Email: "ntsuda@example.com", Password: "2$G^M1PCYh"},
		&User{Email: "kyosukekudo@example.com", Password: "#3Or$E^+)$"},
		&User{Email: "hanakokano@example.com", Password: "eG!m6AXlSr"},
		&User{Email: "kumikokoizumi@example.com", Password: "K9mUtXeR@_"},
		&User{Email: "kondoryohei@example.com", Password: "qg*32uIp*9"},
		&User{Email: "chiyo32@example.com", Password: "@c7CrDkV7J"},
		&User{Email: "sotarosato@example.com", Password: "&!vhXtBL46"},
		&User{Email: "kimurayumiko@example.com", Password: "x$w*ZEDdd6"},
		&User{Email: "hyamaguchi@example.com", Password: "^j15P7Xdd9"},
		&User{Email: "maimiyake@example.com", Password: "!!4Kop^4QR"},
		&User{Email: "naokitsuchiya@example.com", Password: "@!FFgu&$f4"},
		&User{Email: "kijimataichi@example.com", Password: "x8PB^^^h+@"},
		&User{Email: "taichi80@example.com", Password: "%l%G9Kfua0"},
		&User{Email: "aotamiki@example.com", Password: "_R81TxQvtG"},
		&User{Email: "takuma75@example.com", Password: "#22aTNOaVN"},
		&User{Email: "tomomi40@example.com", Password: ")N@8j3Ac)9"},
		&User{Email: "sakamototsubasa@example.com", Password: "_t8u6RkftI"},
		&User{Email: "uharada@example.com", Password: "_4TfB*tmUJ"},
		&User{Email: "yoichi18@example.com", Password: "YZ24)f1bv!"},
		&User{Email: "iyamagishi@example.com", Password: "&F*HXGmd4K"},
		&User{Email: "yumikoyoshida@example.com", Password: "M3S^@n5k#b"},
		&User{Email: "kumiko89@example.com", Password: "z(93C1Zn&F"},
		&User{Email: "kaorisato@example.com", Password: "JLe&2Tyt9Z"},
		&User{Email: "kenichiyoshida@example.com", Password: "(M2IrHjVx^"},
		&User{Email: "hhirokawa@example.com", Password: "E8^10HWt^_"},
		&User{Email: "nanakasuzuki@example.com", Password: "UMYY#Hp1(2"},
	}
)
