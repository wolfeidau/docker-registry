package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

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
	logger.Info("using pidFile", config.PidFile)

	if config.PidFile != "" {

		if err := createPidFile(config.PidFile); err != nil {
			logger.Error(err)
			os.Exit(1)
		}

		defer removePidFile(config.PidFile)

		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, os.Kill, os.Signal(syscall.SIGTERM))
		go func() {
			sig := <-c
			logger.Debug("Received signal '%v', exiting\n", sig)
			removePidFile(config.PidFile)
			os.Exit(0)
		}()
	}

	if err := http.ListenAndServe(config.Listen, NewHandler(config.Data)); err != nil {
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
