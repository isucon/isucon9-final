package pool

import "github.com/chibiegg/isucon9-final/bench/isutrain"

var loggedInQueue = newQueue(10000)

func PutLoggedIn(sess *isutrain.Session) {
	loggedInQueue.Enqueue(sess)
}

func getLoggedIn() (sess *isutrain.Session, err error) {
	sess = loggedInQueue.Dequeue()
	if sess == nil {
		sess, err = isutrain.NewSession()
	}
	return
}

func GetListStations() (*isutrain.Session, error) {
	return getLoggedIn()
}

func GetSearchTrains() (*isutrain.Session, error) {
	return getLoggedIn()
}

func GetListTrainSeats() (*isutrain.Session, error) {
	return getLoggedIn()
}
