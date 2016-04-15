package cluster_test

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/jcelliott/lumber"

	"github.com/nanopack/portal/balance"
	"github.com/nanopack/portal/cluster"
	"github.com/nanopack/portal/config"
	"github.com/nanopack/portal/core"
	"github.com/nanopack/portal/database"
	"github.com/nanopack/portal/proxymgr"
	"github.com/nanopack/portal/vipmgr"
)

var (
	skip = false // skip if redis-server not installed

	testService1 = core.Service{Id: "tcp-192_168_0_15-80", Host: "192.168.0.15", Port: 80, Type: "tcp", Scheduler: "wrr"}
	testService2 = core.Service{Id: "tcp-192_168_0_16-80", Host: "192.168.0.16", Port: 80, Type: "tcp", Scheduler: "wrr"}
	testServer1  = core.Server{Id: "127_0_0_11-8080", Host: "127.0.0.11", Port: 8080, Forwarder: "m", Weight: 5, UpperThreshold: 10, LowerThreshold: 1}
	testServer2  = core.Server{Id: "127_0_0_12-8080", Host: "127.0.0.12", Port: 8080, Forwarder: "m", Weight: 5, UpperThreshold: 10, LowerThreshold: 1}

	key       = "-----BEGIN PRIVATE KEY-----\nMIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQCbZCr7ZOKrTVOd\nXpwatbrGLl8v4SRdJd17ys+Gm/VToe9oF4FMTGq02agtmeiPeaQHpnaXW5cr8oEB\n39q40SVzPEZIkT6n5vbWj+SMpGH1/ppo07H/aNEU0BBwdgvSR+2sU+ypeX04RQuS\n6pdTuXhy4kNoNaKIWyO7OqK9M0Hn8/x2yLyZhKts4gn835IEQiLWSSE0hK/Q0yqX\npx2HxkH14F+rdFDxMn7StEZoyoTaVpGQRpIO6etejg9BoDMXhSB7VgiO5hvLwp43\n36tbWpVadeqkS6YW+V5cs95LE3+oYK6l5Yyz572BqZM0Caitfh//710lbthe6p5+\nSzsNb/Q5AgMBAAECggEATL78O41oJhLa6S6BCvAWfysH+C3KN/crnKheNq1wTQ39\nn/t78KMNUKTvWxZYtgPt75lXmQmzcBElhjd5Xy5swK1USSLzPxnjb7VBu/S0LTrC\nKGPl1a9/FDhu5hxnWkQMLsCEcm9+WPxA6x7R/pfr1VHK2P0keRQKYb5kAe3+7v/c\n7jTMRMmlcY48SBIIObbPClPrpQEhOPIv5Eig0P+1Pmer7HkMVuNtyMropRQ6v5gt\n+nc0ytmwWylZMMbhiF8XHTAKY2xEyUc56zlKjzRCL80iwtaH/Vr4h01zLSwGUH1w\n84oFuwEYyxhm4GZAFwXRX3gf+FD5gV4+mj+4H5wSwQKBgQDMYEaQd/S6EEUbNaHq\n6JDZNSb2Re96mknh7YEyB/oCaID3MsCbuNQMX5uFtDI1mc3vJly17oR1v+et5zhP\nMHl8OZ5wEyArrHcoTE/r8K96jZleeUX9Cz8ujV0ZD/CGoBLL6OlptKt5FHcoga7H\n0ZdE024CHT+DI8PPqpZpu1n0rwKBgQDCpFn5kF5iBkfJBKDciy+i9gWFd6gDhx5I\nnQvwGvAC02BWuPKH6uzmRJYFSvRvfaG1oKqX5xVlAQZksJUMZxqT9j1riGABKXMr\nnnhq8bNyFYDorCaaVfxSt+GB0z/siDYVeZOJlcUIOKviVqH+HMXC9kTfJCTQuF6d\nR+M9pfOvlwKBgBDYlrhtytRTZv7ZKuGMDfR5dx6xoQ3ADfr7crzG/4qXRpoZqtqr\nH39tmgopUkIszVa7GMU+RdjW2qfw+Sk926Wrsi2Wxf4TlzbRI31VN4Gojk3FPUmg\nVbLmoBfiwna2VxZLuoGmDMRMNY43MkryMb/Qla7C7mtG1WsWqpNIiB+tAoGASWoS\nIcZpQxHZW6GqRuUct5uR45CJR6NcMclCanLOmlI94RfrKobaidPOvfpSjgbVyprq\nHVdkw28KiUntPftZk/tpmTib9XQ743TnOHcn1tzzfU8JVGcgP9bpcL1MPBv4QktT\n8a4S3hH6CungOeeCVBHtUjjgxfT0guBNfsAsVMsCgYEAmNVIr1uTRaIAOnSl3H9u\nrCMz2IhsvPHxS2R0VPHiJCjCRld16O8cLjdkf8F1DGVJVbjLgUR8YDmgaGsFrc1d\nKuWr0SEvUEpwWMEhBeBzVrfWUNgfHo4nTP6WmGAj2S4++mk6F44RuPnky1R8Ea/i\nq01TKnEAgdm+zV2a1ydiSpc=\n-----END PRIVATE KEY-----"
	cert      = "-----BEGIN CERTIFICATE-----\nMIIDXTCCAkWgAwIBAgIJAL/FFFuKTjwRMA0GCSqGSIb3DQEBCwUAMEUxCzAJBgNV\nBAYTAlVTMQswCQYDVQQIDAJJRDETMBEGA1UECgwKbmFub2JveC5pbzEUMBIGA1UE\nAwwLcG9ydGFsLnRlc3QwHhcNMTYwMzIzMTQ1NjMzWhcNMTcwMzIzMTQ1NjMzWjBF\nMQswCQYDVQQGEwJVUzELMAkGA1UECAwCSUQxEzARBgNVBAoMCm5hbm9ib3guaW8x\nFDASBgNVBAMMC3BvcnRhbC50ZXN0MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIB\nCgKCAQEAm2Qq+2Tiq01TnV6cGrW6xi5fL+EkXSXde8rPhpv1U6HvaBeBTExqtNmo\nLZnoj3mkB6Z2l1uXK/KBAd/auNElczxGSJE+p+b21o/kjKRh9f6aaNOx/2jRFNAQ\ncHYL0kftrFPsqXl9OEULkuqXU7l4cuJDaDWiiFsjuzqivTNB5/P8dsi8mYSrbOIJ\n/N+SBEIi1kkhNISv0NMql6cdh8ZB9eBfq3RQ8TJ+0rRGaMqE2laRkEaSDunrXo4P\nQaAzF4Uge1YIjuYby8KeN9+rW1qVWnXqpEumFvleXLPeSxN/qGCupeWMs+e9gamT\nNAmorX4f/+9dJW7YXuqefks7DW/0OQIDAQABo1AwTjAdBgNVHQ4EFgQU66LzKbHE\nyE9LCnaqkcEwOeVQ3fgwHwYDVR0jBBgwFoAU66LzKbHEyE9LCnaqkcEwOeVQ3fgw\nDAYDVR0TBAUwAwEB/zANBgkqhkiG9w0BAQsFAAOCAQEAQ2lAzHHyJfyONWfcao6C\nOz5k8Il4eJ3d55qqYvyVBBWp/sFIh9aLGDazbaX7sO55cur/uWp0SiiMw/tt+2nG\n6Yn08l1FeSBDXwvrFOJXScSMEb7Ttl3y2qfJ3z6/rPx6eIBU0c/uzAH+sHiIQNJ1\n7FXD7CvGSIzxU0UU1LEsgM0o5HrOLPubsHmKruM8hcKxHkj9pXKIgY4SJe4BOhwm\nbVh43+VrCDNJf79/KmWrwFXFMg2QvsGS673ps1uGEafGj5vzX4n9S0aCV71ser5P\nmVX2N3jj2WgiYIXI5SmH3BlfR5aGWq4Fq124gi9dxkZljFTolTc6aYyQu0i40B0X\nzQ==\n-----END CERTIFICATE-----"
	testCert  = core.CertBundle{Key: key, Cert: cert}
	testRoute = core.Route{Domain: "portal.test", Page: "routing works\n"}
)

