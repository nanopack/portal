package database

import (
	"errors"
	"fmt"
	"net/url"
	"sync"

	"github.com/coreos/go-iptables/iptables"
	"github.com/nanobox-io/golang-lvs"

	"github.com/nanopack/portal/config"
)

type (
	Backender interface {
		GetServices() ([]lvs.Service, error)
		GetService(lvs.Service) (lvs.Service, error)
		SetServices([]lvs.Service) error
		SetService(lvs.Service) error
		DeleteService(lvs.Service) error
		Init() error
	}
)

var (
	Backend        Backender
	ipvsLock       *sync.RWMutex
	NoServiceError = errors.New("No Service Found")
	NoServerError  = errors.New("No Server Found")
	tab            *iptables.IPTables
)

func init() {
	ipvsLock = &sync.RWMutex{}
	var err error
	var u *url.URL
	u, err = url.Parse(config.DatabaseConnection)
	if err != nil {
		return
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
			return
		}
		err = tab.AppendUnique("filter", "portal", "-j", "RETURN")
		if err != nil {
			return
		}
		err = tab.AppendUnique("filter", "INPUT", "-j", "portal")
		if err != nil {
			return
		}
	}
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
	s := lvs.DefaultIpvs.FindService(service)
	if s == nil {
		return NoServiceError
	}
	err := s.AddServer(server)
	if err != nil {
		return err
	}
	// save to backend
	if Backend != nil {
		err = Backend.SetService(*s)
		if err != nil {
			return err
		}
	}
	return nil
}

