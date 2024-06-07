package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime/debug"

	//"runtime/pprof"
	"time"
)

func main() {
	logFile, err := os.OpenFile("application.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	loggingOutputWriter := io.MultiWriter(os.Stdout, logFile)
	handler := slog.NewJSONHandler(loggingOutputWriter, nil)
	logger := slog.New(handler)
	slog.SetDefault(logger)

	//pprofW, _ := os.Create("profile.cpu")
	//pprof.StartCPUProfile(pprofW)
	//defer pprofW.Close()
	defer func() {
		if r := recover(); r != nil {
			slog.Error("Stacktrace from panic: \n" + string(debug.Stack()))
			debug.PrintStack()
		}
	}()
	exit := make(chan string)
	var appcfg = AppConfiguration{}
	appcfg.LoadConfiguration("./mqutil.yml")
	daemon(appcfg.IbmMq)
	for {
		_, ok := <-exit
		if !ok {
			os.Exit(0)
		}
	}
	//time.Sleep(time.Minute * 2)
	//pprof.StopCPUProfile()
	//slog.Info("End of program")
}

// Daemon watches for unintended behaviour, and eventually tries to fix them
func daemon(mq IbmMq) {
	for i := range mq.QueueManagers {
		var qmgr = &mq.QueueManagers[i]
		slog.Info("Registering new ticker", slog.String("qmgr", qmgr.QmName), slog.Duration("interval", qmgr.CheckInterval))
		ticker := time.NewTicker(qmgr.CheckInterval)
		go asyncStatusInf(ticker, mq, *qmgr)
	}
}

func asyncStatusInf(ticker *time.Ticker, mq IbmMq, qmgr QueueManager) {
	<-ticker.C
	var mqStatus = mq.QueueManagers.MustFind(qmgr.QmName).GetStatus()
	if mqStatus != "RUNNING" {
		//try to start mqgr
		mq.QueueManagers.MustFind(qmgr.QmName).RetriedTimes++
		var s = "Attempting restart of qmgr: " + qmgr.QmName + ". Attempt " + fmt.Sprint(mq.QueueManagers.MustFind(qmgr.QmName).RetriedTimes) + " of " + fmt.Sprint(mq.QueueManagers.MustFind(qmgr.QmName).MaxRetryTimes) + "."
		slog.Info(s, slog.String("qmgr", qmgr.QmName))
		_, e := mq.RunCmd("strmqm " + qmgr.QmName)
		if e != nil {
			slog.Warn("Could not start QM: "+qmgr.QmName, slog.String("qmgr", qmgr.QmName))
			// TODO: Notify via email
		} else {
			mq.QueueManagers.MustFind(qmgr.QmName).RetriedTimes = 0
			slog.Info(qmgr.QmName+" succesfully restarted.", slog.String("qmgr", qmgr.QmName))
			// TODO: Notify via email
		}
		if mq.QueueManagers.MustFind(qmgr.QmName).RetriedTimes >= mq.QueueManagers.MustFind(qmgr.QmName).MaxRetryTimes {
			slog.Error("Could not start QMGR after "+fmt.Sprint(qmgr.MaxRetryTimes)+" attempts!", slog.String("qmgr", qmgr.QmName))
			// TODO: Notify that shit hit the fan
			go asyncWatchQmgrStart(ticker, mq, qmgr)
			return
		}
	} else {
		slog.Info("QueueManager is RUNNING", slog.String("qmgr", qmgr.QmName), slog.String("status", "RUNNING"))
	}
	go asyncStatusInf(ticker, mq, qmgr)
}

func asyncWatchQmgrStart(ticker *time.Ticker, mq IbmMq, qmgr QueueManager) {
	<-ticker.C
	slog.Info("Retrying QueueManager status.", slog.String("qmgr", qmgr.QmName))
	var mqStatus = mq.QueueManagers.MustFind(qmgr.QmName).GetStatus()
	if mqStatus == "RUNNING" {
		slog.Info("QueueManager is back online! Starting watcher.", slog.String("qmgr", qmgr.QmName))
		tickerW := time.NewTicker(qmgr.CheckInterval)
		go asyncStatusInf(tickerW, mq, qmgr)
	} else {
		slog.Warn("QueueManager is still offline!", slog.String("qmgr", qmgr.QmName))
		go asyncWatchQmgrStart(ticker, mq, qmgr)
	}
}