func TestMain(m *testing.M) {
	// clean test dir
	os.RemoveAll("/tmp/clusterTest")

	// initialize backend if redis-server found
	initialize()

	conn, err := redis.DialURL(config.ClusterConnection, redis.DialConnectTimeout(30*time.Second), redis.DialPassword(config.ClusterToken))
	if err != nil {
		return
	}
	hostname, _ := os.Hostname()
	self := fmt.Sprintf("%v:%v", hostname, config.ApiPort)
	defer conn.Do("SREM", "members", self)
	defer conn.Close()

	rtn := m.Run()

	// clean test dir
	os.RemoveAll("/tmp/clusterTest")
	// just in case, ensure clean members
	conn.Do("SREM", "members", self)
	conn.Close()

	os.Exit(rtn)
}

////////////////////////////////////////////////////////////////////////////////
// SERVICES
////////////////////////////////////////////////////////////////////////////////
func TestSetService(t *testing.T) {
	if skip {
		t.SkipNow()
	}
	if err := cluster.SetService(&testService1); err != nil {
		t.Errorf("Failed to SET service - %v", err)
		t.FailNow()
	}

	service, err := cluster.GetService(testService1.Id)
	if err != nil {
		t.Error(err)
	}

	if service.Host != testService1.Host {
		t.Errorf("Read service differs from written service")
	}
}

