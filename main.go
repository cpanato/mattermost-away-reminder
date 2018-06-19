package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/cpanato/mattermost-away-reminder/server"
)

func main() {
	var flagConfigFile string
	flag.StringVar(&flagConfigFile, "config", "config-away.json", "")
	flag.Parse()

	server.LoadConfig(flagConfigFile)

	server.Start()

	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

}
