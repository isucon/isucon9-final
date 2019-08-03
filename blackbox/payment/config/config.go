package config

import (
	"io/ioutil"

	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

// Load parses the YAML input s into a Config.
func Load(s string) (*Config, error) {
	cfg := &Config{}
	err := yaml.Unmarshal([]byte(s), cfg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load")
	}
	return cfg, nil
}

// LoadFile parses the given YAML file into a Config.
func LoadFile(filename string) (*Config, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read config file")
	}
	cfg, err := Load(string(content))
	if err != nil {
		return nil, errors.Wrap(err, "failed to loading config file")
	}
	return cfg, nil
}

type Config struct {
	HttpPort string `yaml:"http_port,omitempty"` // HTTP Port
	GrpcPort string `yaml:"grpc_port,omitempty"` // gRPC Port
}
