// +build linux

package balance

import (
	"fmt"
	"sync"

	"github.com/coreos/go-iptables/iptables"
	"github.com/nanobox-io/golang-lvs"

	"github.com/nanopack/portal/core"
)

var (
	ipvsLock = &sync.RWMutex{}
	tab      *iptables.IPTables
)

type (
	Lvs struct {
	}
)

func (l *Lvs) Init() error {
	var err error
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
			return fmt.Errorf("Failed to create new chain - %v", err)
		}
		err = tab.AppendUnique("filter", "portal", "-j", "RETURN")
		if err != nil {
			return fmt.Errorf("Failed to append to portal chain - %v", err)
		}
		err = tab.AppendUnique("filter", "INPUT", "-j", "portal")
		if err != nil {
			return fmt.Errorf("Failed to append to INPUT chain - %v", err)
		}
	}
	return nil
}

// GetServer
func (l *Lvs) GetServer(svcId, srvId string) (*core.Server, error) {
	// break up ids in order to create dummy lvs.Service
	var err error
	service, err := parseSvc(svcId)
	if err != nil {
		return nil, err
	}

	server, err := parseSrv(srvId)
	if err != nil {
		return nil, err
	}

	ipvsLock.RLock()
	defer ipvsLock.RUnlock()
	s := lvs.DefaultIpvs.FindService(service.Type, service.Host, service.Port)
	if s == nil {
		return nil, NoServiceError
	}

	srv := s.FindServer(server.Host, server.Port)
	if srv == nil {
		return nil, NoServerError
	}
	return lToSrvp(srv), nil
}

// SetServer
func (l *Lvs) SetServer(svcId string, server *core.Server) error {
	service, err := l.GetService(svcId)
	if err != nil {
		return NoServiceError
	}
	lvsServer := srvToL(*server)

	ipvsLock.Lock()
	defer ipvsLock.Unlock()

	// add to lvs
	s := lvs.DefaultIpvs.FindService(service.Type, service.Host, service.Port)
	if s == nil {
		return NoServiceError
	}

	err = s.AddServer(lvsServer)
	if err != nil {
		return err
	}
	return nil
}

// DeleteServer
func (l *Lvs) DeleteServer(svcId, srvId string) error {
	service, err := parseSvc(svcId)
	if err != nil {
		// if service not valid, 'delete' successful
		return nil
	}

	server, err := parseSrv(srvId)
	if err != nil {
		// if invalid servername, 'delete' successful
		return nil
	}

	ipvsLock.Lock()
	defer ipvsLock.Unlock()
	// remove from lvs
	s := lvs.DefaultIpvs.FindService(service.Type, service.Host, service.Port)
	if s == nil {
		return nil
	}

	// if ipvsadm remove fails, should return error
	err = s.RemoveServer(server.Host, server.Port)
	if err != nil {
		return err
	}

	return nil
}

// SetServers
func (l *Lvs) SetServers(svcId string, servers []core.Server) error {
	service, err := l.GetService(svcId)
	if err != nil {
		return NoServiceError // was there a reason this was nil?
	}

	lvsServers := []lvs.Server{}
	for _, srv := range servers {
		lvsServers = append(lvsServers, srvToL(srv))
	}

	ipvsLock.Lock()
	defer ipvsLock.Unlock()
	// add to lvs
	s := lvs.DefaultIpvs.FindService(service.Type, service.Host, service.Port)
	if s == nil {
		return NoServiceError
	}

	for _, isrv := range s.Servers {
		if err = s.RemoveServer(isrv.Host, isrv.Port); err != nil {
			return fmt.Errorf("[ipvsadm] Failed to remove server - %v:%v; %v", isrv.Host, isrv.Port, err.Error())
		}
	}
	for _, lsrv := range lvsServers {
		if err = s.AddServer(lsrv); err != nil {
			return fmt.Errorf("[ipvsadm] Failed to add server - %v:%v; %v", lsrv.Host, lsrv.Port, err.Error())
		}
	}
	return nil
}

// GetService
func (l *Lvs) GetService(id string) (*core.Service, error) {
	service, err := parseSvc(id)
	if err != nil {
		return nil, err
	}

	ipvsLock.RLock()
	defer ipvsLock.RUnlock()

	svc := lvs.DefaultIpvs.FindService(service.Type, service.Host, service.Port)
	if svc == nil {
		return nil, NoServiceError
	}

	return lToSvcp(svc), nil
}

