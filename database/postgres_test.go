package database_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/jcelliott/lumber"

	"github.com/nanopack/portal/config"
	"github.com/nanopack/portal/core"
	"github.com/nanopack/portal/database"
)

var (
	pgskip    = false // skip if postgresql not running
	pgbackend database.Storable
)

// Requires SetService to be run first (initializes database)
func TestSetServicePg(t *testing.T) {
	config.DatabaseConnection = "postgres://postgres@127.0.0.1?sslmode=disable"
	config.Log = lumber.NewConsoleLogger(lumber.LvlInt("FATAL"))

	pgbackend = &database.PostgresDb{}
	err := pgbackend.Init()
	if err != nil {
		fmt.Printf("Failed to connect, skipping - %s\n", err.Error())
		pgskip = true
	}

	if pgskip {
		t.SkipNow()
	}

	if err := pgbackend.SetService(&testService1); err != nil {
		t.Errorf("Failed to SET service - %v", err)
	}

	svc, err := pgbackend.GetService("tcp-192_168_0_15-80")
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

func TestSetServicesPg(t *testing.T) {
	if pgskip {
		t.SkipNow()
	}

	services := []core.Service{}
	services = append(services, testService2)

	if err := pgbackend.SetServices(services); err != nil {
		t.Errorf("Failed to SET services - %v", err)
	}

	if _, err := os.Stat("/tmp/scribbleTest/services/tcp-192_168_0_15-80.json"); !os.IsNotExist(err) {
		t.Errorf("Failed to clear old services on PUT - %v", err)
	}

	svc, err := pgbackend.GetService("tcp-192_168_0_16-80")
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

func TestGetServicesPg(t *testing.T) {
	if pgskip {
		t.SkipNow()
	}

	services, err := pgbackend.GetServices()
	if err != nil {
		t.Errorf("Failed to GET services - %v", err)
	}

	if services[0].Id != testService2.Id {
		t.Errorf("Read service differs from written service")
	}
}

func TestGetServicePg(t *testing.T) {
	if pgskip {
		t.SkipNow()
	}

	service, err := pgbackend.GetService(testService2.Id)
	if err != nil {
		t.Errorf("Failed to GET service - %v", err)
	}

	if service.Id != testService2.Id {
		t.Errorf("Read service differs from written service")
	}
}

func TestDeleteServicePg(t *testing.T) {
	if pgskip {
		t.SkipNow()
	}

	if err := pgbackend.DeleteService(testService2.Id); err != nil {
		t.Errorf("Failed to DELETE service - %v", err)
	}

	_, err := pgbackend.GetService(testService2.Id)
	if err == nil {
		t.Error(err)
	}
}

func TestSetServerPg(t *testing.T) {
	if pgskip {
		t.SkipNow()
	}

	pgbackend.SetService(&testService1)
	if err := pgbackend.SetServer(testService1.Id, &testServer1); err != nil {
		t.Errorf("Failed to SET server - %v", err)
	}

	tSvc, err := pgbackend.GetService("tcp-192_168_0_15-80")
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

func TestSetServersPg(t *testing.T) {
	if pgskip {
		t.SkipNow()
	}

	servers := []core.Server{}
	servers = append(servers, testServer2)
	if err := pgbackend.SetServers(testService1.Id, servers); err != nil {
		t.Errorf("Failed to SET servers - %v", err)
	}

	tSvc, err := pgbackend.GetService("tcp-192_168_0_15-80")
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

func TestGetServersPg(t *testing.T) {
	if pgskip {
		t.SkipNow()
	}

	service, err := pgbackend.GetService(testService1.Id)
	if err != nil {
		t.Errorf("Failed to GET service - %v", err)
	}

	if service.Servers[0].Id != testServer2.Id {
		t.Errorf("Read server differs from written server")
	}
}

func TestGetServerPg(t *testing.T) {
	if pgskip {
		t.SkipNow()
	}

	server, err := pgbackend.GetServer(testService1.Id, testServer2.Id)
	if err != nil {
		t.Errorf("Failed to GET server - %v", err)
	}

	if server.Id != testServer2.Id {
		t.Errorf("Read server differs from written server")
	}
}

func TestDeleteServerPg(t *testing.T) {
	if pgskip {
		t.SkipNow()
	}

	err := pgbackend.DeleteServer(testService1.Id, testServer2.Id)
	if err != nil {
		t.Errorf("Failed to DELETE server - %v", err)
	}

	svc, err := pgbackend.GetService("tcp-192_168_0_15-80")
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

func TestSetRoutePg(t *testing.T) {
	if pgskip {
		t.SkipNow()
	}

	if err := pgbackend.SetRoute(testRoute); err != nil {
		t.Errorf("Failed to SET route - %v", err)
	}

	if err := pgbackend.SetRoute(testRoute); err != nil {
		t.Errorf("Failed to SET route - %v", err)
	}

	routes, err := pgbackend.GetRoutes()
	if err != nil {
		t.Error(err)
	}

	if len(routes) != 1 {
		t.Errorf("Wrong number of routes")
	}
}

func TestSetRoutesPg(t *testing.T) {
	if pgskip {
		t.SkipNow()
	}

	routes := []core.Route{testRoute}

	if err := pgbackend.SetRoutes(routes); err != nil {
		t.Errorf("Failed to SET routes - %v", err)
	}

	routes, err := pgbackend.GetRoutes()
	if err != nil {
		t.Error(err)
	}

	if len(routes) != 1 {
		t.Errorf("Wrong number of routes")
	}
}

func TestDeleteRoutePg(t *testing.T) {
	if pgskip {
		t.SkipNow()
	}

	if err := pgbackend.DeleteRoute(testRoute); err != nil {
		t.Errorf("Failed to DELETE route - %v", err)
	}

	routes, err := pgbackend.GetRoutes()
	if err != nil {
		t.Error(err)
	}

	if len(routes) != 0 {
		t.Errorf("Failed to delete route")
	}
}

func TestSetCertPg(t *testing.T) {
	if pgskip {
		t.SkipNow()
	}

	if err := pgbackend.SetCert(testCert); err != nil {
		t.Errorf("Failed to SET cert - %v", err)
	}

	if err := pgbackend.SetCert(testCert); err != nil {
		t.Errorf("Failed to SET cert - %v", err)
	}

	certs, err := pgbackend.GetCerts()
	if err != nil {
		t.Error(err)
	}

	if len(certs) != 1 {
		t.Errorf("Wrong number of certs")
	}
}

func TestSetCertsPg(t *testing.T) {
	if pgskip {
		t.SkipNow()
	}

	certs := []core.CertBundle{testCert}

	if err := pgbackend.SetCerts(certs); err != nil {
		t.Errorf("Failed to SET certs - %v", err)
	}

	certs, err := pgbackend.GetCerts()
	if err != nil {
		t.Error(err)
	}

	if len(certs) != 1 {
		t.Errorf("Wrong number of certs")
	}
}

func TestDeleteCertPg(t *testing.T) {
	if pgskip {
		t.SkipNow()
	}

	if err := pgbackend.DeleteCert(testCert); err != nil {
		t.Errorf("Failed to DELETE cert - %v", err)
	}

	certs, err := pgbackend.GetCerts()
	if err != nil {
		t.Error(err)
	}

	if len(certs) != 0 {
		t.Errorf("Failed to delete cert")
	}
}
