package api_test

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
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
	"github.com/nanopack/portal/config"
	"github.com/nanopack/portal/core"
	"github.com/nanopack/portal/database"
)

var (
	apiAddr string

	testService  = `{ "host": "192.168.0.15", "port": 80, "type": "tcp", "scheduler": "wrr" }`
	testService2 = `{ "host": "192.168.0.16", "port": 443, "type": "tcp", "scheduler": "wrr" }`
	badService   = `{ "port": "80", "type": "tcp", "scheduler": "wrr" }`
	testServer1  = `{ "host": "127.0.0.11", "port": 8080, "forwarder": "m", "weight": 5 }`
	testServer2  = `{ "host": "127.0.0.12", "port": 8080, "forwarder": "m", "weight": 5 }`
	testServer3  = `{ "host": "127.0.0.13", "port": 8080, "forwarder": "m", "weight": 5 }`
)

func TestMain(m *testing.M) {
	// clean test dir
	os.RemoveAll("/tmp/portalTest")

	// manually configure
	initialize()

	// start api
	go api.StartApi()
	<-time.After(3 * time.Second)
	rtn := m.Run()

	// clean test dir
	os.RemoveAll("/tmp/portalTest")

	os.Exit(rtn)
}

// test all routes for api response
// 1 with good data; at least 1 bad (where applicable)
// Get     - /services                         getServices
// Put     - /services                         putServices
// Post    - /services                         postService
// Get     - /services/{svcId}                 getService
// Put     - /services/{svcId}                 putService
// Delete  - /services/{svcId}                 deleteService
////////////////////////////////////////////////////////////////////////////////
// Get     - /services/{svcId}/servers         getServers
// Put     - /services/{svcId}/servers         putServers
// Post    - /services/{svcId}/servers         postServer
// Get     - /services/{svcId}/servers/{srvId} getServer
// Delete  - /services/{svcId}/servers/{srvId} deleteServer
// Post    - /sync                             postSync
// Get     - /sync                             getSync

////////////////////////////////////////////////////////////////////////////////
// SERVICES
////////////////////////////////////////////////////////////////////////////////
// test get services
func TestGetServices(t *testing.T) {
	body, err := rest("GET", "/services", "")
	if err != nil {
		t.Error(err)
	}
	if string(body) != "[]\n" {
		t.Errorf("%q doesn't match expected out", body)
	}
}

// test put services
func TestPutServices(t *testing.T) {
	// good request test
	resp, err := rest("PUT", "/services", fmt.Sprintf("[%v]", testService))
	if err != nil {
		t.Error(err)
	}

	var services []core.Service
	json.Unmarshal(resp, &services)

	if len(services) != 1 {
		t.Errorf("%q doesn't match expected out", services)
	}

	if len(services) == 1 && services[0].Id != "tcp-192_168_0_15-80" {
		t.Errorf("%q doesn't match expected out", services)
	}

	// bad request test
	resp, err = rest("PUT", "/services", testService)
	if err != nil {
		t.Error(err)
	}

	if !strings.Contains(string(resp), "Bad JSON syntax received in body") {
		t.Errorf("%q doesn't match expected out", resp)
	}

	// clear services
	rest("PUT", "/services", "[]")
}

// test post services
func TestPostService(t *testing.T) {
	// good request test
	resp, err := rest("POST", "/services", testService)
	if err != nil {
		t.Error(err)
	}

	var service core.Service
	json.Unmarshal(resp, &service)

	if service.Id != "tcp-192_168_0_15-80" {
		t.Errorf("%q doesn't match expected out", service)
	}

	// bad request test
	resp, err = rest("POST", "/services", badService)
	if err != nil {
		t.Error(err)
	}

	if !strings.Contains(string(resp), "Bad JSON syntax received in body") {
		t.Errorf("%q doesn't match expected out", resp)
	}
}

