package balance

import (
	"errors"
	"strconv"
	"strings"

	"github.com/nanobox-io/golang-lvs"

	"github.com/nanopack/portal/database"
)

type (
	Balancer interface {
		// need to update
		GetServer(service database.Service, server database.Server) database.Server
		SetServer(service database.Service, server database.Server) error
		DeleteServer(service database.Service, server database.Server) error
		SetServers(service database.Service, servers []database.Server) error
		SyncToLvs() error // probably should just be Sync?
		SyncToPortal() error
		// GetServers(service database.Service) []lvs.servers
		GetService(service database.Service) database.Service
		SetService(service database.Service) error
		DeleteService(service database.Service) error
		GetServices() []lvs.Service
		SetServices(services []database.Service) error
	}
)

var (
	NoServiceError = errors.New("No Service Found")
	NoServerError  = errors.New("No Server Found")
)

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
	if len(srv) != 3 {
		return nil, NoServerError
	}
	p, _ := strconv.Atoi(srv[1])
	return &database.Server{Host: srv[0], Port: p}, nil
}
