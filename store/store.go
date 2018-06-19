package store

import (
	"time"

	l4g "github.com/alecthomas/log4go"
	"github.com/cpanato/mattermost-away-reminder/model"
)

type StoreResult struct {
	Data interface{}
	Err  *model.AppError
}

type StoreChannel chan StoreResult

func Must(sc StoreChannel) interface{} {
	r := <-sc
	if r.Err != nil {
		l4g.Close()
		time.Sleep(time.Second)
		panic(r.Err)
	}

	return r.Data
}

type Store interface {
	Away() AwayStore
	Close()
	DropAllTables()
}

type AwayStore interface {
	Save(away *model.Away) StoreChannel
	ListAll() StoreChannel
	ListUserAway(userName string) StoreChannel
	GetAwaysById(id string) StoreChannel
	GetAwaysForToday(fromDate string) StoreChannel
	DeleteAway(id string) StoreChannel
	DeleteOldAway(date string) StoreChannel
	//	DeleteAwayWithDate(date string) StoreChannel
}