// SetService
func (l *Lvs) SetService(service *core.Service) error {
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

	// add servers, if any
	if len(lvsService.Servers) != 0 {
		s := lvs.DefaultIpvs.FindService(service.Type, service.Host, service.Port)
		if s == nil {
			return fmt.Errorf("Balancer failed to set service")
		}
		for _, srv := range lvsService.Servers {
			err = s.AddServer(srv)
			if err != nil {
				return err
			}
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

	ipvsLock.Lock()
	defer ipvsLock.Unlock()
	svc := lvs.DefaultIpvs.FindService(service.Type, service.Host, service.Port)
	if svc == nil {
		// if not exist, 'delete' successful
		return nil
	}

	// remove from lvs
	err = lvs.DefaultIpvs.RemoveService(service.Type, service.Host, service.Port)
	if err != nil {
		return err
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
// doesn't need to be a pointer method because it doesn't modify original object
func (l *Lvs) GetServices() ([]core.Service, error) {
	ipvsLock.RLock()
	defer ipvsLock.RUnlock()
	svcs := make([]core.Service, 0, 0)
	for _, svc := range lvs.DefaultIpvs.Services {
		svcs = append(svcs, lToSvc(svc))
	}
	return svcs, nil
}

// SetServices
// used also to sync to core (`SetServices(core.GetServices())`)
func (l *Lvs) SetServices(services []core.Service) error {
	var lvsServices []lvs.Service
	for _, svc := range services {
		lvsServices = append(lvsServices, svcToL(svc))
	}
	ipvsLock.Lock()
	defer ipvsLock.Unlock()
	if tab != nil {
		tab.RenameChain("filter", "portal", "portal-old")
	}
	err := lvs.Clear()
	if err != nil {
		if tab != nil {
			tab.RenameChain("filter", "portal-old", "portal")
		}
		return fmt.Errorf("Failed to lvs.Clear() - %v", err.Error())
	}
	err = lvs.Restore(lvsServices)
	if err != nil {
		if tab != nil {
			tab.RenameChain("filter", "portal-old", "portal")
		}
		return fmt.Errorf("Failed to lvs.Restore() - %v", err.Error())
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

// Sync - takes applies ipvsadm rules and save them to lvs.DefaultIpvs.Services
// which should already have the same information
// func (l *Lvs) Sync() error {
func Sync() error {
	// why do we need to modify rules if we already updated backend with current rules?
	ipvsLock.Lock()
	defer ipvsLock.Unlock()
	if tab != nil {
		tab.RenameChain("filter", "portal", "portal-old")
	}
	// save reads the applied ipvsadm rules from the host and saves them as i.Services
	err := lvs.Save()
	if err != nil {
		if tab != nil {
			tab.RenameChain("filter", "portal-old", "portal")
		}
		return fmt.Errorf("Failed to lvs.Save() - %v", err.Error())
	}

	lvsServices := lvs.DefaultIpvs.Services

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

// conversion functions
// takes a lvs.Server and converts it to a core.Server
func lToSrv(server lvs.Server) core.Server {
	srv := core.Server{Host: server.Host, Port: server.Port, Forwarder: server.Forwarder, Weight: server.Weight, UpperThreshold: server.UpperThreshold, LowerThreshold: server.LowerThreshold}
	srv.GenId()
	return srv
}

func lToSrvp(server *lvs.Server) *core.Server {
	srv := &core.Server{Host: server.Host, Port: server.Port, Forwarder: server.Forwarder, Weight: server.Weight, UpperThreshold: server.UpperThreshold, LowerThreshold: server.LowerThreshold}
	srv.GenId()
	return srv
}

// takes a lvs.Service and converts it to a core.Service
func lToSvc(service lvs.Service) core.Service {
	srvs := []core.Server{}
	for i, srv := range service.Servers {
		srvs = append(srvs, lToSrv(srv))
		srvs[i].GenId()
	}
	svc := core.Service{Host: service.Host, Port: service.Port, Type: service.Type, Scheduler: service.Scheduler, Persistence: service.Persistence, Netmask: service.Netmask, Servers: srvs}
	svc.GenId()
	return svc
}

func lToSvcp(service *lvs.Service) *core.Service {
	srvs := []core.Server{}
	for i, srv := range service.Servers {
		srvs = append(srvs, lToSrv(srv))
		srvs[i].GenId()
	}
	svc := &core.Service{Host: service.Host, Port: service.Port, Type: service.Type, Scheduler: service.Scheduler, Persistence: service.Persistence, Netmask: service.Netmask, Servers: srvs}
	svc.GenId()
	return svc
}

// takes a core.Server and converts it to an lvs.Server
func srvToL(server core.Server) lvs.Server {
	return lvs.Server{Host: server.Host, Port: server.Port, Forwarder: server.Forwarder, Weight: server.Weight, UpperThreshold: server.UpperThreshold, LowerThreshold: server.LowerThreshold}
}

// takes a core.Service and converts it to an lvs.Service
func svcToL(server core.Service) lvs.Service {
	srvs := []lvs.Server{}
	for _, srv := range server.Servers {
		srvs = append(srvs, srvToL(srv))
	}
	return lvs.Service{Host: server.Host, Port: server.Port, Type: server.Type, Scheduler: server.Scheduler, Persistence: server.Persistence, Netmask: server.Netmask, Servers: srvs}
}
