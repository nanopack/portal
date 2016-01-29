package balance

import (
	"errors"
	"strconv"
	"strings"

	"github.com/nanopack/portal/database"
)

type (
	Balancer interface {
		GetService(id string) (*database.Service, error)
		SetService(service *database.Service) error
		DeleteService(id string) error
		GetServices() []*database.Service
		SetServices(services []database.Service) error

		GetServer(svcId, srvId string) (*database.Server, error)
		SetServer(svcId string, server *database.Server) error
		DeleteServer(svcId, srvId string) error
		SetServers(svcId string, servers []database.Server) error

		SyncToLvs() error // probably should just be Sync?
		SyncToPortal() error
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
	if len(srv) != 2 {
		return nil, NoServerError
	}
	p, _ := strconv.Atoi(srv[1])
	return &database.Server{Host: srv[0], Port: p}, nil
}
