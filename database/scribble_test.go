package database_test

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	"github.com/jcelliott/lumber"

	"github.com/nanopack/portal/config"
	"github.com/nanopack/portal/core"
	"github.com/nanopack/portal/database"
)

var (
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
	os.RemoveAll("/tmp/scribbleTest")

	rtn := m.Run()

	// clean test dir
	os.RemoveAll("/tmp/scribbleTest")

	os.Exit(rtn)
}

func TestSetService(t *testing.T) {
	config.DatabaseConnection = "scribble:///tmp/scribbleTest"
	config.Log = lumber.NewConsoleLogger(lumber.LvlInt("FATAL"))

	// Backend = &database.ScribbleDatabase{}
	database.Init()

	if err := database.SetService(&testService1); err != nil {
		t.Errorf("Failed to SET service - %v", err)
	}

	service, err := ioutil.ReadFile("/tmp/scribbleTest/services/tcp-192_168_0_15-80.json")
	if err != nil {
		t.Error(err)
	}

	jService, err := toJson(testService1)
	if err != nil {
		t.Error(err)
	}

	if string(service) != string(jService) {
		t.Errorf("Read service differs from written service")
	}
}

func TestSetServices(t *testing.T) {
	services := []core.Service{}
	services = append(services, testService2)

	if err := database.SetServices(services); err != nil {
		t.Errorf("Failed to SET services - %v", err)
	}

	if _, err := os.Stat("/tmp/scribbleTest/services/tcp-192_168_0_15-80.json"); !os.IsNotExist(err) {
		t.Errorf("Failed to clear old services on PUT - %v", err)
	}

	service, err := ioutil.ReadFile("/tmp/scribbleTest/services/tcp-192_168_0_16-80.json")
	if err != nil {
		t.Error(err)
	}

	jService, err := toJson(testService2)
	if err != nil {
		t.Error(err)
	}

	if string(service) != string(jService) {
		t.Errorf("Read service differs from written service")
	}
}

func TestGetServices(t *testing.T) {
	services, err := database.GetServices()
	if err != nil {
		t.Errorf("Failed to GET services - %v", err)
	}

	if services[0].Id != testService2.Id {
		t.Errorf("Read service differs from written service")
	}
}

func TestGetService(t *testing.T) {
	service, err := database.GetService(testService2.Id)
	if err != nil {
		t.Errorf("Failed to GET service - %v", err)
	}

	if service.Id != testService2.Id {
		t.Errorf("Read service differs from written service")
	}
}

func TestDeleteService(t *testing.T) {
	if err := database.DeleteService(testService2.Id); err != nil {
		t.Errorf("Failed to DELETE service - %v", err)
	}

	if _, err := os.Stat("/tmp/scribbleTest/services/tcp-192_168_0_16-80.json"); !os.IsNotExist(err) {
		t.Errorf("Failed to DELETE service - %v", err)
	}
}

func TestSetServer(t *testing.T) {
	database.SetService(&testService1)
	if err := database.SetServer(testService1.Id, &testServer1); err != nil {
		t.Errorf("Failed to SET server - %v", err)
	}

	service, err := ioutil.ReadFile("/tmp/scribbleTest/services/tcp-192_168_0_15-80.json")
	if err != nil {
		t.Error(err)
	}

	svc := testService1
	svc.Servers = append(svc.Servers, testServer1)
	jService, err := toJson(svc)
	if err != nil {
		t.Error(err)
	}

	if string(service) != string(jService) {
		t.Errorf("Read service differs from written service")
	}
}

func TestSetServers(t *testing.T) {
	servers := []core.Server{}
	servers = append(servers, testServer2)
	if err := database.SetServers(testService1.Id, servers); err != nil {
		t.Errorf("Failed to SET servers - %v", err)
	}

	service, err := ioutil.ReadFile("/tmp/scribbleTest/services/tcp-192_168_0_15-80.json")
	if err != nil {
		t.Error(err)
	}

	svc := testService1
	svc.Servers = append(svc.Servers, testServer1)
	jService, err := toJson(svc)
	if err != nil {
		t.Error(err)
	}

	if string(service) == string(jService) {
		t.Errorf("Failed to clear old servers on PUT")
	}
}

