// commands is where all cli logic is, including starting portal as a server.
package commands

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
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
		return fmt.Errorf(fmt.Sprintf("portal %s (%s)", tag, commit))
	}

	if !config.Server {
		ccmd.HelpFunc()(ccmd, args)
		return fmt.Errorf("") // no error, just exit
	}

	return nil
}

func readConfig(ccmd *cobra.Command, args []string) error {
	if err := config.LoadConfigFile(); err != nil {
		return fmt.Errorf("ERROR: Failed to read config - %v", err)
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
			return fmt.Errorf("")
		}
	}

	// ensure proxy ports are unique. we need to check because tls will not listen
	// until a cert is added. we want it to break sooner.
	if config.RouteHttp == config.RouteTls {
		config.Log.Fatal("Proxy addresses must be unique")
		return fmt.Errorf("")
	}
	// need ':' in case tls is double apiport (8080, 80)
	apiPort := fmt.Sprintf(":%s", config.ApiPort)
	if strings.HasSuffix(config.RouteTls, apiPort) {
		config.Log.Fatal("TLS proxy address must be unique")
		return fmt.Errorf("")
	}

	// initialize database
	err := database.Init()
	if err != nil {
		config.Log.Fatal("Database init failed - %v", err)
		return fmt.Errorf("")
	}
	// initialize balancer
	err = balance.Init()
	if err != nil {
		config.Log.Fatal("Balancer init failed - %v", err)
		return fmt.Errorf("")
	}
	// initialize proxymgr
	err = proxymgr.Init()
	if err != nil {
		config.Log.Fatal("Proxymgr init failed - %v", err)
		return fmt.Errorf("")
	}
	// initialize vipmgr
	err = vipmgr.Init()
	if err != nil {
		config.Log.Fatal("Vipmgr init failed - %v", err)
		return fmt.Errorf("")
	}
	// initialize cluster
	err = cluster.Init()
	if err != nil {
		config.Log.Fatal("Cluster init failed - %v", err)
		return fmt.Errorf("")
	}

	go sigHandle()

	// start api
	err = api.StartApi()
	if err != nil {
		config.Log.Fatal("Api start failed - %v", err)
		return fmt.Errorf("")
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
		fmt.Printf("Failted to create request - %v\n", err.Error())
	}
	req.Header.Add("X-AUTH-TOKEN", config.ApiToken)
	res, err := client.Do(req)
	if err != nil {
		// if requesting `https://` failed, server may have been started with `-i`, try `http://`
		uri = fmt.Sprintf("http://%s:%s/%s", config.ApiHost, config.ApiPort, path)
		req, er := http.NewRequest(method, uri, body)
		if er != nil {
			fmt.Printf("Failted to create request - %v\n", er.Error())
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
