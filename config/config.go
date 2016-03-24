package config

import (
	"fmt"
	"path/filepath"

	"github.com/jcelliott/lumber"
	"github.com/spf13/viper"
)

var (
	ApiToken           = ""
	ApiHost            = "127.0.0.1"
	ApiPort            = "8443"
	ApiKey             = ""
	ApiCert            = ""
	ApiKeyPassword     = ""
	ConfigFile         = ""
	DatabaseConnection = "scribble:///var/db/portal"
	ClusterConnection  = "none://"
	ClusterToken       = ""
	Insecure           = false
	LogLevel           = "INFO"
	LogFile            = ""
	Log                lumber.Logger
	RouteHttp          = "0.0.0.0:80"
	RouteTls           = "0.0.0.0:443"
	JustProxy          = false
	Server             = false
)

func LoadConfigFile() error {
	if ConfigFile == "" {
		return nil
	}
	// Set defaults to whatever might be there already
	viper.SetDefault("api-token", ApiToken)
	viper.SetDefault("api-host", ApiHost)
	viper.SetDefault("api-port", ApiPort)
	viper.SetDefault("api-key", ApiKey)
	viper.SetDefault("api-cert", ApiCert)
	viper.SetDefault("api-key-password", ApiKeyPassword)
	viper.SetDefault("db-connection", DatabaseConnection)
	viper.SetDefault("cluster-connection", ClusterConnection)
	viper.SetDefault("cluster-token", ClusterToken)
	viper.SetDefault("insecure", Insecure)
	viper.SetDefault("log-level", LogLevel)
	viper.SetDefault("log-file", LogFile)
	viper.SetDefault("server", Server)

	filename := filepath.Base(ConfigFile)
	viper.SetConfigName(filename[:len(filename)-len(filepath.Ext(filename))])
	viper.AddConfigPath(filepath.Dir(ConfigFile))

	err := viper.ReadInConfig()
	if err != nil {
		return fmt.Errorf("Fatal error config file: %s \n", err)
	}

	// Set values. Config file will override commandline
	ApiToken = viper.GetString("api-token")
	ApiHost = viper.GetString("api-host")
	ApiPort = viper.GetString("api-port")
	ApiKey = viper.GetString("api-key")
	ApiCert = viper.GetString("api-cert")
	ApiKeyPassword = viper.GetString("api-key-password")
	DatabaseConnection = viper.GetString("db-connection")
	ClusterConnection = viper.GetString("cluster-connection")
	ClusterToken = viper.GetString("cluster-token")
	Insecure = viper.GetBool("insecure")
	LogLevel = viper.GetString("log-level")
	LogFile = viper.GetString("log-file")
	Server = viper.GetBool("server")

	return nil
}
