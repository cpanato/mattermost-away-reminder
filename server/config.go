package server

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type AwayServerConfig struct {
	ListenAddress     string
	DriverName        string
	DataSource        string
	MMIncomingWebhook string
	AllowedTokens     []string
}

var Config *AwayServerConfig = &AwayServerConfig{}

func FindConfigFile(fileName string) string {
	if _, err := os.Stat("/tmp/" + fileName); err == nil {
		fileName, _ = filepath.Abs("/tmp/" + fileName)
	} else if _, err := os.Stat("./config/" + fileName); err == nil {
		fileName, _ = filepath.Abs("./config/" + fileName)
	} else if _, err := os.Stat("../config/" + fileName); err == nil {
		fileName, _ = filepath.Abs("../config/" + fileName)
	} else if _, err := os.Stat(fileName); err == nil {
		fileName, _ = filepath.Abs(fileName)
	}

	return fileName
}

func LoadConfig(fileName string) {
	fileName = FindConfigFile(fileName)
	fmt.Println("Loading " + fileName)

	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println("Error opening config file=" + fileName + ", err=" + err.Error())
	}

	decoder := json.NewDecoder(file)
	err = decoder.Decode(Config)
	if err != nil {
		fmt.Println("Error decoding config file=" + fileName + ", err=" + err.Error())
	}
}
