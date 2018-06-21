package model

import (
	"encoding/json"
	"io"
)

type Away struct {
	Id          int
	Start       string
	End         string
	UserName    string
	UserId      string
	Reason      string
	GoogleCalId string
}

func (o *Away) ToJson() (string, error) {
	if b, err := json.Marshal(o); err != nil {
		return "", err
	} else {
		return string(b), nil
	}
}

func AwayFromJson(data io.Reader) (*Away, error) {
	var away Away

	if err := json.NewDecoder(data).Decode(&away); err != nil {
		return nil, err
	} else {
		return &away, nil
	}
}
