package model

import (
	"encoding/json"
	"fmt"
)

type MMIntegrationResponse struct {
	EphemeralText string `json:"ephemeral_text,omitempty"`
}

func GenerateIntegrationResponse(text string) string {
	response := MMIntegrationResponse{
		EphemeralText: text,
	}

	b, err := json.Marshal(response)
	if err != nil {
		fmt.Println("Unable to marshal response")
		return ""
	}
	return string(b)
}
