package commands_test

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jcelliott/lumber"

	"github.com/nanopack/portal/api"
	"github.com/nanopack/portal/balance"
	"github.com/nanopack/portal/cluster"
	"github.com/nanopack/portal/commands"
	"github.com/nanopack/portal/config"
	"github.com/nanopack/portal/database"
	"github.com/nanopack/portal/proxymgr"
)

type (
	// cobra.Command.Execute() 'alias'
	execable func() error

	testString struct {
		stringed string
		returned string
	}
)

var (
	apiAddr string
	Portal  = commands.Portal

	successMsg   = "{\"msg\":\"Success\"}\n"
	testService1 = testString{
		stringed: `{"host":"192.168.0.15","port":80,"type":"tcp","scheduler":"wrr","persistence":0,"netmask":""}`,
		returned: "{\"id\":\"tcp-192_168_0_15-80\",\"host\":\"192.168.0.15\",\"port\":80,\"type\":\"tcp\",\"scheduler\":\"wrr\",\"persistence\":0,\"netmask\":\"\"}\n",
	}
	testServices = testString{
		stringed: `[{"host":"192.168.0.15","port":80,"type":"tcp","scheduler":"wrr","persistence":0,"netmask":""},{"host":"192.168.0.16","port":443,"type":"tcp","scheduler":"wrr"}]`,
		returned: "[{\"id\":\"tcp-192_168_0_15-80\",\"host\":\"192.168.0.15\",\"port\":80,\"type\":\"tcp\",\"scheduler\":\"wrr\",\"persistence\":0,\"netmask\":\"\"},{\"id\":\"tcp-192_168_0_16-443\",\"host\":\"192.168.0.16\",\"port\":443,\"type\":\"tcp\",\"scheduler\":\"wrr\",\"persistence\":0,\"netmask\":\"\"}]\n",
	}
	testServer1 = testString{
		stringed: `{"host":"127.0.0.11","port":8080,"forwarder":"m","weight":5,"upper_threshold":10,"lower_threshold":1}`,
		returned: "{\"id\":\"127_0_0_11-8080\",\"host\":\"127.0.0.11\",\"port\":8080,\"forwarder\":\"m\",\"weight\":5,\"upper_threshold\":10,\"lower_threshold\":1}\n",
	}
	testServers = testString{
		stringed: `[{"host":"127.0.0.11","port":8080,"forwarder":"m","weight":5,"upper_threshold":10,"lower_threshold":1},{"host":"127.0.0.12","port":8080,"forwarder":"m","weight":5,"upper_threshold":10,"lower_threshold":1}]`,
		returned: "[{\"id\":\"127_0_0_11-8080\",\"host\":\"127.0.0.11\",\"port\":8080,\"forwarder\":\"m\",\"weight\":5,\"upper_threshold\":10,\"lower_threshold\":1},{\"id\":\"127_0_0_12-8080\",\"host\":\"127.0.0.12\",\"port\":8080,\"forwarder\":\"m\",\"weight\":5,\"upper_threshold\":10,\"lower_threshold\":1}]\n",
	}
)

func TestMain(m *testing.M) {
	// clean test dir
	os.RemoveAll("/tmp/cliTest")

	// manually configure
	initialize()

	// start api
	go api.StartApi()
	<-time.After(3 * time.Second)
	rtn := m.Run()

	// clean test dir
	os.RemoveAll("/tmp/cliTest")

	os.Exit(rtn)
}

////////////////////////////////////////////////////////////////////////////////
// SERVICES
////////////////////////////////////////////////////////////////////////////////
func TestShowServices(t *testing.T) {
	Portal.SetArgs(strings.Split("show-services", " "))

	out, err := capture(Portal.Execute)
	if err != nil {
		t.Errorf("Failed to execute - %v", err.Error())
	}

	if string(out) != "[]\n" {
		t.Errorf("Unexpected output: %q", string(out))
	}
}

func TestAddServiceFlags(t *testing.T) {
	Portal.SetArgs(strings.Split("add-service -O 192.168.0.15 -R 80 -T tcp -s wrr", " "))

	out, err := capture(Portal.Execute)
	if err != nil {
		t.Errorf("Failed to execute - %v", err.Error())
	}

	if string(out) != testService1.returned {
		t.Errorf("Unexpected output: %q", string(out))
	}
}

