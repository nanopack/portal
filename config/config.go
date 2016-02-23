package config

import (
	"fmt"
	"path/filepath"

	"github.com/jcelliott/lumber"
	"github.com/spf13/viper"
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
	ClusterConnection  string
	ClusterToken       string
	Insecure           bool
	LogLevel           string
	LogFile            string
	Log                lumber.Logger
)

func LoadConfigFile() {
	if ConfigFile != "" {
		// Set defaults to whatever might be there already
		viper.SetDefault("ApiToken", ApiToken)
		viper.SetDefault("ApiHost", ApiHost)
		viper.SetDefault("ApiPort", ApiPort)
		viper.SetDefault("ApiKey", ApiKey)
		viper.SetDefault("ApiCert", ApiCert)
		viper.SetDefault("ApiKeyPassword", ApiKeyPassword)
		viper.SetDefault("DatabaseConnection", DatabaseConnection)
		viper.SetDefault("ClusterConnection", ClusterConnection)
		viper.SetDefault("Insecure", Insecure)
		viper.SetDefault("LogLevel", LogLevel)
		viper.SetDefault("LogFile", LogFile)

		filename := filepath.Base(ConfigFile)
		viper.SetConfigName(filename[:len(filename)-len(filepath.Ext(filename))])
		viper.AddConfigPath(filepath.Dir(ConfigFile))

		err := viper.ReadInConfig()
		if err != nil {
			panic(fmt.Errorf("Fatal error config file: %s \n", err))
		}

		// Set values. Config file will override commandline
		ApiToken = viper.GetString("ApiToken")
		ApiHost = viper.GetString("ApiHost")
		ApiPort = viper.GetString("ApiPort")
		ApiKey = viper.GetString("ApiKey")
		ApiCert = viper.GetString("ApiCert")
		ApiKeyPassword = viper.GetString("ApiKeyPassword")
		DatabaseConnection = viper.GetString("DatabaseConnection")
		ClusterConnection = viper.GetString("ClusterConnection")
		Insecure = viper.GetBool("Insecure")
		LogLevel = viper.GetString("LogLevel")
		LogFile = viper.GetString("LogFile")
	}
}
