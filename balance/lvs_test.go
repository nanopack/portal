// +build linux

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
	// initialize backend if ipvsadm/iptables found
	initialize()

	os.Exit(m.Run())
}

////////////////////////////////////////////////////////////////////////////////
// SERVICES
////////////////////////////////////////////////////////////////////////////////
func TestSetService(t *testing.T) {
	if skip {
		t.SkipNow()
	}
	if err := balance.SetService(&testService1); err != nil {
		t.Errorf("Failed to SET service - %v", err)
		t.FailNow()
	}

	// todo: read from ipvsadm
	service, err := balance.GetService(testService1.Id)
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

	if err := balance.SetServices(services); err != nil {
		t.Errorf("Failed to SET services - %v", err)
		t.FailNow()
	}

	if _, err := os.Stat("/tmp/scribbleTest/services/tcp-192_168_0_15-80.json"); !os.IsNotExist(err) {
		t.Errorf("Failed to clear old services on PUT - %v", err)
	}

	// todo: read from ipvsadm
	service, err := balance.GetService(testService2.Id)
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
	services, err := balance.GetServices()
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
	service, err := balance.GetService(testService2.Id)
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
	if err := balance.DeleteService(testService2.Id); err != nil {
		t.Errorf("Failed to GET service - %v", err)
	}

	// todo: read from ipvsadm
	_, err := balance.GetService(testService2.Id)
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
	balance.SetService(&testService1)
	if err := balance.SetServer(testService1.Id, &testServer1); err != nil {
		t.Errorf("Failed to SET server - %v", err)
		t.FailNow()
	}

	// todo: read from ipvsadm
	service, err := balance.GetService(testService1.Id)
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
	if err := balance.SetServers(testService1.Id, servers); err != nil {
		t.Errorf("Failed to SET servers - %v", err)
		t.FailNow()
	}

	// todo: read from ipvsadm
	service, err := balance.GetService(testService1.Id)
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
	service, err := balance.GetService(testService1.Id)
	if err != nil {
		t.Errorf("Failed to GET service - %v", err)
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
	server, err := balance.GetServer(testService1.Id, testServer2.Id)
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
	err := balance.DeleteServer(testService1.Id, testServer2.Id)
	if err != nil {
		t.Errorf("Failed to DELETE server - %v", err)
	}

	// todo: read from ipvsadm
	service, err := balance.GetService(testService1.Id)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if service.Id != testService1.Id {
		t.Errorf("Read service differs from written service")
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
	ifIptables, err := exec.Command("iptables", "-S").CombinedOutput()
	if err != nil {
		fmt.Printf("Failed to run iptables - %s%v\n", ifIptables, err.Error())
		skip = true
	}
	ifIpvsadm, err := exec.Command("ipvsadm", "--version").CombinedOutput()
	if err != nil {
		fmt.Printf("Failed to run ipvsadm - %s%v\n", ifIpvsadm, err.Error())
		skip = true
	}

	config.Log = lumber.NewConsoleLogger(lumber.LvlInt("FATAL"))

	if !skip {
		// todo: find more friendly way to clear crufty rules only
		err = exec.Command("iptables", "-F").Run()
		if err != nil {
			fmt.Printf("Failed to clear iptables - %v\n", err.Error())
			os.Exit(1)
		}
		err = exec.Command("ipvsadm", "-C").Run()
		if err != nil {
			fmt.Printf("Failed to clear ipvsadm - %v\n", err.Error())
			os.Exit(1)
		}

		balance.Init()
	}
}
