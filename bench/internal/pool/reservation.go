package pool

// 中途半端(未コミットかつ未キャンセル)な予約を覚えておく

var IncompleteReservations = newQueue(10000)
