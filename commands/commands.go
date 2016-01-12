package commands

import (
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

	// Portal.AddCommand()
	// service-add
	// service-remove
	// service-show
	// server-add
	// server-remove
	// server-show
	// services-show
	// services-set
	// servers-show
	// servers-set
	// sync-lvs
	// sync-portal

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