func TestGetServers(t *testing.T) {
	service, err := database.GetService(testService1.Id)
	if err != nil {
		t.Errorf("Failed to GET service - %v", err)
	}

	if service.Servers[0].Id != testServer2.Id {
		t.Errorf("Read server differs from written server")
	}
}

func TestGetServer(t *testing.T) {
	server, err := database.GetServer(testService1.Id, testServer2.Id)
	if err != nil {
		t.Errorf("Failed to GET server - %v", err)
	}

	if server.Id != testServer2.Id {
		t.Errorf("Read server differs from written server")
	}
}

func TestDeleteServer(t *testing.T) {
	err := database.DeleteServer(testService1.Id, testServer2.Id)
	if err != nil {
		t.Errorf("Failed to DELETE server - %v", err)
	}

	service, err := ioutil.ReadFile("/tmp/scribbleTest/services/tcp-192_168_0_15-80.json")
	if err != nil {
		t.Error(err)
	}

	jService, err := toJson(testService1)
	if err != nil {
		t.Error(err)
	}

	if string(service) != string(jService) {
		t.Errorf("Read service differs from written service")
	}
}

func TestSetRoute(t *testing.T) {
	if err := database.SetRoute(testRoute); err != nil {
		t.Errorf("Failed to SET route - %v", err)
	}

	if err := database.SetRoute(testRoute); err != nil {
		t.Errorf("Failed to SET route - %v", err)
	}

	routes, err := database.GetRoutes()
	if err != nil {
		t.Error(err)
	}

	if len(routes) != 1 {
		t.Errorf("Wrong number of routes")
	}
}

func TestSetRoutes(t *testing.T) {
	routes := []core.Route{testRoute}

	if err := database.SetRoutes(routes); err != nil {
		t.Errorf("Failed to SET routes - %v", err)
	}

	routes, err := database.GetRoutes()
	if err != nil {
		t.Error(err)
	}

	if len(routes) != 1 {
		t.Errorf("Wrong number of routes")
	}
}

func TestDeleteRoute(t *testing.T) {
	if err := database.DeleteRoute(testRoute); err != nil {
		t.Errorf("Failed to DELETE route - %v", err)
	}

	routes, err := database.GetRoutes()
	if err != nil {
		t.Error(err)
	}

	if len(routes) != 0 {
		t.Errorf("Failed to delete route")
	}
}

func TestSetCert(t *testing.T) {
	if err := database.SetCert(testCert); err != nil {
		t.Errorf("Failed to SET cert - %v", err)
	}

	if err := database.SetCert(testCert); err != nil {
		t.Errorf("Failed to SET cert - %v", err)
	}

	certs, err := database.GetCerts()
	if err != nil {
		t.Error(err)
	}

	if len(certs) != 1 {
		t.Errorf("Wrong number of certs")
	}
}

func TestSetCerts(t *testing.T) {
	certs := []core.CertBundle{testCert}

	if err := database.SetCerts(certs); err != nil {
		t.Errorf("Failed to SET certs - %v", err)
	}

	certs, err := database.GetCerts()
	if err != nil {
		t.Error(err)
	}

	if len(certs) != 1 {
		t.Errorf("Wrong number of certs")
	}
}

func TestDeleteCert(t *testing.T) {
	if err := database.DeleteCert(testCert); err != nil {
		t.Errorf("Failed to DELETE cert - %v", err)
	}

	certs, err := database.GetCerts()
	if err != nil {
		t.Error(err)
	}

	if len(certs) != 0 {
		t.Errorf("Failed to delete cert")
	}
}

func toJson(v interface{}) ([]byte, error) {
	jsonified, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return nil, err
	}
	return jsonified, nil
}
