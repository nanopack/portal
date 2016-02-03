package balance

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/coreos/go-iptables/iptables"

	"github.com/nanopack/portal/database"
)

type (
	Balanceable interface {
		GetServices() []database.Service
		GetService(id string) (*database.Service, error)
		SetServices(services []database.Service) error
		SetService(service *database.Service) error
		DeleteService(id string) error

		SetServers(svcId string, servers []database.Server) error
		SetServer(svcId string, server *database.Server) error
		DeleteServer(svcId, srvId string) error
		GetServer(svcId, srvId string) (*database.Server, error)

		Sync() error // is this even needed?
	}
)

var (
	Balancer       Balanceable
	tab            *iptables.IPTables
	NoServiceError = errors.New("No Service Found")
	NoServerError  = errors.New("No Server Found")
)

func Init() error {
	Balancer = &Lvs{}

	var err error
	tab, err = iptables.New()
	if err != nil {
		tab = nil
	}
	if tab != nil {
		tab.Delete("filter", "INPUT", "-j", "portal")
		tab.ClearChain("filter", "portal")
		tab.DeleteChain("filter", "portal")
		err = tab.NewChain("filter", "portal")
		if err != nil {
			return fmt.Errorf("Failed to create new chain - %v", err)
		}
		err = tab.AppendUnique("filter", "portal", "-j", "RETURN")
		if err != nil {
			return fmt.Errorf("Failed to append to portal chain - %v", err)
		}
		err = tab.AppendUnique("filter", "INPUT", "-j", "portal")
		if err != nil {
			return fmt.Errorf("Failed to append to INPUT chain - %v", err)
		}
	}
	return nil
}

func GetServices() []database.Service {
	return Balancer.GetServices()
}

func GetService(id string) (*database.Service, error) {
	return Balancer.GetService(id)
}

func SetServices(services []database.Service) error {
	return Balancer.SetServices(services)
}

func SetService(service *database.Service) error {
	return Balancer.SetService(service)
}

func DeleteService(id string) error {
	return Balancer.DeleteService(id)
}

func SetServers(svcId string, servers []database.Server) error {
	return Balancer.SetServers(svcId, servers)
}

func SetServer(svcId string, server *database.Server) error {
	return Balancer.SetServer(svcId, server)
}

func DeleteServer(svcId, srvId string) error {
	return Balancer.DeleteServer(svcId, srvId)
}

func GetServer(svcId, srvId string) (*database.Server, error) {
	return Balancer.GetServer(svcId, srvId)
}

func Sync() error {
	return Balancer.Sync()
}

func parseSvc(serviceId string) (*database.Service, error) {
	s := strings.Replace(serviceId, "_", ".", -1)
	svc := strings.Split(s, "-")
	if len(svc) != 3 {
		return nil, NoServiceError
	}
	p, _ := strconv.Atoi(svc[2])
	return &database.Service{Type: svc[0], Host: svc[1], Port: p}, nil
}

func parseSrv(serverId string) (*database.Server, error) {
	s := strings.Replace(serverId, "_", ".", -1)
	srv := strings.Split(s, "-")
	if len(srv) != 2 {
		return nil, NoServerError
	}
	p, _ := strconv.Atoi(srv[1])
	return &database.Server{Host: srv[0], Port: p}, nil
}
