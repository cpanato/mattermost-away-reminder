package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"gopkg.in/robfig/cron.v2"

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

	timeToPost := fmt.Sprintf("@every %vh", server.Config.WebhookNotificationTimeInHours)
	fmt.Printf("Will post the time away every %v hours\n", server.Config.WebhookNotificationTimeInHours)
	c := cron.New()
	c.AddFunc(timeToPost, server.PostAways)
	c.AddFunc("@daily", server.RemoveOldAways)
	go c.Start()
	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt, os.Kill)
	<-sig

}
