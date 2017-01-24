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
	"github.com/nanopack/portal/proxymgr"
	"github.com/nanopack/portal/vipmgr"
)

var (
	apiAddr string

	testService  = `{ "host": "192.168.0.15", "port": 80, "type": "tcp", "scheduler": "wrr" }`
	testService2 = `{ "host": "192.168.0.16", "port": 443, "type": "tcp", "scheduler": "wrr" }`
	badService   = `{ "port": "80", "type": "tcp", "scheduler": "wrr" }`
	testServer1  = `{ "host": "127.0.0.11", "port": 8080, "forwarder": "m", "weight": 5 }`
	testServer2  = `{ "host": "127.0.0.12", "port": 8080, "forwarder": "m", "weight": 5 }`
	testServer3  = `{ "host": "127.0.0.13", "port": 8080, "forwarder": "m", "weight": 5 }`

	testCert  = `{"key": "-----BEGIN PRIVATE KEY-----\nMIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQCbZCr7ZOKrTVOd\nXpwatbrGLl8v4SRdJd17ys+Gm/VToe9oF4FMTGq02agtmeiPeaQHpnaXW5cr8oEB\n39q40SVzPEZIkT6n5vbWj+SMpGH1/ppo07H/aNEU0BBwdgvSR+2sU+ypeX04RQuS\n6pdTuXhy4kNoNaKIWyO7OqK9M0Hn8/x2yLyZhKts4gn835IEQiLWSSE0hK/Q0yqX\npx2HxkH14F+rdFDxMn7StEZoyoTaVpGQRpIO6etejg9BoDMXhSB7VgiO5hvLwp43\n36tbWpVadeqkS6YW+V5cs95LE3+oYK6l5Yyz572BqZM0Caitfh//710lbthe6p5+\nSzsNb/Q5AgMBAAECggEATL78O41oJhLa6S6BCvAWfysH+C3KN/crnKheNq1wTQ39\nn/t78KMNUKTvWxZYtgPt75lXmQmzcBElhjd5Xy5swK1USSLzPxnjb7VBu/S0LTrC\nKGPl1a9/FDhu5hxnWkQMLsCEcm9+WPxA6x7R/pfr1VHK2P0keRQKYb5kAe3+7v/c\n7jTMRMmlcY48SBIIObbPClPrpQEhOPIv5Eig0P+1Pmer7HkMVuNtyMropRQ6v5gt\n+nc0ytmwWylZMMbhiF8XHTAKY2xEyUc56zlKjzRCL80iwtaH/Vr4h01zLSwGUH1w\n84oFuwEYyxhm4GZAFwXRX3gf+FD5gV4+mj+4H5wSwQKBgQDMYEaQd/S6EEUbNaHq\n6JDZNSb2Re96mknh7YEyB/oCaID3MsCbuNQMX5uFtDI1mc3vJly17oR1v+et5zhP\nMHl8OZ5wEyArrHcoTE/r8K96jZleeUX9Cz8ujV0ZD/CGoBLL6OlptKt5FHcoga7H\n0ZdE024CHT+DI8PPqpZpu1n0rwKBgQDCpFn5kF5iBkfJBKDciy+i9gWFd6gDhx5I\nnQvwGvAC02BWuPKH6uzmRJYFSvRvfaG1oKqX5xVlAQZksJUMZxqT9j1riGABKXMr\nnnhq8bNyFYDorCaaVfxSt+GB0z/siDYVeZOJlcUIOKviVqH+HMXC9kTfJCTQuF6d\nR+M9pfOvlwKBgBDYlrhtytRTZv7ZKuGMDfR5dx6xoQ3ADfr7crzG/4qXRpoZqtqr\nH39tmgopUkIszVa7GMU+RdjW2qfw+Sk926Wrsi2Wxf4TlzbRI31VN4Gojk3FPUmg\nVbLmoBfiwna2VxZLuoGmDMRMNY43MkryMb/Qla7C7mtG1WsWqpNIiB+tAoGASWoS\nIcZpQxHZW6GqRuUct5uR45CJR6NcMclCanLOmlI94RfrKobaidPOvfpSjgbVyprq\nHVdkw28KiUntPftZk/tpmTib9XQ743TnOHcn1tzzfU8JVGcgP9bpcL1MPBv4QktT\n8a4S3hH6CungOeeCVBHtUjjgxfT0guBNfsAsVMsCgYEAmNVIr1uTRaIAOnSl3H9u\nrCMz2IhsvPHxS2R0VPHiJCjCRld16O8cLjdkf8F1DGVJVbjLgUR8YDmgaGsFrc1d\nKuWr0SEvUEpwWMEhBeBzVrfWUNgfHo4nTP6WmGAj2S4++mk6F44RuPnky1R8Ea/i\nq01TKnEAgdm+zV2a1ydiSpc=\n-----END PRIVATE KEY-----", "cert": "-----BEGIN CERTIFICATE-----\nMIIDXTCCAkWgAwIBAgIJAL/FFFuKTjwRMA0GCSqGSIb3DQEBCwUAMEUxCzAJBgNV\nBAYTAlVTMQswCQYDVQQIDAJJRDETMBEGA1UECgwKbmFub2JveC5pbzEUMBIGA1UE\nAwwLcG9ydGFsLnRlc3QwHhcNMTYwMzIzMTQ1NjMzWhcNMTcwMzIzMTQ1NjMzWjBF\nMQswCQYDVQQGEwJVUzELMAkGA1UECAwCSUQxEzARBgNVBAoMCm5hbm9ib3guaW8x\nFDASBgNVBAMMC3BvcnRhbC50ZXN0MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIB\nCgKCAQEAm2Qq+2Tiq01TnV6cGrW6xi5fL+EkXSXde8rPhpv1U6HvaBeBTExqtNmo\nLZnoj3mkB6Z2l1uXK/KBAd/auNElczxGSJE+p+b21o/kjKRh9f6aaNOx/2jRFNAQ\ncHYL0kftrFPsqXl9OEULkuqXU7l4cuJDaDWiiFsjuzqivTNB5/P8dsi8mYSrbOIJ\n/N+SBEIi1kkhNISv0NMql6cdh8ZB9eBfq3RQ8TJ+0rRGaMqE2laRkEaSDunrXo4P\nQaAzF4Uge1YIjuYby8KeN9+rW1qVWnXqpEumFvleXLPeSxN/qGCupeWMs+e9gamT\nNAmorX4f/+9dJW7YXuqefks7DW/0OQIDAQABo1AwTjAdBgNVHQ4EFgQU66LzKbHE\nyE9LCnaqkcEwOeVQ3fgwHwYDVR0jBBgwFoAU66LzKbHEyE9LCnaqkcEwOeVQ3fgw\nDAYDVR0TBAUwAwEB/zANBgkqhkiG9w0BAQsFAAOCAQEAQ2lAzHHyJfyONWfcao6C\nOz5k8Il4eJ3d55qqYvyVBBWp/sFIh9aLGDazbaX7sO55cur/uWp0SiiMw/tt+2nG\n6Yn08l1FeSBDXwvrFOJXScSMEb7Ttl3y2qfJ3z6/rPx6eIBU0c/uzAH+sHiIQNJ1\n7FXD7CvGSIzxU0UU1LEsgM0o5HrOLPubsHmKruM8hcKxHkj9pXKIgY4SJe4BOhwm\nbVh43+VrCDNJf79/KmWrwFXFMg2QvsGS673ps1uGEafGj5vzX4n9S0aCV71ser5P\nmVX2N3jj2WgiYIXI5SmH3BlfR5aGWq4Fq124gi9dxkZljFTolTc6aYyQu0i40B0X\nzQ==\n-----END CERTIFICATE-----"}`
	testRoute = `{"domain": "portal.test", "page": "routing works\n"}`
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
	resp, err = rest("DELETE", "/services/not-real-service", "")
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
// ROUTES
////////////////////////////////////////////////////////////////////////////////
// test get routes
func TestGetRoutes(t *testing.T) {
	body, err := rest("GET", "/routes", "")
	if err != nil {
		t.Error(err)
	}
	if string(body) != "[]\n" {
		t.Errorf("%q doesn't match expected out", body)
	}
}

// test put routes
func TestPutRoutes(t *testing.T) {
	// good request test
	resp, err := rest("PUT", "/routes", fmt.Sprintf("[%v]", testRoute))
	if err != nil {
		t.Error(err)
	}

	var routes []core.Route
	json.Unmarshal(resp, &routes)

	if len(routes) != 1 {
		t.Errorf("%q doesn't match expected out", routes)
	}

	if len(routes) == 1 && routes[0].Domain != "portal.test" {
		t.Errorf("%q doesn't match expected out", routes)
	}

	// bad request test
	resp, err = rest("PUT", "/routes", testRoute)
	if err != nil {
		t.Error(err)
	}

	if !strings.Contains(string(resp), "Bad JSON syntax received in body") {
		t.Errorf("%q doesn't match expected out", resp)
	}

	// clear routes
	rest("PUT", "/routes", "[]")
}

// test post routes
func TestPostRoute(t *testing.T) {
	// good request test
	resp, err := rest("POST", "/routes", testRoute)
	if err != nil {
		t.Error(err)
	}

	var route core.Route
	json.Unmarshal(resp, &route)

	if route.Domain != "portal.test" {
		t.Errorf("%q doesn't match expected out", route)
	}

	// bad request test
	resp, err = rest("POST", "/routes", "[{\"domains\":\"test.comma\", \"page\": 1}]")
	if err != nil {
		t.Error(err)
	}

	if !strings.Contains(string(resp), "Bad JSON syntax received in body") {
		t.Errorf("%q doesn't match expected out", resp)
	}
}

// test delete route
func TestDeleteRoute(t *testing.T) {
	// good request test
	resp, err := rest("DELETE", "/routes", testRoute)
	if err != nil {
		t.Error(err)
	}

	if !strings.Contains(string(resp), "Success") {
		t.Errorf("%q doesn't match expected out", resp)
	}

	// bad request test
	resp, err = rest("DELETE", "/routes", "{}")
	if err != nil {
		t.Error(err)
	}

	if !strings.Contains(string(resp), "Success") {
		t.Errorf("%q doesn't match expected out", resp)
	}
}

////////////////////////////////////////////////////////////////////////////////
// CERTS
////////////////////////////////////////////////////////////////////////////////
// test get certs
func TestGetCerts(t *testing.T) {
	body, err := rest("GET", "/certs", "")
	if err != nil {
		t.Error(err)
	}
	if string(body) != "[]\n" {
		t.Errorf("%q doesn't match expected out", body)
	}
}

// test put certs
func TestPutCerts(t *testing.T) {
	// good request test
	resp, err := rest("PUT", "/certs", fmt.Sprintf("[%v]", testCert))
	if err != nil {
		t.Error(err)
	}

	var certs []core.CertBundle
	json.Unmarshal(resp, &certs)

	if len(certs) != 1 {
		t.Errorf("%q doesn't match expected out", certs)
	}

	// bad request test
	resp, err = rest("PUT", "/certs", testCert)
	if err != nil {
		t.Error(err)
	}

	if !strings.Contains(string(resp), "Bad JSON syntax received in body") {
		t.Errorf("%q doesn't match expected out", resp)
	}

	// clear certs
	rest("PUT", "/certs", "[]")
}

// test post certs
func TestPostCert(t *testing.T) {
	// good request test
	resp, err := rest("POST", "/certs", testCert)
	if err != nil {
		t.Error(err)
	}

	var cert core.CertBundle
	err = json.Unmarshal(resp, &cert)
	if err != nil {
		t.Errorf("Failed to POST cert - %s", err)
	}
	// bad request test
	resp, err = rest("POST", "/certs", "[{\"key\":\"test.comma\", \"cert\": 1}]")
	if err != nil {
		t.Error(err)
	}

	if !strings.Contains(string(resp), "Bad JSON syntax received in body") {
		t.Errorf("%q doesn't match expected out", resp)
	}
}

// test delete cert
func TestDeleteCert(t *testing.T) {
	// good request test
	resp, err := rest("DELETE", "/certs", testCert)
	if err != nil {
		t.Error(err)
	}

	if !strings.Contains(string(resp), "Success") {
		t.Errorf("%q doesn't match expected out", resp)
	}

	// bad request test
	resp, err = rest("DELETE", "/certs", "{}")
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
	req.Header.Add("X-AUTH-TOKEN", "")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Unable to %s %s - %s", method, route, err)
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
	config.RouteHttp = "0.0.0.0:9080"
	config.RouteTls = "0.0.0.0:9443"
	config.Log = lumber.NewConsoleLogger(lumber.LvlInt("FATAL"))
	config.LogLevel = "FATAL"
	apiAddr = fmt.Sprintf("%s:%s", config.ApiHost, config.ApiPort)

	// initialize database
	err := database.Init()
	if err != nil {
		fmt.Printf("Database init failed - %s\n", err)
		os.Exit(1)
	}
	// initialize balancer
	balance.Balancer = &database.ScribbleDatabase{}
	err = balance.Balancer.Init()
	if err != nil {
		fmt.Printf("Balancer init failed - %s\n", err)
		os.Exit(1)
	}
	// initialize proxymgr
	err = proxymgr.Init()
	if err != nil {
		fmt.Printf("Proxymgr init failed - %s\n", err)
		os.Exit(1)
	}
	// initialize vipmgr
	err = vipmgr.Init()
	if err != nil {
		fmt.Printf("Vipmgr init failed - %s\n", err)
		os.Exit(1)
	}
	// initialize clusterer
	err = cluster.Init()
	if err != nil {
		fmt.Printf("Clusterer init failed - %s\n", err)
		os.Exit(1)
	}
	// load saved rules
	services, err := database.Backend.GetServices()
	if err != nil {
		// if error is not about a missing db, continue
		if !strings.Contains(err.Error(), "Found") {
			// todo: this requires backends to return NoServiceError in GetServices
			fmt.Printf("Get services from backend failed - %s\n", err)
			os.Exit(1)
		}
	}
	// apply saved rules
	err = balance.Balancer.SetServices(services)
	if err != nil {
		fmt.Printf("Balancer sync failed - %s\n", err)
		os.Exit(1)
	}
}
