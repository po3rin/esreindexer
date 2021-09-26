package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
	"github.com/po3rin/esreindexer/logger"
)

var Conf Config

type Config struct {
	ApiPort   int    `default:"8888"`
	EsAddress string `default:"http://localhost:9200"`
	EsUser    string
	EsPass    string
}

func init() {
	if err := envconfig.Process("reindexer", &Conf); err != nil {
		logger.L.Fatal(fmt.Sprintf("[ERROR] Failed to process env: %s", err.Error()))
	}
}
