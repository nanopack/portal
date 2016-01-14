package commands

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"

	"github.com/jcelliott/lumber"
	"github.com/spf13/cobra"

	"github.com/nanopack/portal/api"
	"github.com/nanopack/portal/config"
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
	Portal.PersistentFlags().BoolVarP(&config.Insecure, "insecure", "i", false, "Disable tls key checking")
	Portal.PersistentFlags().StringVarP(&config.ApiToken, "api-token", "t", "",
		"Token for API Access")
	Portal.PersistentFlags().StringVarP(&config.ApiHost, "api-host", "H", "127.0.0.1",
		"Listen address for the API")
	Portal.PersistentFlags().StringVarP(&config.ApiPort, "api-port", "P", "8443",
		"Listen address for the API")
	Portal.PersistentFlags().StringVarP(&config.ConfigFile, "conf", "c", "",
		"Configuration file to load")

	Portal.Flags().BoolVarP(&runServer, "server", "s", false, "Run in server mode")

	Portal.AddCommand(serviceAddCmd)
	Portal.AddCommand(serviceRemoveCmd)
	Portal.AddCommand(serviceShowCmd)
	Portal.AddCommand(servicesShowCmd)
	Portal.AddCommand(servicesSetCmd)

	Portal.AddCommand(serverAddCmd)
	Portal.AddCommand(serverRemoveCmd)
	Portal.AddCommand(serverShowCmd)
	Portal.AddCommand(serversShowCmd)
	Portal.AddCommand(serversSetCmd)

	Portal.AddCommand(syncLvsCmd)
	Portal.AddCommand(syncPortalCmd)
}

func startServer() {
	if config.LogFile == "" {
		config.Log = lumber.NewConsoleLogger(lumber.LvlInt(config.LogLevel))
	} else {
		var err error
		config.Log, err = lumber.NewFileLogger(config.LogFile, lumber.LvlInt(config.LogLevel), lumber.ROTATE, 5000, 9, 100)
		if err != nil {
			panic(err)
		}
	}
	// initialize database
	// load saved rules
	err := database.SyncToLvs()
	if err != nil {
		panic(err)
	}
	// start api
	err = api.StartApi()
	if err != nil {
		panic(err)
	}
	return
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
	return client.Do(req)
}