func TestAddServiceJson(t *testing.T) {
	args := strings.Split("add-service -j", " ")
	Portal.SetArgs(append(args, testService1.stringed))

	out, err := capture(Portal.Execute)
	if err != nil {
		t.Errorf("Failed to execute - %v", err.Error())
	}

	if string(out) != testService1.returned {
		t.Errorf("Unexpected output: %q", string(out))
	}
}

func TestRemoveServiceHost(t *testing.T) {
	args := strings.Split("remove-service -O 192.168.0.15 -R 80", " ")
	Portal.SetArgs(args)

	out, err := capture(Portal.Execute)
	if err != nil {
		t.Errorf("Failed to execute - %v", err.Error())
	}

	if string(out) != successMsg {
		t.Errorf("Unexpected output: %q", string(out))
	}
}

func TestRemoveServiceId(t *testing.T) {
	args := strings.Split("remove-service -I tcp-192_168_0_15-80", " ")
	Portal.SetArgs(args)

	out, err := capture(Portal.Execute)
	if err != nil {
		t.Errorf("Failed to execute - %v", err.Error())
	}

	if string(out) != successMsg {
		t.Errorf("Unexpected output: %q", string(out))
	}
}

func TestAddServices(t *testing.T) {
	args := strings.Split("set-services -j", " ")
	Portal.SetArgs(append(args, testServices.stringed))

	out, err := capture(Portal.Execute)
	if err != nil {
		t.Errorf("Failed to execute - %v", err.Error())
	}

	if string(out) != testServices.returned {
		t.Errorf("Unexpected output: %q", string(out))
	}
}

func TestShowServiceHost(t *testing.T) {
	args := strings.Split("show-service -O 192.168.0.15 -R 80", " ")
	Portal.SetArgs(args)

	out, err := capture(Portal.Execute)
	if err != nil {
		t.Errorf("Failed to execute - %v", err.Error())
	}

	if string(out) != testService1.returned {
		t.Errorf("Unexpected output: %q", string(out))
	}
}

func TestShowServiceId(t *testing.T) {
	args := strings.Split("show-service -I tcp-192_168_0_15-80", " ")
	Portal.SetArgs(args)

	out, err := capture(Portal.Execute)
	if err != nil {
		t.Errorf("Failed to execute - %v", err.Error())
	}

	if string(out) != testService1.returned {
		t.Errorf("Unexpected output: %q", string(out))
	}
}

////////////////////////////////////////////////////////////////////////////////
// SERVERS
////////////////////////////////////////////////////////////////////////////////
func TestShowServersHost(t *testing.T) {
	Portal.SetArgs(strings.Split("show-servers -O 192.168.0.15 -R 80", " "))

	out, err := capture(Portal.Execute)
	if err != nil {
		t.Errorf("Failed to execute - %v", err.Error())
	}

	if string(out) != "[]\n" {
		t.Errorf("Unexpected output: %q", string(out))
	}
}

func TestShowServersId(t *testing.T) {
	Portal.SetArgs(strings.Split("show-servers -I tcp-192_168_0_15-80", " "))

	out, err := capture(Portal.Execute)
	if err != nil {
		t.Errorf("Failed to execute - %v", err.Error())
	}

	if string(out) != "[]\n" {
		t.Errorf("Unexpected output: %q", string(out))
	}
}

func TestAddServerFlags(t *testing.T) {
	args := strings.Split("add-server -O 192.168.0.15 -R 80 -o 127.0.0.11 -p 8080 -f m -w 5 -u 10 -l 1", " ")
	Portal.SetArgs(args)

	out, err := capture(Portal.Execute)
	if err != nil {
		t.Errorf("Failed to execute - %v", err.Error())
	}

	if string(out) != testServer1.returned {
		t.Errorf("Unexpected output: %q", string(out))
	}
}

func TestAddServerJson(t *testing.T) {
	args := strings.Split("add-server -I tcp-192_168_0_15-80 -j", " ")
	Portal.SetArgs(append(args, testServer1.stringed))

	out, err := capture(Portal.Execute)
	if err != nil {
		t.Errorf("Failed to execute - %v", err.Error())
	}

	if string(out) != testServer1.returned {
		t.Errorf("Unexpected output: %q", string(out))
	}
}

