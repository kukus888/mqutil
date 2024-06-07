package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type AppConfiguration struct {
	IbmMq IbmMq `yaml:"ibmmq"`
}

type IbmMq struct {
	QueueManagers QueueManagers `yaml:"queueManagers"`
	Connector     MqConnector   `yaml:"connections"`
}

type QueueManagers []QueueManager

type MqConnector struct {
	ConnectorType string `yaml:"connectorType"`
	ContainerName string `yaml:"containerName"`
}

type QueueManager struct {
	IbmMq         *IbmMq
	QmName        string        `yaml:"qmName"`
	CheckInterval time.Duration `yaml:"checkInterval"`
	Status        string
	RetriedTimes  int
	MaxRetryTimes int `yaml:"retryTimes"`
}

// Loads app configuration from a file
func (cfg *AppConfiguration) LoadConfiguration(cfgPath string) {
	var defaultMaxRetryTimes = 5
	var defaultCheckInterval = time.Minute * 2

	if _, err := os.Stat(cfgPath); errors.Is(err, os.ErrNotExist) {
		panic("Configuration file " + cfgPath + " does not exist!")
	}
	f, err := os.ReadFile(cfgPath)
	if err != nil {
		panic(err.Error())
	}
	err = yaml.Unmarshal(f, &cfg)
	if err != nil {
		panic("Failed to unmarshal configuration file! " + err.Error())
	}
	for qmgr := range cfg.IbmMq.QueueManagers {
		var qm = &cfg.IbmMq.QueueManagers[qmgr]
		if qm.MaxRetryTimes == 0 {
			slog.Warn("Warning! QueueManager '" + qm.QmName + "' does not have set retryTimes! (default=" + fmt.Sprint(defaultMaxRetryTimes) + ")")
			qm.MaxRetryTimes = defaultMaxRetryTimes
		}
		if qm.CheckInterval == 0 {
			slog.Warn("Warning! QueueManager '" + qm.QmName + "' does not have set checkInterval! (default=" + defaultCheckInterval.String() + ")")
			qm.CheckInterval = defaultCheckInterval
		}
		cfg.IbmMq.QueueManagers[qmgr].IbmMq = &cfg.IbmMq
	}
}
