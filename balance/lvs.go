package balance

import (
	"fmt"
	"sync"

	"github.com/nanobox-io/golang-lvs"

	"github.com/nanopack/portal/database"
)

var (
	Backend  = database.Backend //&database.Backend
	ipvsLock = &sync.RWMutex{}
)

type (
	Lvs struct {
	}
	lookupService struct {
		Type string
		Host string
		Port int
	}
	lookupServer struct {
		Host string
		Port int
	}
)

// GetServer
func (l *Lvs) GetServer(svcId, srvId string) (*database.Server, error) {
	// break up ids in order to create dummy lvs.Service
	var err error
	service, err := parseSvc(svcId)
	if err != nil {
		return nil, err
	}
	lvsService := svcToL(*service)

	server, err := parseSrv(srvId)
	if err != nil {
		return nil, err
	}
	lvsServer := srvToL(*server)

	ipvsLock.RLock()
	defer ipvsLock.RUnlock()
	s := lvs.DefaultIpvs.FindService(lvsService)
	if s == nil {
		return nil, NoServerError
	}
	return lToSrvp(s.FindServer(lvsServer)), nil
}

// SetServer
func (l *Lvs) SetServer(svcId string, server *database.Server) error {
	service, err := l.GetService(svcId)
	if err != nil {
		return err
	}
	lvsService := svcToL(*service)
	lvsServer := srvToL(*server)

	ipvsLock.Lock()
	defer ipvsLock.Unlock()

	// seems redundant since api passes good service in. todo: Should api pass (string, Server)?
	// add to lvs
	s := lvs.DefaultIpvs.FindService(lvsService)
	if s == nil {
		return NoServiceError
	}
	//

	err = s.AddServer(lvsServer)
	if err != nil {
		return err
	}
	return nil
}

// DeleteServer
func (l *Lvs) DeleteServer(svcId, srvId string) error {
	var err error
	service, err := parseSvc(svcId)
	if err != nil {
		// if service not valic, 'delete' successful
		return nil
	}
	lvsService := svcToL(*service)

	server, err := parseSrv(srvId)
	if err != nil {
		// if invalid servername, 'delete' successful
		return nil
	}
	lvsServer := srvToL(*server)

	ipvsLock.Lock()
	defer ipvsLock.Unlock()
	// remove from lvs
	s := lvs.DefaultIpvs.FindService(lvsService)
	if s == nil {
		return nil
	}
	// hope this doesn't fail
	s.RemoveServer(lvsServer)

	return nil
}