func TestRemoveServerHost(t *testing.T) {
	args := strings.Split("remove-server -O 192.168.0.15 -R 80 -o 127.0.0.11 -p 8080", " ")
	Portal.SetArgs(args)

	out, err := capture(Portal.Execute)
	if err != nil {
		t.Errorf("Failed to execute - %v", err.Error())
	}

	if string(out) != successMsg {
		t.Errorf("Unexpected output: %q", string(out))
	}
}

func TestRemoveServerId(t *testing.T) {
	args := strings.Split("remove-server -I tcp-192_168_0_15-80 -S 127_0_0_11-8080", " ")
	Portal.SetArgs(args)

	out, err := capture(Portal.Execute)
	if err != nil {
		t.Errorf("Failed to execute - %v", err.Error())
	}

	if string(out) != successMsg {
		t.Errorf("Unexpected output: %q", string(out))
	}
}

func TestSetServers(t *testing.T) {
	args := strings.Split("set-servers -I tcp-192_168_0_15-80 -j", " ")
	Portal.SetArgs(append(args, testServers.stringed))

	out, err := capture(Portal.Execute)
	if err != nil {
		t.Errorf("Failed to execute - %v", err.Error())
	}

	if string(out) != testServers.returned {
		t.Errorf("Unexpected output: %q", string(out))
	}
}

func TestShowServerHost(t *testing.T) {
	args := strings.Split("show-server -O 192.168.0.15 -R 80 -o 127.0.0.11 -p 8080", " ")
	Portal.SetArgs(args)

	out, err := capture(Portal.Execute)
	if err != nil {
		t.Errorf("Failed to execute - %v", err.Error())
	}

	if string(out) != testServer1.returned {
		t.Errorf("Unexpected output: %q", string(out))
	}
}

func TestShowServerId(t *testing.T) {
	args := strings.Split("show-server -I tcp-192_168_0_15-80 -S 127_0_0_11-8080", " ")
	Portal.SetArgs(args)

	out, err := capture(Portal.Execute)
	if err != nil {
		t.Errorf("Failed to execute - %v", err.Error())
	}

	if string(out) != testServer1.returned {
		t.Errorf("Unexpected output: %q", string(out))
	}
}

////////////////////////////////////////////////////////////////////////////////
// ROUTES
////////////////////////////////////////////////////////////////////////////////
func TestShowRoutes(t *testing.T) {
	Portal.SetArgs(strings.Split("show-routes", " "))

	out, err := capture(Portal.Execute)
	if err != nil {
		t.Errorf("Failed to execute - %v", err.Error())
	}

	if string(out) != "[]\n" {
		t.Errorf("Unexpected output: %q", string(out))
	}
}

func TestAddRoute(t *testing.T) {
	Portal.SetArgs(strings.Split("add-route -j {\"domain\":\"portal.test\",\"page\":\"testing\"}", " "))

	out, err := capture(Portal.Execute)
	if err != nil {
		t.Errorf("Failed to execute - %v", err.Error())
	}

	if string(out) != "{\"subdomain\":\"\",\"domain\":\"portal.test\",\"path\":\"\",\"targets\":null,\"fwdpath\":\"\",\"page\":\"testing\"}\n" {
		t.Errorf("Unexpected output: %q", string(out))
	}
}

func TestRemoveRoute(t *testing.T) {
	args := strings.Split("remove-route -d portal.test", " ")
	Portal.SetArgs(args)

	out, err := capture(Portal.Execute)
	if err != nil {
		t.Errorf("Failed to execute - %v", err.Error())
	}

	if string(out) != successMsg {
		t.Errorf("Unexpected output: %q", string(out))
	}
}

