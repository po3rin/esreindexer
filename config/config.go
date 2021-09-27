package config

import (
	"log"
	"time"

	"github.com/kelseyhightower/envconfig"
)

var Conf Config

type Config struct {
	LoggingLevel   string `default:"DEBUG"`
	ApiPort        int    `default:"8888"`
	EsAddress      string `default:"http://localhost:9200"`
	EsUser         string
	EsPass         string
	ExpireDuration time.Duration `default:"48h"`
}

func init() {
	if err := envconfig.Process("reindexer", &Conf); err != nil {
		log.Fatalf("[ERROR] Failed to process env: %s", err.Error())
	}
}
