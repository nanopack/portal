package balance

import (
	"errors"

	"github.com/nanobox-io/golang-lvs"

	"github.com/nanopack/portal/database"
)

type (
	Balancer interface {
		// Init() error
		// GetServer(service lvs.Service, server lvs.Server) *lvs.Server
		// SetServer(service lvs.Service, server lvs.Server) error
		// DeleteServer(service lvs.Service, server lvs.Server) error
		// // GetServers(service lvs.Service) []lvs.servers
		// SetServers(service lvs.Service, servers []lvs.Server) error
		// GetService(service lvs.Service) *lvs.Service
		// SetService(service lvs.Service) error
		// DeleteService(service lvs.Service) error
		// GetServices() []lvs.Service
		// SetServices(services []lvs.Service) error
		// SyncToLvs() error
		// SyncToPortal() error

		// database. maybe? or what
		// Init() error
		// GetServer(service database.Service, server database.Server) *database.Server
		// SetServer(service database.Service, server database.Server) error
		// DeleteServer(service database.Service, server database.Server) error
		// // GetServers(service database.Service) []database.servers
		// SetServers(service database.Service, servers []database.Server) error
		// GetService(service database.Service) *database.Service
		// SetService(service database.Service) error
		// DeleteService(service database.Service) error
		// GetServices() []database.Service
		// SetServices(services []database.Service) error
		// SyncToLvs() error
		// SyncToPortal() error

		GetServer(service database.Service, server database.Server) *lvs.Server
		SetServer(service database.Service, server database.Server) error
		DeleteServer(service database.Service, server database.Server) error
		SetServers(service database.Service, servers []database.Server) error
		SyncToLvs() error
		SyncToPortal() error
		// GetServers(service database.Service) []lvs.servers
		GetService(service database.Service) *lvs.Service
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