// test get service
func TestGetService(t *testing.T) {
	// good request test
	resp, err := rest("GET", "/services/tcp-192_168_0_15-80", "")
	if err != nil {
		t.Error(err)
	}

	var service core.Service
	json.Unmarshal(resp, &service)

	if service.Host != "192.168.0.15" {
		t.Errorf("%q doesn't match expected out", service)
	}

	// bad request test
	resp, err = rest("GET", "/services/not-real", "")
	if err != nil {
		t.Error(err)
	}

	if !strings.Contains(string(resp), "No Service Found") {
		t.Errorf("%q doesn't match expected out", resp)
	}
}

// test put services
func TestPutService(t *testing.T) {
	// good request test
	resp, err := rest("PUT", "/services/tcp-192_168_0_15-80", testService2)
	if err != nil {
		t.Error(err)
	}

	var service core.Service
	json.Unmarshal(resp, &service)

	if service.Id != "tcp-192_168_0_16-443" {
		t.Errorf("%q doesn't match expected out", service)
	}

	// verify old service is gone
	resp, err = rest("GET", "/services/tcp-192_168_0_15-80", "")
	if err != nil {
		t.Error(err)
	}

	if !strings.Contains(string(resp), "No Service Found") {
		t.Errorf("%q doesn't match expected out", resp)
	}

	// bad request test
	resp, err = rest("PUT", "/services/not-real", testService2)
	if err != nil {
		t.Error(err)
	}

	if !strings.Contains(string(resp), "No Service Found") {
		t.Errorf("%q doesn't match expected out", resp)
	}
}

// test delete service
func TestDeleteService(t *testing.T) {
	// good request test
	resp, err := rest("DELETE", "/services/tcp-192_168_0_16-443", "")
	if err != nil {
		t.Error(err)
	}

	if !strings.Contains(string(resp), "Success") {
		t.Errorf("%q doesn't match expected out", resp)
	}

	// bad request test
	resp, err = rest("DELETE", "/services/not-real", "")
	if err != nil {
		t.Error(err)
	}

	if !strings.Contains(string(resp), "Success") {
		t.Errorf("%q doesn't match expected out", resp)
	}
}

////////////////////////////////////////////////////////////////////////////////
// SERVERS
////////////////////////////////////////////////////////////////////////////////
// test get servers
func TestGetServers(t *testing.T) {
	rest("POST", "/services", testService)
	resp, err := rest("GET", "/services/tcp-192_168_0_15-80/servers", "")
	if err != nil {
		t.Error(err)
	}
	if string(resp) != "[]\n" {
		t.Errorf("%q doesn't match expected out", resp)
	}
}

// test put servers
func TestPutServers(t *testing.T) {
	resp, err := rest("PUT", "/services/tcp-192_168_0_15-80/servers", fmt.Sprintf("[%v,%v]", testServer1, testServer2))
	if err != nil {
		t.Error(err)
	}

	var servers []core.Server
	json.Unmarshal(resp, &servers)

	if len(servers) != 2 {
		t.Errorf("%q doesn't match expected out", servers)
	}

	if len(servers) > 0 && servers[0].Id != "127_0_0_11-8080" {
		t.Errorf("%q doesn't match expected out", servers)
	}

	// bad request test
	resp, err = rest("PUT", "/services/tcp-192_168_0_15-80/servers", testServer3)
	if err != nil {
		t.Error(err)
	}

	if !strings.Contains(string(resp), "Bad JSON syntax received in body") {
		t.Errorf("%q doesn't match expected out", resp)
	}
}

// test post server
func TestPostServer(t *testing.T) {
	resp, err := rest("POST", "/services/tcp-192_168_0_15-80/servers", testServer3)
	if err != nil {
		t.Error(err)
	}

	var server core.Server
	json.Unmarshal(resp, &server)

	if server.Id != "127_0_0_13-8080" {
		t.Errorf("%q doesn't match expected out", server)
	}

	// bad request test
	resp, err = rest("POST", "/services/tcp-192_168_0_15-80/servers", fmt.Sprintf("[%v]", testServer3))
	if err != nil {
		t.Error(err)
	}

	if !strings.Contains(string(resp), "Bad JSON syntax received in body") {
		t.Errorf("%q doesn't match expected out", resp)
	}
}

