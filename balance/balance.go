// balance handles the load balancing portion of portal.
package balance

import (
	"errors"
	"strconv"
	"strings"

	"github.com/nanopack/portal/config"
	"github.com/nanopack/portal/core"
)

var (
	Balancer       core.Backender
	NoServiceError = errors.New("No Service Found")
	NoServerError  = errors.New("No Server Found")
)

func Init() error {
	if config.JustProxy {
		Balancer = nil
		return nil
	}
	Balancer = &Lvs{}

	return Balancer.Init()
}

func GetServices() ([]core.Service, error) {
	if Balancer == nil {
		return nil, nil
	}
	return Balancer.GetServices()
}

func GetService(id string) (*core.Service, error) {
	if Balancer == nil {
		return nil, nil
	}
	return Balancer.GetService(id)
}

func SetServices(services []core.Service) error {
	if Balancer == nil {
		return nil
	}
	return Balancer.SetServices(services)
}

func SetService(service *core.Service) error {
	if Balancer == nil {
		return nil
	}
	return Balancer.SetService(service)
}

func DeleteService(id string) error {
	if Balancer == nil {
		return nil
	}
	return Balancer.DeleteService(id)
}

func SetServers(svcId string, servers []core.Server) error {
	if Balancer == nil {
		return nil
	}
	return Balancer.SetServers(svcId, servers)
}

func SetServer(svcId string, server *core.Server) error {
	if Balancer == nil {
		return nil
	}
	return Balancer.SetServer(svcId, server)
}

func DeleteServer(svcId, srvId string) error {
	if Balancer == nil {
		return nil
	}
	return Balancer.DeleteServer(svcId, srvId)
}

func GetServer(svcId, srvId string) (*core.Server, error) {
	if Balancer == nil {
		return nil, nil
	}
	return Balancer.GetServer(svcId, srvId)
}

func parseSvc(serviceId string) (*core.Service, error) {
	if Balancer == nil {
		return nil, nil
	}
	s := strings.Replace(serviceId, "_", ".", -1)
	svc := strings.Split(s, "-")
	if len(svc) != 3 {
		return nil, NoServiceError
	}
	p, _ := strconv.Atoi(svc[2])
	return &core.Service{Type: svc[0], Host: svc[1], Port: p}, nil
}

func parseSrv(serverId string) (*core.Server, error) {
	if Balancer == nil {
		return nil, nil
	}
	s := strings.Replace(serverId, "_", ".", -1)
	srv := strings.Split(s, "-")
	if len(srv) != 2 {
		return nil, NoServerError
	}
	p, _ := strconv.Atoi(srv[1])
	return &core.Server{Host: srv[0], Port: p}, nil
}
