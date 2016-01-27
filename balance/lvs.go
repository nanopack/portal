package balance

import (
	"errors"
	"fmt"
	"strconv"
	"sync"

	"github.com/nanobox-io/golang-lvs"

	"github.com/nanopack/portal/database"
)

var (
	Backend        = database.Backend //&database.Backend
	ipvsLock       = &sync.RWMutex{}
	// ipvsLock       *sync.RWMutex
	// ipvsLock       = &database.IpvsLock
	NoServiceError = errors.New("No Service Found")
	NoServerError  = errors.New("No Server Found")
	tab            = database.Tab //&database.Tab
)

type (
	Lvs struct {
	}
)

// databaseify the get-server
// move to an lvs.GetServer
// GetServer
func (l *Lvs) GetServer(service database.Service, server database.Server) *lvs.Server {
	// error would have been caught on json marshalling
	p, _ := strconv.Atoi(service.Port)
	lvsService := lvs.Service{Type: service.Proto, Host: service.Ip, Port: p}
	ipvsLock.RLock()
	defer ipvsLock.RUnlock()
	s := lvs.DefaultIpvs.FindService(lvsService)
	if s == nil {
		return nil
	}
	p, _ = strconv.Atoi(server.Port)
	lvsServer := lvs.Server{Host: server.Ip, Port: p}
	return s.FindServer(lvsServer)
}

