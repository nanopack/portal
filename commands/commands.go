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
	"github.com/nanopack/portal/vipmgr"
)

var (
	// to be populated by linker
	tag    string
	commit string

	Portal = &cobra.Command{
		Use:               "portal",
		Short:             "portal - load balancer/proxy",
		Long:              ``,
		PersistentPreRunE: readConfig,
		PreRunE:           preFlight,
		RunE:              startPortal,
		SilenceErrors:     true,
		SilenceUsage:      true,
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

	Portal.AddCommand(vipAddCmd)
	Portal.AddCommand(vipsSetCmd)
	Portal.AddCommand(vipsShowCmd)
	Portal.AddCommand(vipRemoveCmd)
}

func preFlight(ccmd *cobra.Command, args []string) error {
	if config.Version {
		fmt.Printf("portal %s (%s)\n", tag, commit)
		return fmt.Errorf("") // no error, just exit
	}

	if !config.Server {
		ccmd.HelpFunc()(ccmd, args)
		return fmt.Errorf("") // no error, just exit
	}

	return nil
}

func readConfig(ccmd *cobra.Command, args []string) error {
	if err := config.LoadConfigFile(); err != nil {
		fmt.Printf("ERROR: Failed to read config - %v\n", err)
		return err
	}

	return nil
}

func startPortal(ccmd *cobra.Command, args []string) error {
	if config.LogFile == "" {
		config.Log = lumber.NewConsoleLogger(lumber.LvlInt(config.LogLevel))
	} else {
		var err error
		config.Log, err = lumber.NewFileLogger(config.LogFile, lumber.LvlInt(config.LogLevel), lumber.ROTATE, 5000, 9, 100)
		if err != nil {
			config.Log.Fatal("File logger init failed - %v", err)
			return err
		}
	}
	// initialize database
	err := database.Init()
	if err != nil {
		config.Log.Fatal("Database init failed - %v", err)
		return err
	}
	// initialize balancer
	err = balance.Init()
	if err != nil {
		config.Log.Fatal("Balancer init failed - %v", err)
		return err
	}
	// initialize proxymgr
	err = proxymgr.Init()
	if err != nil {
		config.Log.Fatal("Proxymgr init failed - %v", err)
		return err
	}
	// initialize vipmgr
	err = vipmgr.Init()
	if err != nil {
		config.Log.Fatal("Vipmgr init failed - %v", err)
		return err
	}
	// initialize cluster
	err = cluster.Init()
	if err != nil {
		config.Log.Fatal("Cluster init failed - %v", err)
		return err
	}

	go sigHandle()

	// start api
	err = api.StartApi()
	if err != nil {
		config.Log.Fatal("Api start failed - %v", err)
		return err
	}
	return nil
}

func sigHandle() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		switch <-sigs {
		default:
			// clear balancer rules - (stop balancing if we are offline)
			balance.SetServices(make([]core.Service, 0, 0))
			// clear vips
			vipmgr.SetVips(make([]core.Vip, 0, 0))
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
	}

	req, err := http.NewRequest(method, uri, body)
	if err != nil {
		panic(err)
	}
	req.Header.Add("X-AUTH-TOKEN", config.ApiToken)
	res, err := client.Do(req)
	if err != nil {
		// if requesting `https://` failed, server may have been started with `-i`, try `http://`
		uri = fmt.Sprintf("http://%s:%s/%s", config.ApiHost, config.ApiPort, path)
		req, er := http.NewRequest(method, uri, body)
		if er != nil {
			panic(er)
		}
		req.Header.Add("X-AUTH-TOKEN", config.ApiToken)
		var err2 error
		res, err2 = client.Do(req)
		if err2 != nil {
			// return original error to client
			return nil, err
		}
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
