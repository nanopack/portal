package balance_test

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/jcelliott/lumber"

	"github.com/nanopack/portal/balance"
	"github.com/nanopack/portal/config"
	"github.com/nanopack/portal/core"
)

var (
	skip    = false // skip if iptables/ipvsadm not installed
	Backend core.Backender

	testService1 = core.Service{Id: "tcp-192_168_0_15-80", Host: "192.168.0.15", Port: 80, Type: "tcp", Scheduler: "wrr"}
	testService2 = core.Service{Id: "tcp-192_168_0_16-80", Host: "192.168.0.16", Port: 80, Type: "tcp", Scheduler: "wrr"}
	testServer1  = core.Server{Id: "127_0_0_11-8080", Host: "127.0.0.11", Port: 8080, Forwarder: "m", Weight: 5, UpperThreshold: 10, LowerThreshold: 1}
	testServer2  = core.Server{Id: "127_0_0_12-8080", Host: "127.0.0.12", Port: 8080, Forwarder: "m", Weight: 5, UpperThreshold: 10, LowerThreshold: 1}
)

func TestMain(m *testing.M) {
	ifIptables, err := exec.Command("iptables", "-S").CombinedOutput()
	if err != nil {
		fmt.Printf("FAIL - %s%v\n", ifIptables, err.Error())
		skip = true
	}
	ifIpvsadm, err := exec.Command("ipvsadm", "--version").CombinedOutput()
	if err != nil {
		fmt.Printf("FAIL - %s%v\n", ifIpvsadm, err.Error())
		skip = true
	}

	config.Log = lumber.NewConsoleLogger(lumber.LvlInt("FATAL"))

	if !skip {
		Backend = &balance.Lvs{}
		Backend.Init()
	}

	rtn := m.Run()

	os.Exit(rtn)
}

func TestSetService(t *testing.T) {
	if skip {
		t.SkipNow()
	}
	if err := Backend.SetService(&testService1); err != nil {
		t.Errorf("Failed to SET service - %v", err)
	}

	// todo: read from ipvsadm
	service, err := Backend.GetService(testService1.Id)
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

	if err := Backend.SetServices(services); err != nil {
		t.Errorf("Failed to SET services - %v", err)
	}

	if _, err := os.Stat("/tmp/scribbleTest/services/tcp-192_168_0_15-80.json"); !os.IsNotExist(err) {
		t.Errorf("Failed to clear old services on PUT - %v", err)
	}

	// todo: read from ipvsadm
	service, err := Backend.GetService(testService2.Id)
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
	services, err := Backend.GetServices()
	if err != nil {
		t.Errorf("Failed to GET services - %v", err)
	}

	if services[0].Id != testService2.Id {
		t.Errorf("Read service differs from written service")
	}
}

func TestGetService(t *testing.T) {
	if skip {
		t.SkipNow()
	}
	service, err := Backend.GetService(testService2.Id)
	if err != nil {
		t.Errorf("Failed to GET service - %v", err)
	}

	if service.Id != testService2.Id {
		t.Errorf("Read service differs from written service")
	}
}

func TestDeleteService(t *testing.T) {
	if skip {
		t.SkipNow()
	}
	if err := Backend.DeleteService(testService2.Id); err != nil {
		t.Errorf("Failed to GET service - %v", err)
	}

	// todo: read from ipvsadm
	_, err := Backend.GetService(testService2.Id)
	if !strings.Contains(err.Error(), "No Service Found") {
		t.Error(err)
	}
}

func TestSetServer(t *testing.T) {
	if skip {
		t.SkipNow()
	}
	Backend.SetService(&testService1)
	if err := Backend.SetServer(testService1.Id, &testServer1); err != nil {
		t.Errorf("Failed to SET server - %v", err)
	}

	// todo: read from ipvsadm
	service, err := Backend.GetService(testService1.Id)
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
	if err := Backend.SetServers(testService1.Id, servers); err != nil {
		t.Errorf("Failed to SET servers - %v", err)
	}

	// todo: read from ipvsadm
	service, err := Backend.GetService(testService1.Id)
	if err != nil {
		t.Error(err)
	}

	svc := testService1
	svc.Servers = append(svc.Servers, testServer1)

	if service.Servers[0].Host != svc.Servers[0].Host {
		t.Errorf("Failed to clear old servers on PUT")
	}
}

func TestGetServers(t *testing.T) {
	if skip {
		t.SkipNow()
	}
	service, err := Backend.GetService(testService1.Id)
	if err != nil {
		t.Errorf("Failed to GET service - %v", err)
	}

	if service.Servers[0].Id != testServer2.Id {
		t.Errorf("Read server differs from written server")
	}
}

func TestGetServer(t *testing.T) {
	if skip {
		t.SkipNow()
	}
	server, err := Backend.GetServer(testService1.Id, testServer2.Id)
	if err != nil {
		t.Errorf("Failed to GET server - %v", err)
	}

	if server.Id != testServer2.Id {
		t.Errorf("Read server differs from written server")
	}
}

func TestDeleteServer(t *testing.T) {
	if skip {
		t.SkipNow()
	}
	err := Backend.DeleteServer(testService1.Id, testServer2.Id)
	if err != nil {
		t.Errorf("Failed to DELETE server - %v", err)
	}

	// todo: read from ipvsadm
	service, err := Backend.GetService(testService1.Id)
	if err != nil {
		t.Error(err)
	}

	if service.Id != testService1.Id {
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
