package balance

import (
	"errors"

	"github.com/nanobox-io/golang-lvs"

	"github.com/nanopack/portal/database"
)

type (
	Balancer interface {
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
