package conf

import "github.com/kelseyhightower/envconfig"

type Configuration struct {
	Listen, Data, Namespace, Redis, Secret, Pass string
	Debug                                        bool
}

func LoadConfiguration() (*Configuration, error) {

	var conf Configuration
	err := envconfig.Process("registry", &conf)

	if err != nil {
		return nil, err
	}

	if conf.Listen == "" {
		conf.Listen = ":5000"
	}

	if conf.Redis == "" {
		conf.Redis = "127.0.0.1:6379"
	}

	if conf.Secret == "" {
		conf.Secret = "TodlelOfBooHybUmtOifOul6"
	}

	if conf.Namespace == "" {
		conf.Namespace = "library"
	}

	if conf.Data == "" {
		conf.Data = "/Users/markw/docker/docker_index"
	}

	if conf.Pass == "" {
		conf.Pass = "test1234asdfg"
	}

	return &conf, nil
}
