package main

import (
	"flag"
	"os"
	"os/signal"
	"touch_fish_missile/src/client"
	"touch_fish_missile/src/config"
	"touch_fish_missile/src/datebase"
	"touch_fish_missile/src/log"
	"touch_fish_missile/src/server"
)

var serverMode bool
var token string
var group string
var logLevel string

func cmd() {
	flag.BoolVar(&serverMode, "s", false, "start with server mode or client mode; true clientMode false serverMode")
	flag.StringVar(&token, "token", "unknown", "client token")
	flag.StringVar(&logLevel, "logLevel", "error", "log level: panic fatal error warn info debug trace")
	flag.StringVar(&group, "group", "default", "group tag")
	flag.Parse()
}

func main() {
	cmd()
	conf := config.GetInstance()
	conf.LogLevel = logLevel
	log.Init()
	if conf.SaveChatRecord {
		db := datebase.InitDB()
		defer db.Close()
	}
	if serverMode {
		log.Info("Now Start With Server Mode")
		server.StartServer()
	} else {
		log.Info("Now Start With Client Mode")
		client.StartClient(token, group)
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	log.Info("Interrupt")
}
