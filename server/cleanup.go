package server

import (
	"fmt"
	"strconv"
	"time"
)

func removeOldAways() {
	timeNow, err := time.Parse(LAYOUT, time.Now().Format(LAYOUT))
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Will remove old aways that is old than %v", timeNow)

	if result := <-Srv.Store.Away().DeleteOldAway(strconv.FormatInt(timeNow.Unix(), 10)); result.Err != nil {
		fmt.Println(result.Err.Error())
	}
	fmt.Println("Finish the cleanup of old aways")
}