func TestSetRoutes(t *testing.T) {
	Portal.SetArgs(strings.Split("set-routes -j [{\"domain\":\"portal.test\",\"page\":\"testing\"}]", " "))

	out, err := capture(Portal.Execute)
	if err != nil {
		t.Errorf("Failed to execute - %v", err.Error())
	}

	if string(out) != "[{\"subdomain\":\"\",\"domain\":\"portal.test\",\"path\":\"\",\"targets\":null,\"fwdpath\":\"\",\"page\":\"testing\"}]\n" {
		t.Errorf("Unexpected output: %q", string(out))
	}
}

////////////////////////////////////////////////////////////////////////////////
// CERTS
////////////////////////////////////////////////////////////////////////////////
func TestShowCerts(t *testing.T) {
	Portal.SetArgs(strings.Split("show-certs", " "))

	out, err := capture(Portal.Execute)
	if err != nil {
		t.Errorf("Failed to execute - %v", err.Error())
	}

	if string(out) != "[]\n" {
		t.Errorf("Unexpected output: %q", string(out))
	}
}

func TestAddCert(t *testing.T) {
	Portal.SetArgs(strings.Split("add-cert -j {\"key\":\"portal.test\",\"cert\":\"certified\"}", " "))

	out, err := capture(Portal.Execute)
	if err != nil {
		t.Errorf("Failed to execute - %v", err.Error())
	}

	if !strings.Contains(string(out), "error") {
		t.Errorf("Unexpected output: %q", string(out))
	}
}

func TestRemoveCert(t *testing.T) {
	args := strings.Split("remove-cert -j {\"key\":\"portal.test\",\"cert\":\"certified\"}", " ")
	Portal.SetArgs(args)

	out, err := capture(Portal.Execute)
	if err != nil {
		t.Errorf("Failed to execute - %v", err.Error())
	}

	if string(out) != successMsg {
		t.Errorf("Unexpected output: %q", string(out))
	}
}

func TestSetCerts(t *testing.T) {
	Portal.SetArgs(strings.Split("set-certs -j [{\"key\":\"portal.test\",\"cert\":\"certified\"}]", " "))

	out, err := capture(Portal.Execute)
	if err != nil {
		t.Errorf("Failed to execute - %v", err.Error())
	}

	if !strings.Contains(string(out), "error") {
		t.Errorf("Unexpected output: %q", string(out))
	}
}

func TestRunServer(t *testing.T) {
	config.JustProxy = true
	config.RouteHttp = "0.0.0.0:9085"
	config.RouteTls = "0.0.0.0:9448"

	Portal.SetArgs(strings.Split("-s -d /tmp/portalServer -l FATAL -P 8446", " "))

	go Portal.Execute()
	time.Sleep(time.Second)
}

////////////////////////////////////////////////////////////////////////////////
// PRIVS
////////////////////////////////////////////////////////////////////////////////
func capture(fn execable) ([]byte, error) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := fn()
	os.Stdout = oldStdout
	w.Close() // do not defer after os.Pipe()
	if err != nil {
		return nil, err
	}

	return ioutil.ReadAll(r)
}

// manually configure and start internals
func initialize() {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	config.DatabaseConnection = "scribble:///tmp/cliTest"
	config.ClusterConnection = "none://"
	config.ApiHost = "127.0.0.1"
	config.ApiPort = "8445"
	config.ApiToken = ""
	config.RouteHttp = "0.0.0.0:9081"
	config.RouteTls = "0.0.0.0:9444"
	config.LogLevel = "FATAL"
	config.Log = lumber.NewConsoleLogger(lumber.LvlInt(config.LogLevel))
	apiAddr = fmt.Sprintf("%v:%v", config.ApiHost, config.ApiPort)

	// initialize database
	err := database.Init()
	if err != nil {
		fmt.Printf("Database init failed - %v\n", err)
		os.Exit(1)
	}
	// initialize balancer
	balance.Balancer = &database.ScribbleDatabase{}
	err = balance.Balancer.Init()
	if err != nil {
		fmt.Printf("Balancer init failed - %v\n", err)
		os.Exit(1)
	}
	// initialize proxymgr
	err = proxymgr.Init()
	if err != nil {
		fmt.Printf("Proxymgr init failed - %v\n", err)
		os.Exit(1)
	}
	// initialize clusterer
	err = cluster.Init()
	if err != nil {
		fmt.Printf("Clusterer init failed - %v\n", err)
		os.Exit(1)
	}
}
