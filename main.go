package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"github.com/Sirupsen/logrus"
	"github.com/wolfeidau/docker-registry/conf"
)

var logger = logrus.New()

func init() {
	logger.Formatter = new(logrus.JSONFormatter)
	logger.Level = logrus.DebugLevel
}

func createPidFile(pidFile string) error {

	if pidString, err := ioutil.ReadFile(pidFile); err == nil {
		pid, err := strconv.Atoi(string(pidString))
		if err == nil {
			if _, err := os.Stat(fmt.Sprintf("/proc/%d/", pid)); err == nil {
				return fmt.Errorf("pid file found, ensure docker-registry is not running or delete %s", pidFile)
			}
		}
	}

	file, err := os.Create(pidFile)
	if err != nil {
		return err
	}

	defer file.Close()

	_, err = fmt.Fprintf(file, "%d", os.Getpid())
	return err
}

func removePidFile(pidFile string) {
	if err := os.Remove(pidFile); err != nil {
		logger.Error("Error removing %s: %s", pidFile, err)
	}
}

func startServer(config *conf.Configuration) {
	logger.Info("using version ", Version)
	logger.Info("starting server on ", config.Listen)
	logger.Info("using dataDir ", config.Data)

	users := NewSingleUserStore(config.Pass)

	auth := NewBasicAuth(users, config.Secret)

	if err := http.ListenAndServe(config.Listen, NewHandler(config.Data, auth)); err != nil {
		logger.Error(err.Error())
	}
}

func main() {

	conf, err := conf.LoadConfiguration()

	if err != nil {
		fmt.Printf("Unable to load env %s", err)
		os.Exit(-1)
	}

	logger.Level = logrus.InfoLevel
	if conf.Debug {
		logger.Level = logrus.DebugLevel
	}
	startServer(conf)
}
