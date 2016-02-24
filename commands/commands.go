package commands

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/jcelliott/lumber"
	"github.com/spf13/cobra"

	"github.com/nanopack/portal/api"
	"github.com/nanopack/portal/balance"
	"github.com/nanopack/portal/cluster"
	"github.com/nanopack/portal/config"
	"github.com/nanopack/portal/core"
	"github.com/nanopack/portal/database"
)

var (
	runServer bool
	Portal    = &cobra.Command{
		Use:   "",
		Short: "",
		Long:  ``,

		Run: func(ccmd *cobra.Command, args []string) {
			if runServer {
				startServer()
				return
			}
			// Show the help if not starting the server
			ccmd.HelpFunc()(ccmd, args)
		},
	}
)

func init() {
	// Portal.PersistentFlags()
	Portal.PersistentFlags().BoolVarP(&config.Insecure, "insecure", "i", true, "Disable tls key checking")
	Portal.PersistentFlags().StringVarP(&config.ApiToken, "api-token", "t", "", "Token for API Access")
	Portal.PersistentFlags().StringVarP(&config.ApiHost, "api-host", "H", "127.0.0.1", "Listen address for the API")
	Portal.PersistentFlags().StringVarP(&config.ApiPort, "api-port", "P", "8443", "Listen address for the API")
	Portal.PersistentFlags().StringVarP(&config.ConfigFile, "conf", "c", "", "Configuration file to load")

	Portal.Flags().StringVarP(&config.ApiKey, "api-key", "k", "", "SSL key for the api")
	Portal.Flags().StringVarP(&config.ApiCert, "api-crt", "C", "", "SSL cert for the api")
	Portal.Flags().StringVarP(&config.ApiKeyPassword, "api-key-password", "p", "", "Password for the SSL key")
	Portal.Flags().StringVarP(&config.DatabaseConnection, "db-connection", "d", "scribble:///var/db/portal", "Database connection string")
	Portal.Flags().StringVarP(&config.ClusterConnection, "cluster-connection", "r", "redis://127.0.0.1:6379", "Cluster connection string")
	Portal.Flags().StringVarP(&config.ClusterToken, "cluster-token", "T", "", "Cluster security token")
	Portal.Flags().StringVarP(&config.LogLevel, "log-level", "L", "INFO", "Log level to output")
	Portal.Flags().StringVarP(&config.LogFile, "log-file", "l", "", "Log file to write to")

	Portal.Flags().BoolVarP(&runServer, "server", "s", false, "Run in server mode")

	Portal.AddCommand(serviceAddCmd)
	Portal.AddCommand(serviceRemoveCmd)
	Portal.AddCommand(serviceShowCmd)
	Portal.AddCommand(servicesShowCmd)
	Portal.AddCommand(servicesSetCmd)
	Portal.AddCommand(serviceSetCmd)

	Portal.AddCommand(serverAddCmd)
	Portal.AddCommand(serverRemoveCmd)
	Portal.AddCommand(serverShowCmd)
	Portal.AddCommand(serversShowCmd)
	Portal.AddCommand(serversSetCmd)
}

func startServer() {
	config.LoadConfigFile()
	if config.LogFile == "" {
		config.Log = lumber.NewConsoleLogger(lumber.LvlInt(config.LogLevel))
	} else {
		var err error
		config.Log, err = lumber.NewFileLogger(config.LogFile, lumber.LvlInt(config.LogLevel), lumber.ROTATE, 5000, 9, 100)
		if err != nil {
			config.Log.Fatal("File logger init failed - %v", err)
			os.Exit(1)
		}
	}
	// initialize database
	err := database.Init()
	if err != nil {
		config.Log.Fatal("Database init failed - %v", err)
		os.Exit(1)
	}
	// initialize balancer
	err = balance.Init()
	if err != nil {
		config.Log.Fatal("Balancer init failed - %v", err)
		os.Exit(1)
	}
	// load saved rules
	services, err := database.GetServices()
	if err != nil {
		config.Log.Fatal("Get services from backend failed - %v", err)
		os.Exit(1)
	}
	// apply saved rules
	err = balance.SetServices(services)
	if err != nil {
		config.Log.Fatal("Balancer sync failed - %v", err)
		os.Exit(1)
	}
	// initialize cluster
	err = cluster.Init()
	if err != nil {
		config.Log.Fatal("Cluster init failed - %v", err)
		os.Exit(1)
	}

	go sigHandle()

	// start api
	err = api.StartApi()
	if err != nil {
		config.Log.Fatal("Api start failed - %v", err)
		os.Exit(1)
	}
	return
}

func sigHandle() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		switch <-sigs {
		default:
			// clear balancer rules - (stop balancing if we are offline)
			balance.SetServices(make([]core.Service, 0, 0))
			fmt.Println()
			os.Exit(0)
		}
	}()
}

func rest(path string, method string, body io.Reader) (*http.Response, error) {
	var client *http.Client
	if config.Insecure {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client = &http.Client{Transport: tr}
	} else {
		client = http.DefaultClient
	}

	uri := fmt.Sprintf("https://%s:%s/%s", config.ApiHost, config.ApiPort, path)
	req, err := http.NewRequest(method, uri, body)
	if err != nil {
		panic(err)
	}
	req.Header.Add("X-NANOBOX-TOKEN", config.ApiToken)
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode == 401 {
		return nil, fmt.Errorf("401 Unauthorized. Please specify api token (-t 'token')")
	}
	return res, nil
}

func fail(format string, args ...interface{}) {
	fmt.Printf(fmt.Sprintf("%v\n", format), args...)
	os.Exit(1)
}