// SetServer
func (l *Lvs) SetServer(service database.Service, server database.Server) error {
	p, _ := strconv.Atoi(service.Port)
	lvsService := lvs.Service{Type: service.Proto, Host: service.Ip, Port: p}
	p, _ = strconv.Atoi(server.Port)
	lvsServer := lvs.Server{Host: server.Ip, Port: p}

	ipvsLock.Lock()
	defer ipvsLock.Unlock()
	// add to lvs
	s := lvs.DefaultIpvs.FindService(lvsService)
	if s == nil {
		return NoServiceError
	}
	err := s.AddServer(lvsServer)
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
func (l *Lvs) DeleteServer(service database.Service, server database.Server) error {
	p, _ := strconv.Atoi(service.Port)
	lvsService := lvs.Service{Type: service.Proto, Host: service.Ip, Port: p}
	p, _ = strconv.Atoi(server.Port)
	lvsServer := lvs.Server{Host: server.Ip, Port: p}

	ipvsLock.Lock()
	defer ipvsLock.Unlock()
	// remove from lvs
	s := lvs.DefaultIpvs.FindService(lvsService)
	if s == nil {
		return nil
	}
	s.RemoveServer(lvsServer)
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
// func (l *Lvs) GetServers(service database.Service) []lvs.servers {
// 	return service.Servers
// }

// SetServers
func (l *Lvs) SetServers(service database.Service, servers []database.Server) error {
	p, _ := strconv.Atoi(service.Port)
	lvsService := lvs.Service{Type: service.Proto, Host: service.Ip, Port: p}
	lvsServers := []lvs.Server{}
	for _, srv := range servers {
		p, _ = strconv.Atoi(srv.Port)
		lvsServers = append(lvsServers, lvs.Server{Host: srv.Ip, Port: p})
	}

	ipvsLock.Lock()
	defer ipvsLock.Unlock()
	// add to lvs
	s := lvs.DefaultIpvs.FindService(lvsService)
	if s == nil {
		return NoServiceError
	}
	// Add Servers
AddServers:
	for i := range lvsServers {
		for j := range s.Servers {
			if lvsServers[i].Host == s.Servers[j].Host && lvsServers[i].Port == s.Servers[j].Port {
				// what is this? goto?
				continue AddServers
			}
		}
		s.AddServer(lvsServers[i])
	}
	// Remove Servers
RemoveServers:
	for i := range s.Servers {
		for j := range lvsServers {
			if s.Servers[i].Host == lvsServers[j].Host && s.Servers[i].Port == lvsServers[j].Port {
				continue RemoveServers
			}
		}
		s.RemoveServer(s.Servers[i])
	}
	return nil
}

// GetService
func (l *Lvs) GetService(service database.Service) *lvs.Service {
	p, _ := strconv.Atoi(service.Port)
	lvsService := lvs.Service{Type: service.Proto, Host: service.Ip, Port: p}

	ipvsLock.RLock()
	defer ipvsLock.RUnlock()
	return lvs.DefaultIpvs.FindService(lvsService)
}

// SetService
func (l *Lvs) SetService(service database.Service) error {
	p, _ := strconv.Atoi(service.Port)
	lvsService := lvs.Service{Type: service.Proto, Host: service.Ip, Port: p}

	ipvsLock.Lock()
	defer ipvsLock.Unlock()
	// add to lvs
	err := lvs.DefaultIpvs.AddService(lvsService)
	if err != nil {
		return err
	}
	// save to backend
	if Backend != nil {
		err := Backend.SetService(lvsService)
		if err != nil {
			return err
		}
	}
	if tab != nil {
		err := tab.Insert("filter", "portal", 1, "-p", lvsService.Type, "-d", lvsService.Host, "--dport", fmt.Sprintf("%d", lvsService.Port), "-j", "ACCEPT")
		if err != nil {
			return err
		}
	}
	return nil
}

// DeleteService
func (l *Lvs) DeleteService(service database.Service) error {
	p, _ := strconv.Atoi(service.Port)
	lvsService := lvs.Service{Type: service.Proto, Host: service.Ip, Port: p}

	ipvsLock.Lock()
	defer ipvsLock.Unlock()
	// remove from lvs
	err := lvs.DefaultIpvs.RemoveService(lvsService)
	if err != nil {
		return err
	}
	// remove from backend
	if Backend != nil {
		err := Backend.DeleteService(lvsService)
		if err != nil {
			return err
		}
	}
	if tab != nil {
		err := tab.Delete("filter", "portal", "-p", lvsService.Type, "-d", lvsService.Host, "--dport", fmt.Sprintf("%d", lvsService.Port), "-j", "ACCEPT")
		if err != nil {
			return err
		}
	}
	return nil
}

// GetServices
func (l *Lvs) GetServices() []lvs.Service {
	ipvsLock.RLock()
	defer ipvsLock.RUnlock()
	return lvs.DefaultIpvs.Services
}

// SetServices
func (l *Lvs) SetServices(services []database.Service) error {
	lvsServices := []lvs.Service{}
	for _, svc := range services {
		p, _ := strconv.Atoi(svc.Port)
		lvsServices = append(lvsServices, lvs.Service{Type: svc.Proto, Host: svc.Ip, Port: p})
	}
	ipvsLock.Lock()
	defer ipvsLock.Unlock()
	if tab != nil {
		tab.RenameChain("filter", "portal", "portal-old")
	}
	err := lvs.DefaultIpvs.Clear()
	if err != nil {
		if tab != nil {
			tab.RenameChain("filter", "portal-old", "portal")
		}
		return err
	}
	err = lvs.DefaultIpvs.Restore(lvsServices)
	if err != nil {
		if tab != nil {
			tab.RenameChain("filter", "portal-old", "portal")
		}
		return err
	}
	if Backend != nil {
		err := Backend.SetServices(lvsServices)
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
		for i := range lvsServices {
			err := tab.Insert("filter", "portal", 1, "-p", lvsServices[i].Type, "-d", lvsServices[i].Host, "--dport", fmt.Sprintf("%d", lvsServices[i].Port), "-j", "ACCEPT")
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
func (l *Lvs) SyncToLvs() error {
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
	err = lvs.Clear()
	if err != nil {
		if tab != nil {
			tab.RenameChain("filter", "portal-old", "portal")
		}
		return err
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
			err := tab.Insert("filter", "portal", 1, "-p", services[i].Type, "-d", services[i].Host, "--dport", fmt.Sprintf("%d", services[i].Port), "-j", "ACCEPT")
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
func (l *Lvs) SyncToPortal() error {
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
		for i := range lvs.DefaultIpvs.Services {
			err := tab.Insert("filter", "portal", 1, "-p", lvs.DefaultIpvs.Services[i].Type, "-d", lvs.DefaultIpvs.Services[i].Host, "--dport", fmt.Sprintf("%d", lvs.DefaultIpvs.Services[i].Port), "-j", "ACCEPT")
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
