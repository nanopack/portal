// commands is where all cli logic is, including starting portal as a server.
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
	"github.com/nanopack/portal/proxymgr"
)

var (
	runServer bool
	Portal    = &cobra.Command{
		Use:   "portal",
		Short: "portal - load balancer/proxy",
		Long:  ``,

		Run: startPortal,
	}
)

func init() {
	config.AddFlags(Portal)
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

	Portal.AddCommand(routeAddCmd)
	Portal.AddCommand(routesSetCmd)
	Portal.AddCommand(routesShowCmd)
	Portal.AddCommand(routeRemoveCmd)

	Portal.AddCommand(certAddCmd)
	Portal.AddCommand(certsSetCmd)
	Portal.AddCommand(certsShowCmd)
	Portal.AddCommand(certRemoveCmd)
}

func startPortal(ccmd *cobra.Command, args []string) {
	if err := config.LoadConfigFile(); err != nil {
		config.Log.Fatal("Failed to read config - %v", err)
		os.Exit(1)
	}

	if !config.Server {
		ccmd.HelpFunc()(ccmd, args)
		return
	}

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
	// initialize proxymgr
	err = proxymgr.Init()
	if err != nil {
		config.Log.Fatal("Proxymgr init failed - %v", err)
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
	client = http.DefaultClient
	uri := fmt.Sprintf("https://%s:%s/%s", config.ApiHost, config.ApiPort, path)

	if config.Insecure {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		uri = fmt.Sprintf("http://%s:%s/%s", config.ApiHost, config.ApiPort, path)
	}

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
