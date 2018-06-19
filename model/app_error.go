package model

import (
	"encoding/json"
)

type AppError struct {
	Id            string                 `json:"id"`
	Message       string                 `json:"message"`               // Message to be display to the end user without debugging information
	DetailedError string                 `json:"detailed_error"`        // Internal error string to help the developer
	RequestId     string                 `json:"request_id,omitempty"`  // The RequestId that's also set in the header
	StatusCode    int                    `json:"status_code,omitempty"` // The http status code
	Where         string                 `json:"-"`                     // The function where it happened in the form of Struct.Func
	IsOAuth       bool                   `json:"is_oauth,omitempty"`    // Whether the error is OAuth specific
	params        map[string]interface{} `json:"-"`
}

func (er *AppError) Error() string {
	return er.Where + ": " + er.Message + " " + er.DetailedError
}

func (er *AppError) ToJson() string {
	b, err := json.Marshal(er)
	if err != nil {
		return ""
	} else {
		return string(b)
	}
}

func NewLocAppError(where string, id string, params map[string]interface{}, details string) *AppError {
	ap := &AppError{}
	ap.Id = id
	ap.params = params
	ap.Message = id
	ap.Where = where
	ap.DetailedError = details
	ap.StatusCode = 500
	ap.IsOAuth = false
	return ap
}
