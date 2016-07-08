// config is a central location for configuration options. It also contains
// config file parsing logic.
package config

import (
	"fmt"
	"path/filepath"

	"github.com/jcelliott/lumber"
	"github.com/spf13/cobra"
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
	ProxyIgnore        = false
	RouteHttp          = "0.0.0.0:80"
	RouteTls           = "0.0.0.0:443"
	Balancer           = "lvs"
	WorkDir            = "/var/db/portal"
	JustProxy          = false
	Server             = false
	Version            = false
)

func AddFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().BoolVarP(&Insecure, "insecure", "i", Insecure, "Disable tls key checking (client) and listen on http (server)")
	cmd.PersistentFlags().StringVarP(&ApiToken, "api-token", "t", ApiToken, "Token for API Access")
	cmd.PersistentFlags().StringVarP(&ApiHost, "api-host", "H", ApiHost, "Listen address for the API")
	cmd.PersistentFlags().StringVarP(&ApiPort, "api-port", "P", ApiPort, "Listen address for the API")
	cmd.PersistentFlags().StringVarP(&ConfigFile, "conf", "c", ConfigFile, "Configuration file to load")

	cmd.Flags().StringVarP(&ApiKey, "api-key", "k", ApiKey, "SSL key for the api")
	cmd.Flags().StringVarP(&ApiCert, "api-cert", "C", ApiCert, "SSL cert for the api")
	cmd.Flags().StringVarP(&ApiKeyPassword, "api-key-password", "p", ApiKeyPassword, "Password for the SSL key")
	cmd.Flags().StringVarP(&DatabaseConnection, "db-connection", "d", DatabaseConnection, "Database connection string")
	cmd.Flags().StringVarP(&ClusterConnection, "cluster-connection", "r", ClusterConnection, "Cluster connection string (redis://127.0.0.1:6379)")
	cmd.Flags().StringVarP(&ClusterToken, "cluster-token", "T", ClusterToken, "Cluster security token")
	cmd.Flags().StringVarP(&LogLevel, "log-level", "l", LogLevel, "Log level to output")
	cmd.Flags().StringVarP(&LogFile, "log-file", "L", LogFile, "Log file to write to")
	cmd.Flags().BoolVarP(&ProxyIgnore, "proxy-ignore-upstream", "u", Server, "Ignore upstream's(target's) certs when routing")
	cmd.Flags().StringVarP(&RouteHttp, "proxy-http", "x", RouteHttp, "Address to listen on for proxying http")
	cmd.Flags().StringVarP(&RouteTls, "proxy-tls", "X", RouteTls, "Address to listen on for proxying https")
	cmd.Flags().StringVarP(&Balancer, "balancer", "b", Balancer, "Load balancer to use (nginx|lvs)")
	cmd.Flags().StringVarP(&WorkDir, "work-dir", "w", WorkDir, "Directory for portal to use (balancer config)")

	cmd.Flags().BoolVarP(&Server, "server", "s", Server, "Run in server mode")
	cmd.Flags().BoolVarP(&Version, "version", "v", Version, "Print version info and exit")
	cmd.Flags().BoolVarP(&JustProxy, "just-proxy", "j", JustProxy, "Proxy only (no tcp/udp load balancing)")

}

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
	viper.SetDefault("just-proxy", JustProxy)
	viper.SetDefault("log-level", LogLevel)
	viper.SetDefault("log-file", LogFile)
	viper.SetDefault("server", Server)
	viper.SetDefault("proxy-http", RouteHttp)
	viper.SetDefault("proxy-tls", RouteTls)
	viper.SetDefault("proxy-ignore-upstream", ProxyIgnore)
	viper.SetDefault("balancer", Balancer)
	viper.SetDefault("work-dir", WorkDir)

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
	JustProxy = viper.GetBool("just-proxy")
	LogLevel = viper.GetString("log-level")
	LogFile = viper.GetString("log-file")
	Server = viper.GetBool("server")
	RouteHttp = viper.GetString("proxy-http")
	RouteTls = viper.GetString("proxy-tls")
	ProxyIgnore = viper.GetBool("proxy-ignore-upstream")
	Balancer = viper.GetString("balancer")
	WorkDir = viper.GetString("work-dir")

	return nil
}
