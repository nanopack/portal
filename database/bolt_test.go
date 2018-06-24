package database_test

import (
	"testing"

	"github.com/jcelliott/lumber"

	"github.com/nanopack/portal/config"
	"github.com/nanopack/portal/core"
	"github.com/nanopack/portal/database"
)

var (
	boltbackend database.Storable
)

// Requires Init to be run first (initializes database)
func TestSetServiceBolt(t *testing.T) {
	config.DatabaseConnection = "bolt:///tmp/boltTest.bolt"
	config.Log = lumber.NewConsoleLogger(lumber.LvlInt("FATAL"))

	boltbackend = &database.BoltDb{}
	boltbackend.Init()

	if err := boltbackend.SetService(&testService1); err != nil {
		t.Errorf("Failed to SET service - %s", err)
	}

	svc, err := boltbackend.GetService("tcp-192_168_0_15-80")
	if err != nil {
		t.Error(err)
	}

	service, err := toJson(svc)
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

func TestSetServicesBolt(t *testing.T) {
	services := []core.Service{}
	services = append(services, testService2)

	if err := boltbackend.SetServices(services); err != nil {
		t.Errorf("Failed to SET services - %s", err)
	}

	if _, err := boltbackend.GetService("tcp-192_168_0_15-80"); err == nil {
		t.Errorf("Failed to clear old services on PUT")
	}

	svc, err := boltbackend.GetService("tcp-192_168_0_16-80")
	if err != nil {
		t.Error(err)
	}

	service, err := toJson(svc)
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

func TestGetServicesBolt(t *testing.T) {
	services, err := boltbackend.GetServices()
	if err != nil {
		t.Errorf("Failed to GET services - %s", err)
	}

	if services[0].Id != testService2.Id {
		t.Errorf("Read service differs from written service")
	}
}

func TestGetServiceBolt(t *testing.T) {
	service, err := boltbackend.GetService(testService2.Id)
	if err != nil {
		t.Errorf("Failed to GET service - %s", err)
	}

	if service.Id != testService2.Id {
		t.Errorf("Read service differs from written service")
	}
}

func TestDeleteServiceBolt(t *testing.T) {
	if err := boltbackend.DeleteService(testService2.Id); err != nil {
		t.Errorf("Failed to DELETE service - %s", err)
	}

	_, err := boltbackend.GetService(testService2.Id)
	if err == nil {
		t.Error(err)
	}
}

func TestSetServerBolt(t *testing.T) {
	boltbackend.SetService(&testService1)
	if err := boltbackend.SetServer(testService1.Id, &testServer1); err != nil {
		t.Errorf("Failed to SET server - %s", err)
	}

	tSvc, err := boltbackend.GetService("tcp-192_168_0_15-80")
	if err != nil {
		t.Error(err)
	}

	service, err := toJson(&tSvc)
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

func TestSetServersBolt(t *testing.T) {
	servers := []core.Server{}
	servers = append(servers, testServer2)
	if err := boltbackend.SetServers(testService1.Id, servers); err != nil {
		t.Errorf("Failed to SET servers - %s", err)
	}

	tSvc, err := boltbackend.GetService("tcp-192_168_0_15-80")
	if err != nil {
		t.Error(err)
	}

	service, err := toJson(&tSvc)
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

func TestGetServersBolt(t *testing.T) {
	service, err := boltbackend.GetService(testService1.Id)
	if err != nil {
		t.Errorf("Failed to GET service - %s", err)
	}

	if service.Servers[0].Id != testServer2.Id {
		t.Errorf("Read server differs from written server")
	}
}

func TestGetServerBolt(t *testing.T) {
	server, err := boltbackend.GetServer(testService1.Id, testServer2.Id)
	if err != nil {
		t.Errorf("Failed to GET server - %s", err)
	}

	if server.Id != testServer2.Id {
		t.Errorf("Read server differs from written server")
	}
}

func TestDeleteServerBolt(t *testing.T) {
	err := boltbackend.DeleteServer(testService1.Id, testServer2.Id)
	if err != nil {
		t.Errorf("Failed to DELETE server - %s", err)
	}

	svc, err := boltbackend.GetService("tcp-192_168_0_15-80")
	if err != nil {
		t.Error(err)
	}

	service, err := toJson(svc)
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

func TestSetRouteBolt(t *testing.T) {
	if err := boltbackend.SetRoute(testRoute); err != nil {
		t.Errorf("Failed to SET route - %s", err)
	}

	if err := boltbackend.SetRoute(testRoute); err != nil {
		t.Errorf("Failed to SET route - %s", err)
	}

	routes, err := boltbackend.GetRoutes()
	if err != nil {
		t.Error(err)
	}

	if len(routes) != 1 {
		t.Errorf("Wrong number of routes")
	}
}

func TestSetRoutesBolt(t *testing.T) {
	routes := []core.Route{testRoute}

	if err := boltbackend.SetRoutes(routes); err != nil {
		t.Errorf("Failed to SET routes - %s", err)
	}

	routes, err := boltbackend.GetRoutes()
	if err != nil {
		t.Error(err)
	}

	if len(routes) != 1 {
		t.Errorf("Wrong number of routes")
	}
}

func TestDeleteRouteBolt(t *testing.T) {
	if err := boltbackend.DeleteRoute(testRoute); err != nil {
		t.Errorf("Failed to DELETE route - %s", err)
	}

	routes, err := boltbackend.GetRoutes()
	if err != nil {
		t.Error(err)
	}

	if len(routes) != 0 {
		t.Errorf("Failed to delete route")
	}
}

func TestSetCertBolt(t *testing.T) {
	if err := boltbackend.SetCert(testCert); err != nil {
		t.Errorf("Failed to SET cert - %s", err)
	}

	if err := boltbackend.SetCert(testCert); err != nil {
		t.Errorf("Failed to SET cert - %s", err)
	}

	certs, err := boltbackend.GetCerts()
	if err != nil {
		t.Error(err)
	}

	if len(certs) != 1 {
		t.Errorf("Wrong number of certs")
	}
}

func TestSetCertsBolt(t *testing.T) {
	certs := []core.CertBundle{testCert}

	if err := boltbackend.SetCerts(certs); err != nil {
		t.Errorf("Failed to SET certs - %s", err)
	}

	certs, err := boltbackend.GetCerts()
	if err != nil {
		t.Error(err)
	}

	if len(certs) != 1 {
		t.Errorf("Wrong number of certs")
	}
}

func TestDeleteCertBolt(t *testing.T) {
	if err := boltbackend.DeleteCert(testCert); err != nil {
		t.Errorf("Failed to DELETE cert - %s", err)
	}

	certs, err := boltbackend.GetCerts()
	if err != nil {
		t.Error(err)
	}

	if len(certs) != 0 {
		t.Errorf("Failed to delete cert")
	}
}
