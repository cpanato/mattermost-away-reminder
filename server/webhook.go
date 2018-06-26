package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/cpanato/mattermost-away-reminder/model"
)

func Send(webhookUrl string, payload model.MMSlashResponse) {
	marshalContent, _ := json.Marshal(payload)
	var jsonStr = []byte(marshalContent)
	req, err := http.NewRequest("POST", webhookUrl, bytes.NewBuffer(jsonStr))
	req.Header.Set("X-Custom-Header", "mattermostAway")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

}

func PostAways() {
	t, err := time.Parse(LAYOUT, time.Now().Format(LAYOUT))
	if err != nil {
		fmt.Println(err)
	}

	var aways []*model.Away
	if result := <-Srv.Store.Away().GetAwaysForToday(strconv.FormatInt(t.Unix(), 10)); result.Err != nil {
		fmt.Println(result.Err.Error())
		return
	} else if len(result.Data.([]*model.Away)) == 0 {
		fmt.Println("nothing for today")
		return
	} else {
		aways = result.Data.([]*model.Away)
	}

	attachment1 := []model.MMAttachment{}
	for _, away := range aways {
		fmt.Println(away)

		startInt, _ := strconv.ParseInt(away.Start, 10, 64)
		endInt, _ := strconv.ParseInt(away.End, 10, 64)

		msg := fmt.Sprintf("**@%v** will be out from **%v** to **%v**: %v", away.UserName, time.Unix(startInt, 0).Format(LAYOUT), time.Unix(endInt, 0).Format(LAYOUT), away.Reason)

		attach := model.MMAttachment{}
		attachment1 = append(attachment1, *attach.AddField(model.MMField{Title: "Time Away", Value: msg}))
	}
	payload := model.MMSlashResponse{
		Text:        "List of #TimeAway for " + time.Now().Format(LAYOUT),
		Username:    "TimeAway",
		IconUrl:     "https://png.icons8.com/ios/1600/sunbathe.png",
		Attachments: attachment1,
	}
	if Config.MMIncomingWebhook != "" {
		Send(Config.MMIncomingWebhook, payload)
	}

	return
}
