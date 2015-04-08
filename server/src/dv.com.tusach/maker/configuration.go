package maker

import (
	"encoding/json"
	"log"
	"os"
)

type Configuration struct {
	ServerPath        string `json:serverPath`
	DBFilename        string `json:dbFilename`
	ServerBindAddress string `json:serverBindAddress`
	ServerBindPort    int    `json:serverBindPort`
}

var configuration Configuration

func GetConfiguration() Configuration {
	return configuration
}

func LoadConfig(filename string) {
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			panic(err)
		}
	}()

	decoder := json.NewDecoder(f)
	err = decoder.Decode(&configuration)
	if err != nil {
		panic(err)
	}
	log.Printf("loaded configuration: %+v\n", configuration)
	if configuration.ServerPath == "" {
		panic("Missing config parameter: serverPath")
	}
	if configuration.DBFilename == "" {
		panic("Missing config parameter: dbFilename")
	}
}
