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
	"github.com/nanopack/portal/routemgr"
)

var (
	skip = false // skip if redis-server not installed

	testService1 = core.Service{Id: "tcp-192_168_0_15-80", Host: "192.168.0.15", Port: 80, Type: "tcp", Scheduler: "wrr"}
	testService2 = core.Service{Id: "tcp-192_168_0_16-80", Host: "192.168.0.16", Port: 80, Type: "tcp", Scheduler: "wrr"}
	testServer1  = core.Server{Id: "127_0_0_11-8080", Host: "127.0.0.11", Port: 8080, Forwarder: "m", Weight: 5, UpperThreshold: 10, LowerThreshold: 1}
	testServer2  = core.Server{Id: "127_0_0_12-8080", Host: "127.0.0.12", Port: 8080, Forwarder: "m", Weight: 5, UpperThreshold: 10, LowerThreshold: 1}
)

func TestMain(m *testing.M) {
	// clean test dir
	os.RemoveAll("/tmp/clusterTest")

	// initialize backend if redis-server found
	initialize()

	rtn := m.Run()

	// clean test dir
	os.RemoveAll("/tmp/clusterTest")

	conn, err := redis.DialURL(config.ClusterConnection, redis.DialConnectTimeout(30*time.Second), redis.DialPassword(config.ClusterToken))
	if err != nil {
		return
	}
	hostname, _ := os.Hostname()
	self := fmt.Sprintf("%v:%v", hostname, config.ApiPort)
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

	config.RoutePortHttp = 9082
	config.RoutePortTls = 9445
	config.Log = lumber.NewConsoleLogger(lumber.LvlInt("FATAL"))

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

		// initialize routemgr
		err = routemgr.Init()
		if err != nil {
			fmt.Printf("Routemgr init failed - %v\n", err)
			os.Exit(1)
		}

		err = cluster.Init()
		if err != nil {
			fmt.Printf("cluster init failed - %v\n", err)
			os.Exit(1)
		}

	}
}