func TestSetServices(t *testing.T) {
	if skip {
		t.SkipNow()
	}
	services := []core.Service{}
	services = append(services, testService2)

	if err := cluster.SetServices(services); err != nil {
		t.Errorf("Failed to SET services - %v", err)
		t.FailNow()
	}

	if _, err := os.Stat("/tmp/scribbleTest/services/tcp-192_168_0_15-80.json"); !os.IsNotExist(err) {
		t.Errorf("Failed to clear old services on PUT - %v", err)
	}

	service, err := cluster.GetService(testService2.Id)
	if err != nil {
		t.Error(err)
	}

	if service.Host != testService2.Host {
		t.Errorf("Read service differs from written service")
	}
}

func TestGetServices(t *testing.T) {
	if skip {
		t.SkipNow()
	}
	// don't use cluster.GetServices()
	services, err := database.GetServices()
	if err != nil {
		t.Errorf("Failed to GET services - %v", err)
		t.FailNow()
	}

	if services[0].Id != testService2.Id {
		t.Errorf("Read service differs from written service")
	}
}

func TestGetService(t *testing.T) {
	if skip {
		t.SkipNow()
	}
	service, err := cluster.GetService(testService2.Id)
	if err != nil {
		t.Errorf("Failed to GET service - %v", err)
		t.FailNow()
	}

	if service.Id != testService2.Id {
		t.Errorf("Read service differs from written service")
	}
}

func TestDeleteService(t *testing.T) {
	if skip {
		t.SkipNow()
	}
	if err := cluster.DeleteService(testService2.Id); err != nil {
		t.Errorf("Failed to DELETE service - %v", err)
		t.FailNow()
	}

	_, err := cluster.GetService(testService2.Id)
	if !strings.Contains(err.Error(), "No Service Found") {
		t.Error(err)
	}
}

////////////////////////////////////////////////////////////////////////////////
// SERVERS
////////////////////////////////////////////////////////////////////////////////
func TestSetServer(t *testing.T) {
	if skip {
		t.SkipNow()
	}
	cluster.SetService(&testService1)
	if err := cluster.SetServer(testService1.Id, &testServer1); err != nil {
		t.Errorf("Failed to SET server - %v", err)
		t.FailNow()
	}

	service, err := cluster.GetService(testService1.Id)
	if err != nil {
		t.Error(err)
	}

	svc := testService1
	svc.Servers = append(svc.Servers, testServer1)

	if service.Servers[0].Host != svc.Servers[0].Host {
		t.Errorf("Read service differs from written service")
	}
}

