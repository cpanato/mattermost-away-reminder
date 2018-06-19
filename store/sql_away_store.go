package store

import (
	"database/sql"
	"fmt"

	"github.com/cpanato/mattermost-away-reminder/model"
)

type SqlAwayStore struct {
	*SqlStore
}

func NewSqlAwayStore(sqlStore *SqlStore) AwayStore {
	s := &SqlAwayStore{sqlStore}

	for _, db := range sqlStore.GetAllConns() {
		table := db.AddTableWithName(model.Away{}, "Away").SetKeys(true, "Id")
		table.ColMap("Start").SetMaxSize(128)
		table.ColMap("End").SetMaxSize(128)
		table.ColMap("UserName").SetMaxSize(128)
		table.ColMap("Reason").SetMaxSize(128)
	}

	return s
}

func (s SqlAwayStore) CreateIndexesIfNotExists() {
	s.CreateColumnIfNotExists("Away", "Start", "varchar(128)", "varchar(128)", "")
	s.CreateColumnIfNotExists("Away", "End", "varchar(128)", "varchar(128)", "")
}

func (s SqlAwayStore) Save(away *model.Away) StoreChannel {
	storeChannel := make(StoreChannel)

	go func() {
		result := StoreResult{}

		if err := s.GetMaster().Insert(away); err != nil {
			if _, err := s.GetMaster().Update(away); err != nil {
				result.Err = model.NewLocAppError("Mattermost Time Away", "Could not insert or update Away table", nil,
					fmt.Sprintf("UserName=%v, Start=%v, End=%v, Reason=%v, err=%v", away.UserName, away.Start, away.End, away.Reason, err.Error()))
			}
		}

		if result.Err == nil {
			result.Data = away
		}

		storeChannel <- result
		close(storeChannel)
	}()

	return storeChannel
}

func (s SqlAwayStore) GetAwaysForToday(todayDate string) StoreChannel {
	storeChannel := make(StoreChannel)

	go func() {
		result := StoreResult{}

		var aways []*model.Away
		if _, err := s.GetReplica().Select(&aways,
			`SELECT
        *
      FROM
        Away
      WHERE
        Start <= :todayDate and End >= :todayDate`, map[string]interface{}{"todayDate": todayDate}); err != nil {
			if err != sql.ErrNoRows {
				result.Err = model.NewLocAppError("Mattermost Time Away", "Could not get any away", nil,
					fmt.Sprintf("todayDate=%v, err=%v", todayDate, err.Error()))
			} else {
				result.Data = nil
			}
		} else {
			result.Data = aways
		}

		storeChannel <- result
		close(storeChannel)
	}()

	return storeChannel
}

func (s SqlAwayStore) GetAwaysById(id string) StoreChannel {
	storeChannel := make(StoreChannel)

	go func() {
		result := StoreResult{}

		var away *model.Away
		if err := s.GetReplica().SelectOne(&away,
			`SELECT
        *
      FROM
        Away
      WHERE Id = :id`, map[string]interface{}{"id": id}); err != nil {
			result.Err = model.NewLocAppError("Mattermost Time Away", "Could not list Away", nil, err.Error())
		} else {
			result.Data = away
		}

		storeChannel <- result
		close(storeChannel)
	}()

	return storeChannel
}

func (s SqlAwayStore) ListAll() StoreChannel {
	storeChannel := make(StoreChannel)

	go func() {
		result := StoreResult{}

		var aways []*model.Away
		if _, err := s.GetReplica().Select(&aways,
			`SELECT
        *
      FROM
        Away`); err != nil {
			result.Err = model.NewLocAppError("Mattermost Time Away", "Could not list Away", nil, err.Error())
		} else {
			result.Data = aways
		}

		storeChannel <- result
		close(storeChannel)
	}()

	return storeChannel
}

func (s SqlAwayStore) ListUserAway(userName string) StoreChannel {
	storeChannel := make(StoreChannel)

	go func() {
		result := StoreResult{}

		var aways []*model.Away
		if _, err := s.GetReplica().Select(&aways,
			`SELECT
        *
      FROM
        Away
      WHERE UserName = :userName`, map[string]interface{}{"userName": userName}); err != nil {
			result.Err = model.NewLocAppError("Mattermost Time Away", "Could not list Away", nil, err.Error())
		} else {
			result.Data = aways
		}

		storeChannel <- result
		close(storeChannel)
	}()

	return storeChannel
}

func (s SqlAwayStore) DeleteAway(id string) StoreChannel {
	storeChannel := make(StoreChannel)

	go func() {
		result := StoreResult{}

		if _, err := s.GetMaster().Exec("DELETE FROM Away WHERE Id = :id", map[string]interface{}{"id": id}); err != nil {
			result.Err = model.NewLocAppError("Mattermost Time Away", "Could not delete the away", nil, "id="+id+", err="+err.Error())
		}

		storeChannel <- result
		close(storeChannel)
	}()

	return storeChannel
}

func (s SqlAwayStore) DeleteOldAway(date string) StoreChannel {
	storeChannel := make(StoreChannel)

	go func() {
		result := StoreResult{}

		if _, err := s.GetMaster().Exec("DELETE FROM Away WHERE End < :date", map[string]interface{}{"date": date}); err != nil {
			result.Err = model.NewLocAppError("Mattermost Time Away", "Could not delete the away", nil, "date="+date+", err="+err.Error())
		}

		storeChannel <- result
		close(storeChannel)
	}()

	return storeChannel
}
