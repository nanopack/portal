package database_test

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	"github.com/jcelliott/lumber"

	"github.com/nanopack/portal/config"
	"github.com/nanopack/portal/database"
)

var (
	Backend database.Backender

	testService1 = database.Service{Id: "tcp-192_168_0_15-80", Host: "192.168.0.15", Port: 80, Type: "tcp", Scheduler: "wrr"}
	testService2 = database.Service{Id: "tcp-192_168_0_16-80", Host: "192.168.0.16", Port: 80, Type: "tcp", Scheduler: "wrr"}
	testServer1  = database.Server{Id: "127_0_0_11-8080", Host: "127.0.0.11", Port: 8080, Forwarder: "m", Weight: 5, UpperThreshold: 10, LowerThreshold: 1}
	testServer2  = database.Server{Id: "127_0_0_12-8080", Host: "127.0.0.12", Port: 8080, Forwarder: "m", Weight: 5, UpperThreshold: 10, LowerThreshold: 1}
)

func TestMain(m *testing.M) {
	// clean test dir
	os.RemoveAll("/tmp/portalTest")

	config.DatabaseConnection = "scribble:///tmp/portalTest"
	config.Log = lumber.NewConsoleLogger(lumber.LvlInt("FATAL"))

	Backend = &database.ScribbleDatabase{}
	Backend.Init()

	rtn := m.Run()

	// clean test dir
	os.RemoveAll("/tmp/portalTest")

	os.Exit(rtn)
}

func TestSetService(t *testing.T) {
	if err := Backend.SetService(&testService1); err != nil {
		t.Errorf("Failed to SET service - %v", err)
	}

	service, err := ioutil.ReadFile("/tmp/portalTest/services/tcp-192_168_0_15-80.json")
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
	services := []database.Service{}
	services = append(services, testService2)

	if err := Backend.SetServices(services); err != nil {
		t.Errorf("Failed to SET services - %v", err)
	}

	if _, err := os.Stat("/tmp/portalTest/services/tcp-192_168_0_15-80.json"); !os.IsNotExist(err) {
		t.Errorf("Failed to clear old services on PUT - %v", err)
	}

	service, err := ioutil.ReadFile("/tmp/portalTest/services/tcp-192_168_0_16-80.json")
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
	services, err := Backend.GetServices()
	if err != nil {
		t.Errorf("Failed to GET services - %v", err)
	}

	if services[0].Id != testService2.Id {
		t.Errorf("Read service differs from written service")
	}
}

func TestGetService(t *testing.T) {
	service, err := Backend.GetService(testService2.Id)
	if err != nil {
		t.Errorf("Failed to GET service - %v", err)
	}

	if service.Id != testService2.Id {
		t.Errorf("Read service differs from written service")
	}
}

func TestDeleteService(t *testing.T) {
	if err := Backend.DeleteService(testService2.Id); err != nil {
		t.Errorf("Failed to GET service - %v", err)
	}

	if _, err := os.Stat("/tmp/portalTest/services/tcp-192_168_0_16-80.json"); !os.IsNotExist(err) {
		t.Errorf("Failed to DELETE service - %v", err)
	}
}

func TestSetServer(t *testing.T) {
	Backend.SetService(&testService1)
	if err := Backend.SetServer(testService1.Id, &testServer1); err != nil {
		t.Errorf("Failed to SET server - %v", err)
	}

	service, err := ioutil.ReadFile("/tmp/portalTest/services/tcp-192_168_0_15-80.json")
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
	servers := []database.Server{}
	servers = append(servers, testServer2)
	if err := Backend.SetServers(testService1.Id, servers); err != nil {
		t.Errorf("Failed to SET servers - %v", err)
	}

	service, err := ioutil.ReadFile("/tmp/portalTest/services/tcp-192_168_0_15-80.json")
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
	service, err := Backend.GetService(testService1.Id)
	if err != nil {
		t.Errorf("Failed to GET service - %v", err)
	}

	if service.Servers[0].Id != testServer2.Id {
		t.Errorf("Read server differs from written server")
	}
}

func TestGetServer(t *testing.T) {
	server, err := Backend.GetServer(testService1.Id, testServer2.Id)
	if err != nil {
		t.Errorf("Failed to GET server - %v", err)
	}

	if server.Id != testServer2.Id {
		t.Errorf("Read server differs from written server")
	}
}

func TestDeleteServer(t *testing.T) {
	err := Backend.DeleteServer(testService1.Id, testServer2.Id)
	if err != nil {
		t.Errorf("Failed to DELETE server - %v", err)
	}

	service, err := ioutil.ReadFile("/tmp/portalTest/services/tcp-192_168_0_15-80.json")
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

func toJson(v interface{}) ([]byte, error) {
	jsonified, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return nil, err
	}
	return jsonified, nil
}