func TestSetServers(t *testing.T) {
	if skip {
		t.SkipNow()
	}
	servers := []core.Server{}
	servers = append(servers, testServer2)
	if err := cluster.SetServers(testService1.Id, servers); err != nil {
		t.Errorf("Failed to SET servers - %v", err)
		t.FailNow()
	}

	service, err := cluster.GetService(testService1.Id)
	if err != nil {
		t.Error(err)
	}

	svc := testService1
	svc.Servers = append(svc.Servers, testServer2)

	if service.Servers[0].Host != svc.Servers[0].Host {
		t.Errorf("Failed to clear old servers on PUT")
	}
}

func TestGetServers(t *testing.T) {
	if skip {
		t.SkipNow()
	}
	service, err := cluster.GetService(testService1.Id)
	if err != nil {
		t.Errorf("Failed to GET service - %v", err)
		t.FailNow()
	}

	if service.Host == "" || len(service.Servers) == 0 {
		t.Errorf("GOT empty service")
		t.FailNow()
	}

	if service.Servers[0].Id != testServer2.Id {
		t.Errorf("Read server differs from written server")
	}
}

func TestGetServer(t *testing.T) {
	if skip {
		t.SkipNow()
	}
	server, err := cluster.GetServer(testService1.Id, testServer2.Id)
	if err != nil {
		t.Errorf("Failed to GET server - %v", err)
		t.FailNow()
	}

	if server.Id != testServer2.Id {
		t.Errorf("Read server differs from written server")
	}
}

func TestDeleteServer(t *testing.T) {
	if skip {
		t.SkipNow()
	}
	err := cluster.DeleteServer(testService1.Id, testServer2.Id)
	if err != nil {
		t.Errorf("Failed to DELETE server - %v", err)
	}

	service, err := cluster.GetService(testService1.Id)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if service.Id != testService1.Id {
		t.Errorf("Read service differs from written service")
	}
}

////////////////////////////////////////////////////////////////////////////////
// ROUTES
////////////////////////////////////////////////////////////////////////////////
func TestSetRoute(t *testing.T) {
	if skip {
		t.SkipNow()
	}
	if err := cluster.SetRoute(testRoute); err != nil {
		t.Errorf("Failed to SET route - %v", err)
		t.FailNow()
	}

	routes, err := database.GetRoutes()
	if err != nil {
		t.Error(err)
	}

	if len(routes) != 1 || routes[0].Domain != testRoute.Domain {
		t.Errorf("Read route differs from written route")
	}
}

func TestSetRoutes(t *testing.T) {
	if skip {
		t.SkipNow()
	}
	routes := []core.Route{testRoute}

	if err := cluster.SetRoutes(routes); err != nil {
		t.Errorf("Failed to SET routes - %v", err)
		t.FailNow()
	}

	// don't use cluster.GetRoutes()
	routes, err := database.GetRoutes()
	if err != nil {
		t.Error(err)
	}

	if len(routes) != 1 || routes[0].Domain != testRoute.Domain {
		t.Errorf("Read route differs from written route")
	}
}

func TestGetRoutes(t *testing.T) {
	if skip {
		t.SkipNow()
	}

	// don't use cluster.GetRoutes() // todo:?
	routes, err := cluster.GetRoutes()
	if err != nil {
		t.Errorf("Failed to GET routes - %v", err)
		t.FailNow()
	}

	if len(routes) == 1 && routes[0].Domain != testRoute.Domain {
		t.Errorf("Read route differs from written route")
	}
}

func TestDeleteRoute(t *testing.T) {
	if skip {
		t.SkipNow()
	}
	if err := cluster.DeleteRoute(testRoute); err != nil {
		t.Errorf("Failed to DELETE route - %v", err)
		t.FailNow()
	}

	// don't use cluster.GetRoutes()
	routes, err := database.GetRoutes()
	if len(routes) != 0 {
		t.Error("Failed to DELETE route - %v", err)
	}
}

