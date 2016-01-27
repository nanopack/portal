package database

import (
	"fmt"
	"net/url"
	// "sync"

	"github.com/coreos/go-iptables/iptables"
	// "github.com/nanobox-io/golang-lvs"

	"github.com/nanopack/portal/config"
)

type (
	Backender interface {
		// Init() error
		// GetServices() ([]lvs.Service, error)
		// GetService(service lvs.Service) (lvs.Service, error)
		// SetServices(services []lvs.Service) error
		// SetService(service lvs.Service) error
		// DeleteService(service lvs.Service) error

		// CURRENTLY IS
		Init() error
		GetServices() ([]Service, error)
		GetService(service Service) (Service, error)
		SetServices(services []Service) error
		SetService(service Service) error
		DeleteService(service Service) error

		// // NEEDS TO BE
		// GetServices() ([]Service, error)
		// GetService(string) (Service, error)
		// SetServices([]Service) error
		// SetService(Service) error
		// DeleteService(string) error
		// Init() error

		
		// GetServer(service database.Service, server database.Server) *lvs.Server
		// SetServer(service database.Service, server database.Server) error
		// DeleteServer(service database.Service, server database.Server) error
		// SetServers(service database.Service, servers []database.Server) error
		// SyncToLvs() error
		// SyncToPortal() error
		// // GetServers(service database.Service) []lvs.servers
		// GetService(service Service) *lvs.Service
		// SetService(service Service) error
		// DeleteService(service Service) error
		// GetServices() []lvs.Service
		// SetServices(services []Service) error

		// GetServices() ([]lvs.Service, error)
		// GetService(lvs.Service) (lvs.Service, error)
		// SetServices([]lvs.Service) error
		// SetService(lvs.Service) error
		// DeleteService(lvs.Service) error

	}
	Server struct {
		Id    string `json:"id,omitempty"`
		Ip    string `json:"ip"`
		Port  int `json:"port"`
	}
	Service struct {
		Id      string   `json:"id,omitempty"`
		Ip      string   `json:"ip"`
		Port    int   `json:"port"`
		Proto   string   `json:"proto"`
		Servers []Server `json:"servers,omitempty"`  // will we need?
	}
)

var (
	Backend  Backender
	Tab      *iptables.IPTables
	// ipvsLock *sync.RWMutex
)

func Init() error {
	// ipvsLock = &sync.RWMutex{}
	var err error
	var u *url.URL
	u, err = url.Parse(config.DatabaseConnection)
	if err != nil {
		return fmt.Errorf("Failed to parse db connection - %v", err)
	}
	switch u.Scheme {
	case "scribble":
		Backend = &ScribbleDatabase{}
	default:
		Backend = nil
	}
	if Backend != nil {
		err = Backend.Init()
		if err != nil {
			Backend = nil
		}
	}
	Tab, err = iptables.New()
	if err != nil {
		Tab = nil
	}
	if Tab != nil {
		Tab.Delete("filter", "INPUT", "-j", "portal")
		Tab.ClearChain("filter", "portal")
		Tab.DeleteChain("filter", "portal")
		err = Tab.NewChain("filter", "portal")
		if err != nil {
			return fmt.Errorf("Failed to create new chain - %v", err)
		}
		err = Tab.AppendUnique("filter", "portal", "-j", "RETURN")
		if err != nil {
			return fmt.Errorf("Failed to append to portal chain - %v", err)
		}
		err = Tab.AppendUnique("filter", "INPUT", "-j", "portal")
		if err != nil {
			return fmt.Errorf("Failed to append to INPUT chain - %v", err)
		}
	}
	return nil
}