// test get server
func TestGetServer(t *testing.T) {
	// good request test
	resp, err := rest("GET", "/services/tcp-192_168_0_15-80/servers/127_0_0_11-8080", "")
	if err != nil {
		t.Error(err)
	}

	var server core.Server
	json.Unmarshal(resp, &server)

	if server.Host != "127.0.0.11" {
		t.Errorf("%q doesn't match expected out", server)
	}

	// bad request test
	resp, err = rest("GET", "/services/tcp-192_168_0_15-80/servers/unreal", "")
	if err != nil {
		t.Error(err)
	}

	if !strings.Contains(string(resp), "No Server Found") {
		t.Errorf("%q doesn't match expected out", resp)
	}
}

// test delete server
func TestDeleteServer(t *testing.T) {
	// good request test
	resp, err := rest("DELETE", "/services/tcp-192_168_0_15-80/servers/127_0_0_11-8080", "")
	if err != nil {
		t.Error(err)
	}

	if !strings.Contains(string(resp), "Success") {
		t.Errorf("%q doesn't match expected out", resp)
	}

	// bad request test
	resp, err = rest("DELETE", "/services/tcp-192_168_0_15-80/servers/unreal", "")
	if err != nil {
		t.Error(err)
	}

	if !strings.Contains(string(resp), "Success") {
		t.Errorf("%q doesn't match expected out", resp)
	}
}

////////////////////////////////////////////////////////////////////////////////
// PRIVS
////////////////////////////////////////////////////////////////////////////////
// hit api and return response body
func rest(method, route, data string) ([]byte, error) {
	body := bytes.NewBuffer([]byte(data))

	req, _ := http.NewRequest(method, fmt.Sprintf("https://%s%s", apiAddr, route), body)
	req.Header.Add("X-NANOBOX-TOKEN", "")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Unable to %v %v - %v", method, route, err)
	}
	defer res.Body.Close()

	b, _ := ioutil.ReadAll(res.Body)

	return b, nil
}

// manually configure and start internals
func initialize() {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	config.DatabaseConnection = "scribble:///tmp/portalTest"
	config.ClusterConnection = "none://"
	config.ApiHost = "127.0.0.1"
	config.ApiPort = "8444"
	config.ApiToken = ""
	config.Log = lumber.NewConsoleLogger(lumber.LvlInt("FATAL"))
	apiAddr = fmt.Sprintf("%v:%v", config.ApiHost, config.ApiPort)

	// initialize database
	err := database.Init()
	if err != nil {
		fmt.Printf("Database init failed - %v\n", err)
		os.Exit(1)
	}
	// initialize clusterer
	err = cluster.Init()
	if err != nil {
		fmt.Printf("Clusterer init failed - %v\n", err)
		os.Exit(1)
	}
	// initialize balancer
	balance.Balancer = &database.ScribbleDatabase{}
	err = balance.Balancer.Init()
	if err != nil {
		fmt.Printf("Balancer init failed - %v\n", err)
		os.Exit(1)
	}
	// load saved rules
	services, err := database.Backend.GetServices()
	if err != nil {
		// if error is not about a missing db, continue
		if !strings.Contains(err.Error(), "Found") {
			// todo: this requires backends to return NoServiceError in GetServices
			fmt.Printf("Get services from backend failed - %v\n", err)
			os.Exit(1)
		}
	}
	// apply saved rules
	err = balance.Balancer.SetServices(services)
	if err != nil {
		fmt.Printf("Balancer sync failed - %v\n", err)
		os.Exit(1)
	}
}
