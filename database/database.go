package database

import (
	"sync"

	"github.com/nanobox-io/golang-lvs"
)

type (
	Something interface {
		GetServices() ([]lvs.Service, error)
		GetService() (lvs.Service, error)
		SetServices([]lvs.Service) error
		SetService(lvs.Service) error
		DeleteService(lvs.Service) error
	}
)

var (
	Backends    map[string]Something
	Backend     Something
	backendLock *sync.RWMutex
	ipvsLock    *sync.RWMutex
)

func init() {
	backendLock = &sync.Mutex{}
	ipvsLock = &sync.Mutex{}
}

// GetServer
func GetServer(service lvs.Service, server lvs.Server) *lvs.Server {
	ipvsLock.RLock()
	defer ipvsLock.RUnlock()
	s := lvs.DefaultIpvs.FindService(service)
	if s == nil {
		return nil
	}
	return s.FindServer(server)
}

// SetServer
func SetServer(service lvs.Service, server lvs.Server) error {
	ipvsLock.Lock()
	defer ipvsLock.Unlock()
	// add to lvs
	// save to backend
	Backend.SetService(service)
	return nil
}

// DeleteServer
func DeleteServer(service lvs.Service, host string, port int) error {
	ipvsLock.Lock()
	defer ipvsLock.Unlock()
	// remove from backend
	Backend.SetService(service)
	// remove from lvs
	return nil
}

// GetServers
// func GetServers(service lvs.Service) []lvs.servers {
// 	return service.Servers
// }

// SetServers
// GetService
func GetService(service lvs.Service) lvs.Service {
	ipvsLock.RLock()
	defer ipvsLock.RUnlock()
	return lvs.DefaultIpvs.FindService(service)
}

// SetService
func SetService(service lvs.Service) error {
	ipvsLock.Lock()
	defer ipvsLock.Unlock()
	// add to lvs
	// save to backend
	Backend.SetService(service)
	return nil
}

// DeleteService
func DeleteService(proto, host string, port int) error {
	ipvsLock.Lock()
	defer ipvsLock.Unlock()
	// remove from backend
	Backend.DeleteService(service)
	// remove from lvs
	return nil
}

// GetServices
// SetServices

// SyncLvs
func SyncToLvs() error {
	ipvsLock.Lock()
	defer ipvsLock.Unlock()
	return lvs.Restore(services)
}

// SyncToPortal
func SyncToPortal() error {
	ipvsLock.Lock()
	defer ipvsLock.Unlock()
	var err error
	services, err = lvs.Save(services)
	return err
}
