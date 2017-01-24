package balance_test

import (
	"fmt"
	"os/exec"
	"testing"

	"github.com/jcelliott/lumber"

	"github.com/nanopack/portal/balance"
	"github.com/nanopack/portal/config"
	"github.com/nanopack/portal/core"
)

////////////////////////////////////////////////////////////////////////////////
// SERVICES
////////////////////////////////////////////////////////////////////////////////
func TestSetServiceNginx(t *testing.T) {
	if !nginxPrep() {
		t.SkipNow()
	}
	if err := balance.SetService(&testService1); err != nil {
		t.Errorf("Failed to SET service - %s", err)
		t.FailNow()
	}

	if err := balance.SetService(&testService1); err != nil {
		t.Errorf("Failed to SET service - %s", err)
		t.FailNow()
	}

	service, err := balance.GetService(testService1.Id)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if service.Host != testService1.Host {
		t.Errorf("Read service differs from written service")
	}
}

func TestSetServicesNginx(t *testing.T) {
	if !nginxPrep() {
		t.SkipNow()
	}

	if err := balance.SetService(&testService1); err != nil {
		t.Errorf("Failed to SET services - %s", err)
		t.FailNow()
	}

	if err := balance.SetServices([]core.Service{testService2}); err != nil {
		t.Errorf("Failed to SET services - %s", err)
		t.FailNow()
	}

	_, err := balance.GetService(testService1.Id)
	if err == nil {
		t.Errorf("Failed to clear old services on PUT - %s", err)
		t.FailNow()
	}

	service, err := balance.GetService(testService2.Id)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if service.Host != testService2.Host {
		t.Errorf("Read service differs from written service")
	}
}

func TestGetServicesNginx(t *testing.T) {
	if !nginxPrep() {
		t.SkipNow()
	}

	if err := balance.SetServices([]core.Service{testService2}); err != nil {
		t.Errorf("Failed to SET services - %s", err)
		t.FailNow()
	}

	services, err := balance.GetServices()
	if err != nil {
		t.Errorf("Failed to GET services - %s", err)
		t.FailNow()
	}

	if services[0].Id != testService2.Id {
		t.Errorf("Read service differs from written service")
	}
}

func TestGetServiceNginx(t *testing.T) {
	if !nginxPrep() {
		t.SkipNow()
	}

	service, err := balance.GetService(testService2.Id)
	if err == nil {
		t.Errorf("Failed to fail GETTING service - %s", err)
		t.FailNow()
	}

	if err := balance.SetServices([]core.Service{testService2}); err != nil {
		t.Errorf("Failed to SET services - %s", err)
		t.FailNow()
	}

	service, err = balance.GetService(testService2.Id)
	if err != nil {
		t.Errorf("Failed to GET service - %s", err)
		t.FailNow()
	}

	if service.Id != testService2.Id {
		t.Errorf("Read service differs from written service")
	}
}

func TestDeleteServiceNginx(t *testing.T) {
	if !nginxPrep() {
		t.SkipNow()
	}
	if err := balance.DeleteService(testService2.Id); err != nil {
		t.Errorf("Failed to DELETE nonexistant service - %s", err)
	}

	balance.SetService(&testService2)

	if err := balance.DeleteService(testService2.Id); err != nil {
		t.Errorf("Failed to DELETE service - %s", err)
	}

	_, err := balance.GetService(testService2.Id)
	if err == nil {
		t.Error(err)
	}
}

////////////////////////////////////////////////////////////////////////////////
// SERVERS
////////////////////////////////////////////////////////////////////////////////

func TestSetServerNginx(t *testing.T) {
	if !nginxPrep() {
		t.SkipNow()
	}

	balance.SetService(&testService1)

	if err := balance.SetServer(testService1.Id, &testServer1); err != nil {
		t.Errorf("Failed to SET server - %s", err)
		t.FailNow()
	}

	if err := balance.SetServer(testService1.Id, &testServer1); err != nil {
		t.Errorf("Failed to SET server - %s", err)
		t.FailNow()
	}

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

func TestSetServersNginx(t *testing.T) {
	if !nginxPrep() {
		t.SkipNow()
	}

	servers := []core.Server{testServer2}
	if err := balance.SetServers(testService1.Id, servers); err == nil {
		t.Errorf("Failed to fail SETTING servers - %s", err)
		t.FailNow()
	}

	balance.SetService(&testService1)

	if err := balance.SetServers(testService1.Id, servers); err != nil {
		t.Errorf("Failed to SET servers - %s", err)
		t.FailNow()
	}

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

func TestGetServersNginx(t *testing.T) {
	if !nginxPrep() {
		t.SkipNow()
	}

	svc := testService1
	svc.Servers = append(svc.Servers, testServer2)
	balance.SetService(&svc)

	service, err := balance.GetService(testService1.Id)
	if err != nil {
		t.Errorf("Failed to GET servers - %s", err)
		t.FailNow()
	}

	if service.Servers[0].Id != testServer2.Id {
		t.Errorf("Read server differs from written server")
	}
}

func TestGetServerNginx(t *testing.T) {
	if !nginxPrep() {
		t.SkipNow()
	}

	server, err := balance.GetServer(testService1.Id, "not-real")
	if err == nil {
		t.Errorf("Failed to fail GETTING server - %s", err)
		t.FailNow()
	}

	balance.SetService(&testService1)

	server, err = balance.GetServer(testService1.Id, "not-real")
	if err == nil {
		t.Errorf("Failed to fail GETTING server - %s", err)
		t.FailNow()
	}

	svc := testService1
	svc.Servers = append(svc.Servers, testServer2)
	balance.SetService(&svc)

	server, err = balance.GetServer(testService1.Id, testServer2.Id)
	if err != nil {
		t.Errorf("Failed to GET server - %s", err)
		t.FailNow()
	}

	if server.Id != testServer2.Id {
		t.Errorf("Read server differs from written server")
	}
}

func TestDeleteServerNginx(t *testing.T) {
	if !nginxPrep() {
		t.SkipNow()
	}

	err := balance.DeleteServer("not-real-thing", testServer2.Id)
	if err == nil {
		t.Errorf("Failed to DELETE nonexistant server - %s", err)
	}

	balance.SetService(&testService1)

	err = balance.DeleteServer(testService1.Id, testServer2.Id)
	if err != nil {
		t.Errorf("Failed to DELETE nonexistant server - %s", err)
	}

	balance.SetServer(testService1.Id, &testServer2)

	err = balance.DeleteServer(testService1.Id, testServer2.Id)
	if err != nil {
		t.Errorf("Failed to DELETE server - %s", err)
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
func nginxPrep() bool {
	nginx, err := exec.Command("which", "nginx").CombinedOutput()
	// nginx, err := exec.Command("nginx", "-v").CombinedOutput()
	if err != nil {
		fmt.Printf("Failed to run nginx - %s%s\n", nginx, err)
		return false
	}

	config.Log = lumber.NewConsoleLogger(lumber.LvlInt("FATAL"))

	config.Balancer = "nginx"
	config.WorkDir = "/tmp/portal"

	// todo: write config file for tests to /tmp/portal/portal-nginx.conf

	err = balance.Init()
	// skip tests if failed to init
	if err != nil {
		fmt.Printf("Failed to initialize nginx - %s%s\n", nginx, err)
		return false
	}
	return true
}
