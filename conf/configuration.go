package conf

import (
	"io/ioutil"
	"log"

	"github.com/BurntSushi/toml"
)

type Configuration struct {
	Listen  string `toml:"listen"`
	Data    string `toml:"data"`
	PidFile string `toml:"pid-file"`
	Debug   bool   `toml:"debug"`
}

func LoadConfiguration(fileName string) *Configuration {
	config, err := parseTomlConfiguration(fileName)
	if err != nil {
		log.Println("Couldn't parse configuration file: " + fileName)
		panic(err)
	}
	return config
}

func parseTomlConfiguration(filename string) (*Configuration, error) {
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	conf := &Configuration{}
	_, err = toml.Decode(string(body), conf)
	if err != nil {
		return nil, err
	}

	if conf.Listen == "" {
		conf.Listen = ":5000"
	}

	if conf.Data == "" {
		conf.Data = "/Users/markw/docker/docker_index"
	}

	return conf, nil
}