////////////////////////////////////////////////////////////////////////////////
// CERTS
////////////////////////////////////////////////////////////////////////////////
func TestSetCert(t *testing.T) {
	if skip {
		t.SkipNow()
	}
	if err := cluster.SetCert(testCert); err != nil {
		t.Errorf("Failed to SET cert - %v", err)
		t.FailNow()
	}

	certs, err := database.GetCerts()
	if err != nil {
		t.Error(err)
	}

	if len(certs) != 1 || certs[0].Cert != testCert.Cert {
		t.Errorf("Read cert differs from written cert")
	}
}

func TestSetCerts(t *testing.T) {
	if skip {
		t.SkipNow()
	}
	certs := []core.CertBundle{testCert}

	if err := cluster.SetCerts(certs); err != nil {
		t.Errorf("Failed to SET certs - %v", err)
		t.FailNow()
	}

	// don't use cluster.GetCerts()
	certs, err := database.GetCerts()
	if err != nil {
		t.Error(err)
	}

	if len(certs) != 1 || certs[0].Cert != testCert.Cert {
		t.Errorf("Read cert differs from written cert")
	}
}

func TestGetCerts(t *testing.T) {
	if skip {
		t.SkipNow()
	}

	// don't use cluster.GetCerts() // todo:?
	certs, err := cluster.GetCerts()
	if err != nil {
		t.Errorf("Failed to GET certs - %v", err)
		t.FailNow()
	}

	if len(certs) == 1 && certs[0].Cert != testCert.Cert {
		t.Errorf("Read cert differs from written cert")
	}
}

func TestDeleteCert(t *testing.T) {
	if skip {
		t.SkipNow()
	}
	if err := cluster.DeleteCert(testCert); err != nil {
		t.Errorf("Failed to DELETE cert - %v", err)
		t.FailNow()
	}

	// don't use cluster.GetCerts()
	certs, err := database.GetCerts()
	if len(certs) != 0 {
		t.Error("Failed to DELETE cert - %v", err)
	}
}

////////////////////////////////////////////////////////////////////////////////
// NONE CLUSTERER
////////////////////////////////////////////////////////////////////////////////

func TestNoneSetService(t *testing.T) {
	config.ClusterConnection = "none://"
	cluster.Init()

	if err := cluster.SetService(&testService1); err != nil {
		t.Errorf("Failed to SET service - %v", err)
		t.FailNow()
	}

	service, err := cluster.GetService(testService1.Id)
	if err != nil {
		t.Error(err)
	}

	if service.Host != testService1.Host {
		t.Errorf("Read service differs from written service")
	}
}

func TestNoneSetServices(t *testing.T) {
	services := []core.Service{}
	services = append(services, testService2)

	if err := cluster.SetServices(services); err != nil {
		t.Errorf("Failed to SET services - %v", err)
		t.FailNow()
	}

	if _, err := os.Stat("/tmp/scribbleTest/services/tcp-192_168_0_15-80.json"); !os.IsNotExist(err) {
		t.Errorf("Failed to clear old services on PUT - %v", err)
	}

	service, err := cluster.GetService(testService2.Id)
	if err != nil {
		t.Error(err)
	}

	if service.Host != testService2.Host {
		t.Errorf("Read service differs from written service")
	}
}

func TestNoneGetServices(t *testing.T) {
	// don't use cluster.GetServices()
	services, err := database.GetServices()
	if err != nil {
		t.Errorf("Failed to GET services - %v", err)
		t.FailNow()
	}

	if services[0].Id != testService2.Id {
		t.Errorf("Read service differs from written service")
	}
}

func TestNoneGetService(t *testing.T) {
	service, err := cluster.GetService(testService2.Id)
	if err != nil {
		t.Errorf("Failed to GET service - %v", err)
		t.FailNow()
	}

	if service.Id != testService2.Id {
		t.Errorf("Read service differs from written service")
	}
}

