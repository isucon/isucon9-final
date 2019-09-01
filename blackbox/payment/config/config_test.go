package config

import (
	"flag"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	configFile := flag.String("config-file", "testdata/conf.good.yml", "config file path")
	flag.Parse()
	_, err := LoadFile(*configFile)
	if err != nil {
		t.Errorf("Error parsing %s: %s", "testdata/conf.good.yml", err)
	}
}
