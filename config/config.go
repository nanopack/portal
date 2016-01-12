package config

import (
	"github.com/jcelliott/lumber"
)

var (
	ApiToken           string
	ApiHost            string
	ApiPort            string
	ApiKey             string
	ApiCert            string
	ApiKeyPassword     string
	ConfigFile         string
	DatabaseConnection string
	Insecure           bool
	LogLevel           string
	LogFile            string
	Log                lumber.Logger
)

func LoadConfigFile() {

}