func TestNoneDeleteService(t *testing.T) {
	if err := cluster.DeleteService(testService2.Id); err != nil {
		t.Errorf("Failed to DELETE service - %v", err)
		t.FailNow()
	}

	_, err := cluster.GetService(testService2.Id)
	if !strings.Contains(err.Error(), "No Service Found") {
		t.Error(err)
	}
}

////////////////////////////////////////////////////////////////////////////////
// SERVERS
////////////////////////////////////////////////////////////////////////////////
func TestNoneSetServer(t *testing.T) {
	cluster.SetService(&testService1)
	if err := cluster.SetServer(testService1.Id, &testServer1); err != nil {
		t.Errorf("Failed to SET server - %v", err)
		t.FailNow()
	}

	service, err := cluster.GetService(testService1.Id)
	if err != nil {
		t.Error(err)
	}

	svc := testService1
	svc.Servers = append(svc.Servers, testServer1)

	if service.Servers[0].Host != svc.Servers[0].Host {
		t.Errorf("Read service differs from written service")
	}
}

func TestNoneSetServers(t *testing.T) {
	servers := []core.Server{}
	servers = append(servers, testServer2)
	if err := cluster.SetServers(testService1.Id, servers); err != nil {
		t.Errorf("Failed to SET servers - %v", err)
		t.FailNow()
	}

	service, err := cluster.GetService(testService1.Id)
	if err != nil {
		t.Error(err)
	}

	svc := testService1
	svc.Servers = append(svc.Servers, testServer2)

	if service.Servers[0].Host != svc.Servers[0].Host {
		t.Errorf("Failed to clear old servers on PUT")
	}
}

func TestNoneGetServers(t *testing.T) {
	service, err := cluster.GetService(testService1.Id)
	if err != nil {
		t.Errorf("Failed to GET service - %v", err)
		t.FailNow()
	}

	if service.Host == "" || len(service.Servers) == 0 {
		t.Errorf("GOT empty service")
		t.FailNow()
	}

	if service.Servers[0].Id != testServer2.Id {
		t.Errorf("Read server differs from written server")
	}
}

func TestNoneGetServer(t *testing.T) {
	server, err := cluster.GetServer(testService1.Id, testServer2.Id)
	if err != nil {
		t.Errorf("Failed to GET server - %v", err)
		t.FailNow()
	}

	if server.Id != testServer2.Id {
		t.Errorf("Read server differs from written server")
	}
}

func TestNoneDeleteServer(t *testing.T) {
	err := cluster.DeleteServer(testService1.Id, testServer2.Id)
	if err != nil {
		t.Errorf("Failed to DELETE server - %v", err)
	}

	service, err := cluster.GetService(testService1.Id)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if service.Id != testService1.Id {
		t.Errorf("Read service differs from written service")
	}
}

////////////////////////////////////////////////////////////////////////////////
// ROUTES
////////////////////////////////////////////////////////////////////////////////
func TestNoneSetRoute(t *testing.T) {
	if err := cluster.SetRoute(testRoute); err != nil {
		t.Errorf("Failed to SET route - %v", err)
		t.FailNow()
	}

	routes, err := database.GetRoutes()
	if err != nil {
		t.Error(err)
	}

	if len(routes) != 1 || routes[0].Domain != testRoute.Domain {
		t.Errorf("Read route differs from written route")
	}
}

func TestNoneSetRoutes(t *testing.T) {
	routes := []core.Route{testRoute}

	if err := cluster.SetRoutes(routes); err != nil {
		t.Errorf("Failed to SET routes - %v", err)
		t.FailNow()
	}

	// don't use cluster.GetRoutes()
	routes, err := database.GetRoutes()
	if err != nil {
		t.Error(err)
	}

	if len(routes) != 1 || routes[0].Domain != testRoute.Domain {
		t.Errorf("Read route differs from written route")
	}
}

func TestNoneGetRoutes(t *testing.T) {

	// don't use cluster.GetRoutes() // todo:?
	routes, err := cluster.GetRoutes()
	if err != nil {
		t.Errorf("Failed to GET routes - %v", err)
		t.FailNow()
	}

	if len(routes) == 1 && routes[0].Domain != testRoute.Domain {
		t.Errorf("Read route differs from written route")
	}
}

