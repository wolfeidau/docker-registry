package conf

import "github.com/kelseyhightower/envconfig"

type Configuration struct {
	Listen  string
	Data    string
	PidFile string
	Debug   bool
}

func LoadConfiguration(fileName string) (*Configuration, error) {

	var conf Configuration
	err := envconfig.Process("myapp", &conf)

	if err != nil {
		return nil, err
	}

	if conf.Listen == "" {
		conf.Listen = ":5000"
	}

	if conf.Data == "" {
		conf.Data = "/Users/markw/docker/docker_index"
	}

	return &conf, nil
}