// DeleteServer
func DeleteServer(service lvs.Service, server lvs.Server) error {
	ipvsLock.Lock()
	defer ipvsLock.Unlock()
	// remove from lvs
	s := lvs.DefaultIpvs.FindService(service)
	if s == nil {
		return nil
	}
	s.RemoveServer(server)
	// remove from backend
	if Backend != nil {
		err := Backend.SetService(*s)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetServers
// func GetServers(service lvs.Service) []lvs.servers {
// 	return service.Servers
// }

// SetServers
func SetServers(service lvs.Service, servers []lvs.Server) error {
	ipvsLock.Lock()
	defer ipvsLock.Unlock()
	// add to lvs
	s := lvs.DefaultIpvs.FindService(service)
	if s == nil {
		return NoServiceError
	}
	// Add Servers
AddServers:
	for i := range servers {
		for j := range s.Servers {
			if servers[i].Host == s.Servers[j].Host && servers[i].Port == s.Servers[j].Port {
				continue AddServers
			}
		}
		s.AddServer(servers[i])
	}
	// Remove Servers
RemoveServers:
	for i := range s.Servers {
		for j := range servers {
			if s.Servers[i].Host == servers[j].Host && s.Servers[i].Port == servers[j].Port {
				continue RemoveServers
			}
		}
		s.RemoveServer(s.Servers[i])
	}
	return nil
}

// GetService
func GetService(service lvs.Service) *lvs.Service {
	ipvsLock.RLock()
	defer ipvsLock.RUnlock()
	return lvs.DefaultIpvs.FindService(service)
}

// SetService
func SetService(service lvs.Service) error {
	ipvsLock.Lock()
	defer ipvsLock.Unlock()
	// add to lvs
	err := lvs.DefaultIpvs.AddService(service)
	if err != nil {
		return err
	}
	// save to backend
	if Backend != nil {
		err := Backend.SetService(service)
		if err != nil {
			return err
		}
	}
	if tab != nil {
		err := tab.Insert("filter", "portal", 0, "-p", service.Type, "-d", service.Host, "--dport", fmt.Sprintf("%d", service.Port), "-j", "ACCEPT")
		if err != nil {
			return err
		}
	}
	return nil
}

// DeleteService
func DeleteService(service lvs.Service) error {
	ipvsLock.Lock()
	defer ipvsLock.Unlock()
	// remove from lvs
	err := lvs.DefaultIpvs.RemoveService(service)
	if err != nil {
		return err
	}
	// remove from backend
	if Backend != nil {
		err := Backend.DeleteService(service)
		if err != nil {
			return err
		}
	}
	if tab != nil {
		err := tab.Delete("filter", "portal", "-p", service.Type, "-d", service.Host, "--dport", fmt.Sprintf("%d", service.Port), "-j", "ACCEPT")
		if err != nil {
			return err
		}
	}
	return nil
}

// GetServices
func GetServices() []lvs.Service {
	ipvsLock.RLock()
	defer ipvsLock.RUnlock()
	return lvs.DefaultIpvs.Services
}

// SetServices
func SetServices(services []lvs.Service) error {
	ipvsLock.Lock()
	defer ipvsLock.Unlock()
	if tab != nil {
		tab.RenameChain("filter", "portal", "portal-old")
	}
	lvs.DefaultIpvs.Restore(services)
	if Backend != nil {
		err := Backend.SetServices(services)
		if err != nil {
			if tab != nil {
				tab.RenameChain("filter", "portal-old", "portal")
			}
			return err
		}
	}
	if tab != nil {
		tab.NewChain("filter", "portal")
		tab.ClearChain("filter", "portal")
		tab.AppendUnique("filter", "portal", "-j", "RETURN")
		for i := range services {
			err := tab.Insert("filter", "portal", 0, "-p", services[i].Type, "-d", services[i].Host, "--dport", fmt.Sprintf("%d", services[i].Port), "-j", "ACCEPT")
			if err != nil {
				tab.ClearChain("filter", "portal")
				tab.DeleteChain("filter", "portal")
				tab.RenameChain("filter", "portal-old", "portal")
				return err
			}
		}
		tab.AppendUnique("filter", "INPUT", "-j", "portal")
		tab.Delete("filter", "INPUT", "-j", "portal-old")
		tab.ClearChain("filter", "portal-old")
		tab.DeleteChain("filter", "portal-old")
	}
	return nil
}

// SyncLvs
func SyncToLvs() error {
	ipvsLock.Lock()
	defer ipvsLock.Unlock()
	if tab != nil {
		tab.RenameChain("filter", "portal", "portal-old")
	}
	var err error
	var services []lvs.Service
	if Backend != nil {
		services, err = Backend.GetServices()
		if err != nil {
			if tab != nil {
				tab.RenameChain("filter", "portal-old", "portal")
			}
			return err
		}
	} else {
		services = []lvs.Service{}
	}
	err = lvs.Restore(services)
	if err != nil {
		if tab != nil {
			tab.RenameChain("filter", "portal-old", "portal")
		}
		return err
	}
	if tab != nil {
		tab.NewChain("filter", "portal")
		tab.ClearChain("filter", "portal")
		tab.AppendUnique("filter", "portal", "-j", "RETURN")
		for i := range services {
			err := tab.Insert("filter", "portal", 0, "-p", services[i].Type, "-d", services[i].Host, "--dport", fmt.Sprintf("%d", services[i].Port), "-j", "ACCEPT")
			if err != nil {
				tab.ClearChain("filter", "portal")
				tab.DeleteChain("filter", "portal")
				tab.RenameChain("filter", "portal-old", "portal")
				return err
			}
		}
		tab.AppendUnique("filter", "INPUT", "-j", "portal")
		tab.Delete("filter", "INPUT", "-j", "portal-old")
		tab.ClearChain("filter", "portal-old")
		tab.DeleteChain("filter", "portal-old")
	}
	return nil
}

// SyncToPortal
func SyncToPortal() error {
	ipvsLock.Lock()
	defer ipvsLock.Unlock()
	if tab != nil {
		tab.RenameChain("filter", "portal", "portal-old")
	}
	err := lvs.Save()
	if err != nil {
		if tab != nil {
			tab.RenameChain("filter", "portal-old", "portal")
		}
		return err
	}
	if Backend != nil {
		err := Backend.SetServices(lvs.DefaultIpvs.Services)
		if err != nil {
			if tab != nil {
				tab.RenameChain("filter", "portal-old", "portal")
			}
			return err
		}
	}
	if tab != nil {
		tab.NewChain("filter", "portal")
		tab.ClearChain("filter", "portal")
		tab.AppendUnique("filter", "portal", "-j", "RETURN")
		for i := range services {
			err := tab.Insert("filter", "portal", 0, "-p", services[i].Type, "-d", services[i].Host, "--dport", fmt.Sprintf("%d", services[i].Port), "-j", "ACCEPT")
			if err != nil {
				tab.ClearChain("filter", "portal")
				tab.DeleteChain("filter", "portal")
				tab.RenameChain("filter", "portal-old", "portal")
				return err
			}
		}
		tab.AppendUnique("filter", "INPUT", "-j", "portal")
		tab.Delete("filter", "INPUT", "-j", "portal-old")
		tab.ClearChain("filter", "portal-old")
		tab.DeleteChain("filter", "portal-old")
	}
	return nil
}