func TestNoneDeleteRoute(t *testing.T) {
	if err := cluster.DeleteRoute(testRoute); err != nil {
		t.Errorf("Failed to DELETE route - %v", err)
		t.FailNow()
	}

	// don't use cluster.GetRoutes()
	routes, err := database.GetRoutes()
	if len(routes) != 0 {
		t.Error("Failed to DELETE route - %v", err)
	}
}

////////////////////////////////////////////////////////////////////////////////
// CERTS
////////////////////////////////////////////////////////////////////////////////
func TestNoneSetCert(t *testing.T) {
	if err := cluster.SetCert(testCert); err != nil {
		t.Errorf("Failed to SET cert - %v", err)
		t.FailNow()
	}

	certs, err := database.GetCerts()
	if err != nil {
		t.Error(err)
	}

	if len(certs) != 1 || certs[0].Cert != testCert.Cert {
		t.Errorf("Read cert differs from written cert")
	}
}

func TestNoneSetCerts(t *testing.T) {
	certs := []core.CertBundle{testCert}

	if err := cluster.SetCerts(certs); err != nil {
		t.Errorf("Failed to SET certs - %v", err)
		t.FailNow()
	}

	// don't use cluster.GetCerts()
	certs, err := database.GetCerts()
	if err != nil {
		t.Error(err)
	}

	if len(certs) != 1 || certs[0].Cert != testCert.Cert {
		t.Errorf("Read cert differs from written cert")
	}
}

func TestNoneGetCerts(t *testing.T) {

	// don't use cluster.GetCerts() // todo:?
	certs, err := cluster.GetCerts()
	if err != nil {
		t.Errorf("Failed to GET certs - %v", err)
		t.FailNow()
	}

	if len(certs) == 1 && certs[0].Cert != testCert.Cert {
		t.Errorf("Read cert differs from written cert")
	}
}

func TestNoneDeleteCert(t *testing.T) {
	if err := cluster.DeleteCert(testCert); err != nil {
		t.Errorf("Failed to DELETE cert - %v", err)
		t.FailNow()
	}

	// don't use cluster.GetCerts()
	certs, err := database.GetCerts()
	if len(certs) != 0 {
		t.Error("Failed to DELETE cert - %v", err)
	}
}

////////////////////////////////////////////////////////////////////////////////
// PRIVS
////////////////////////////////////////////////////////////////////////////////
func toJson(v interface{}) ([]byte, error) {
	jsonified, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return nil, err
	}
	return jsonified, nil
}

func initialize() {
	rExec, err := exec.Command("redis-server", "-v").CombinedOutput()
	if err != nil {
		fmt.Printf("Failed to run redis-server - %s%v\n", rExec, err.Error())
		skip = true
	}

	config.RouteHttp = "0.0.0.0:9082"
	config.RouteTls = "0.0.0.0:9445"
	config.LogLevel = "FATAL"
	config.Log = lumber.NewConsoleLogger(lumber.LvlInt(config.LogLevel))

	if !skip {
		config.ClusterConnection = "redis://127.0.0.1:6379"
		config.DatabaseConnection = "scribble:///tmp/clusterTest"

		err = database.Init()
		if err != nil {
			fmt.Printf("database init failed - %v\n", err)
			os.Exit(1)
		}

		balance.Balancer = &database.ScribbleDatabase{}
		err = balance.Balancer.Init()
		if err != nil {
			fmt.Printf("balance init failed - %v\n", err)
			os.Exit(1)
		}

		// initialize proxymgr
		err = proxymgr.Init()
		if err != nil {
			fmt.Printf("Proxymgr init failed - %v\n", err)
			os.Exit(1)
		}

		// initialize vipmgr
		err = vipmgr.Init()
		if err != nil {
			fmt.Printf("Vipmgr init failed - %v\n", err)
			os.Exit(1)
		}

		err = cluster.Init()
		if err != nil {
			fmt.Printf("Cluster init failed - %v\n", err)
			os.Exit(1)
		}
	}
}