// SetServers
func (l *Lvs) SetServers(svcId string, servers []database.Server) error {
	service, err := l.GetService(svcId)
	if err != nil {
		return nil
	}

	lvsService := svcToL(*service)
	lvsServers := []lvs.Server{}
	for _, srv := range servers {
		lvsServers = append(lvsServers, srvToL(srv))
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
func (l *Lvs) GetService(id string) (*database.Service, error) {
	service, err := parseSvc(id)
	if err != nil {
		return nil, err
	}
	lvsService := lvs.Service{Type: service.Type, Host: service.Host, Port: service.Port}

	ipvsLock.RLock()
	defer ipvsLock.RUnlock()
	// why doesn't this always return nil on none found?
	svc := lvs.DefaultIpvs.FindService(lvsService)
	if svc == nil {
		return nil, NoServiceError
	}

	return lToSvcp(svc), nil
}

// SetService
func (l *Lvs) SetService(service *database.Service) error {
	lvsService := svcToL(*service)

	ipvsLock.Lock()
	defer ipvsLock.Unlock()
	// add to lvs
	err := lvs.DefaultIpvs.AddService(lvsService)
	if err != nil {
		return err
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
func (l *Lvs) DeleteService(id string) error {
	service, err := parseSvc(id)
	if err != nil {
		// if invalid service, 'delete' succesful
		return nil
	}

	lvsService := lvs.Service{Type: service.Type, Host: service.Host, Port: service.Port}

	ipvsLock.Lock()
	defer ipvsLock.Unlock()
	svc := lvs.DefaultIpvs.FindService(lvsService)
	if svc == nil {
		// if not exist, 'delete' successful
		return nil
	}

	// remove from lvs
	err = lvs.DefaultIpvs.RemoveService(lvsService)
	if err != nil {
		return err
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
// doesn't need to be a pointer method because it doesn't modify original object
func (l *Lvs) GetServices() []database.Service {
	ipvsLock.RLock()
	defer ipvsLock.RUnlock()
	svcs := []database.Service{}
	for _, svc := range lvs.DefaultIpvs.Services {
		svcs = append(svcs, lToSvc(svc))
	}
	return svcs
}

// SetServices
func (l *Lvs) SetServices(services []database.Service) error {
	lvsServices := []lvs.Service{}
	for _, svc := range services {
		lvsServices = append(lvsServices, svcToL(svc))
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
func (l *Lvs) SyncToBalancer(services []database.Service) error {
	ipvsLock.Lock()
	defer ipvsLock.Unlock()
	if tab != nil {
		tab.RenameChain("filter", "portal", "portal-old")
	}
	var err error

	err = lvs.Clear()
	if err != nil {
		if tab != nil {
			tab.RenameChain("filter", "portal-old", "portal")
		}
		return fmt.Errorf("Failed to lvs.Clear() - %v", err.Error())
	}
	var lvsServices []lvs.Service
	for _, svc := range services {
		lvsServices = append(lvsServices, svcToL(svc))
	}
	err = lvs.Restore(lvsServices)
	if err != nil {
		if tab != nil {
			tab.RenameChain("filter", "portal-old", "portal")
		}
		return fmt.Errorf("Failed to lvs.Restore() (Perhaps bad database entry?) - %v", err.Error())
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
				return fmt.Errorf("Failed to tab.Insert() - %v", err.Error())
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
	// why do we need to modify rules if we already updated backend with current rules?
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

// conversion functions
// takes a lvs.Server and converts it to a database.Server
func lToSrv(server lvs.Server) database.Server {
	srv := database.Server{Host: server.Host, Port: server.Port, Forwarder: server.Forwarder, Weight: server.Weight, UpperThreshold: server.UpperThreshold, LowerThreshold: server.LowerThreshold}
	srv.GenId()
	return srv
}
func lToSrvp(server *lvs.Server) *database.Server {
	srv := &database.Server{Host: server.Host, Port: server.Port, Forwarder: server.Forwarder, Weight: server.Weight, UpperThreshold: server.UpperThreshold, LowerThreshold: server.LowerThreshold}
	srv.GenId()
	return srv
}

// takes a lvs.Service and converts it to a database.Service
func lToSvc(service lvs.Service) database.Service {
	srvs := []database.Server{}
	for i, srv := range service.Servers {
		srvs = append(srvs, lToSrv(srv))
		srvs[i].GenId()
	}
	svc := database.Service{Host: service.Host, Port: service.Port, Type: service.Type, Scheduler: service.Scheduler, Persistence: service.Persistence, Netmask: service.Netmask, Servers: srvs}
	svc.GenId()
	return svc
}
func lToSvcp(service *lvs.Service) *database.Service {
	srvs := []database.Server{}
	for i, srv := range service.Servers {
		srvs = append(srvs, lToSrv(srv))
		srvs[i].GenId()
	}
	svc := &database.Service{Host: service.Host, Port: service.Port, Type: service.Type, Scheduler: service.Scheduler, Persistence: service.Persistence, Netmask: service.Netmask, Servers: srvs}
	svc.GenId()
	return svc
}

// takes a database.Server and converts it to an lvs.Server
func srvToL(server database.Server) lvs.Server {
	return lvs.Server{Host: server.Host, Port: server.Port, Forwarder: server.Forwarder, Weight: server.Weight, UpperThreshold: server.UpperThreshold, LowerThreshold: server.LowerThreshold}
}

// takes a database.Service and converts it to an lvs.Service
func svcToL(server database.Service) lvs.Service {
	srvs := []lvs.Server{}
	for _, srv := range server.Servers {
		srvs = append(srvs, srvToL(srv))
	}
	return lvs.Service{Host: server.Host, Port: server.Port, Type: server.Type, Scheduler: server.Scheduler, Persistence: server.Persistence, Netmask: server.Netmask, Servers: srvs}
}
