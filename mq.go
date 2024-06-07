package main

import (
	"errors"
	"log/slog"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

//Contains various mq functions

// Runs a command on desired Ibm Mq
// pipeIn - stdin to be piped into desired command
func (mq *IbmMq) RunCmd(cmd string, pipeIn ...string) (string, error) {
	var baseCmd = ""
	switch mq.Connector.ConnectorType {
	case "docker":
		baseCmd += "docker exec -i ibmmq-nase " + cmd
	case "local":
		baseCmd += cmd
	case "ssh":
		return "NotImpl", errors.ErrUnsupported
	default:
		panic("Unsupported connector type! Unsupported connector type '" + mq.Connector.ConnectorType + "' in IbmMq.Connector.ConnectorType! Check your yaml configuration (ibmmq.connections.connectorType)")
	}
	uid, e := uuid.NewRandom()
	id := uid.String()
	if e != nil {
		id = time.UTC.String()
		slog.Error("Unable to generate UUID!", slog.String("err", e.Error()))
	}
	slog.Debug("Running command", slog.String("cmd", baseCmd), slog.String("pipeIn", strings.Join(pipeIn, " ")), slog.String("cmdid", id))
	command := exec.Command("/bin/bash", "-c", baseCmd)
	if pipeIn != nil {
		command.Stdin = strings.NewReader(strings.Join(pipeIn, "\n"))
	}
	out, err := command.Output()
	if err != nil {
		slog.Error("Command completed with errors!", slog.String("cmd", baseCmd), slog.String("pipeIn", strings.Join(pipeIn, " ")), slog.String("cmdid", id), slog.String("err", err.Error()))
	} else {
		slog.Debug("Command completed successfully.", slog.String("cmdid", id))
	}
	return string(out), err
}

func (qmgr QueueManagers) Exists(name string) bool {
	for i := range qmgr {
		if qmgr[i].QmName == name {
			return true
		}
	}
	return false
}

// Finds QueueManager in QueueManagers by name
func (qmgr QueueManagers) Find(name string) (*QueueManager, error) {
	for i := range qmgr {
		if qmgr[i].QmName == name {
			return &qmgr[i], nil
		}
	}
	return nil, errors.New("Unable to find QueueManager with name: " + name)
}

// Finds QueueManager in QueueManagers by name. Does not return errors
func (qmgr QueueManagers) MustFind(name string) *QueueManager {
	for i := range qmgr {
		if qmgr[i].QmName == name {
			return &qmgr[i]
		}
	}
	return nil
}

// Runs an mqsc command on selected QueueManager
func (mq *IbmMq) RunMqscCmd(qmgrName string, mqcmd string) (string, error) {
	if !mq.QueueManagers.Exists(qmgrName) {
		return "", errors.New("Unable to find QueueManager '" + qmgrName + "' in IbmMq object! Is it declared in yml configuration file?")
	}
	return mq.RunCmd("runmqsc "+qmgrName, mqcmd)
}

// Returns status of the queue manager
func (qmgr *QueueManager) GetStatus() string {
	out, err := qmgr.IbmMq.RunMqscCmd(qmgr.QmName, "DIS QMSTATUS")
	if err == nil {
		statusRegex := regexp.MustCompile(`STATUS\(([A-Z]*)\)`)
		var s = statusRegex.FindStringSubmatch(out)
		return s[1]
	}
	slog.Error("Unable to get status", slog.String("qmgr", qmgr.QmName))
	return "ERROR"
}

// Stops the queue manager
func (qmgr *QueueManager) Stop(name string) error {
	if !qmgr.IbmMq.QueueManagers.Exists(qmgr.QmName) {
		return errors.New("QueueManager " + qmgr.QmName + " does not exist!")
	}
	_, e := qmgr.IbmMq.RunCmd("endmqm " + qmgr.QmName)
	return e
}

// Starts the queue manager
func (qmgr *QueueManager) Start(name string) error {
	if !qmgr.IbmMq.QueueManagers.Exists(qmgr.QmName) {
		return errors.New("QueueManager " + qmgr.QmName + " does not exist!")
	}
	_, e := qmgr.IbmMq.RunCmd("strmqm " + qmgr.QmName)
	return e
}
