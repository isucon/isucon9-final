package main

import (
	"testing"
)

func TestgetUsableTrainClassList(t *testing.T) {

	fromStation := Station{1, "全部止まる", 10.0, true, true, true}
	ret := getUsableTrainClassList(fromStation, fromStation)

	if len(ret) != 3 {
		t.Fatalf("failed test %#v", ret)
	}

	fromStation = Station{1, "ちょっと止まる", 10.0, false, true, true}
	ret = getUsableTrainClassList(fromStation, fromStation)

	if len(ret) != 2 {
		t.Fatalf("failed test %#v", ret)
	}

	fromStation = Station{1, "各駅", 10.0, false, false, true}
	ret = getUsableTrainClassList(fromStation, fromStation)

	if len(ret) != 1 {
		t.Fatalf("failed test %#v", ret)
	}
}
