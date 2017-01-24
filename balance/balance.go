// +build linux darwin

// balance handles the load balancing portion of portal.
package balance

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/coreos/go-iptables/iptables"

	"github.com/nanopack/portal/config"
	"github.com/nanopack/portal/core"
)

var (
	Balancer       core.Backender
	NoServiceError = errors.New("No Service Found")
	NoServerError  = errors.New("No Server Found")

	tab *iptables.IPTables
)

func Init() error {
	if config.JustProxy {
		Balancer = nil
		return nil
	}

	// decide which balancer to use
	switch config.Balancer {
	case "lvs":
		Balancer = &Lvs{}
	case "nginx":
		Balancer = &Nginx{}
	default:
		Balancer = &Lvs{} // faster
	}

	var err error
	tab, err = iptables.New()
	if err != nil {
		tab = nil
	}
	// don't break if we can't use iptables
	if _, err = tab.List("filter", "INPUT"); err != nil {
		config.Log.Error("Could not use iptables, continuing without - %s", err)
		tab = nil
	}
	if tab != nil {
		tab.Delete("filter", "INPUT", "-j", "portal")
		tab.ClearChain("filter", "portal")
		tab.DeleteChain("filter", "portal")
		err = tab.NewChain("filter", "portal")
		if err != nil {
			return fmt.Errorf("Failed to create new chain - %s", err)
		}
		err = tab.AppendUnique("filter", "portal", "-j", "RETURN")
		if err != nil {
			return fmt.Errorf("Failed to append to portal chain - %s", err)
		}
		err = tab.AppendUnique("filter", "INPUT", "-j", "portal")
		if err != nil {
			return fmt.Errorf("Failed to append to INPUT chain - %s", err)
		}

		// Allow router through by default (ports 80/443)
		err = tab.Insert("filter", "portal", 1, "-p", "tcp", "--dport", "80", "-j", "ACCEPT")
		if err != nil {
			return err
		}
		err = tab.Insert("filter", "portal", 1, "-p", "tcp", "--dport", "443", "-j", "ACCEPT")
		if err != nil {
			return err
		}
	}

	return Balancer.Init()
}

func GetServices() ([]core.Service, error) {
	if Balancer == nil {
		return nil, nil
	}
	return Balancer.GetServices()
}

func GetService(id string) (*core.Service, error) {
	if Balancer == nil {
		return nil, nil
	}
	return Balancer.GetService(id)
}

func SetServices(services []core.Service) error {
	if Balancer == nil {
		return nil
	}
	if tab != nil {
		tab.RenameChain("filter", "portal", "portal-old")
	}
	err := Balancer.SetServices(services)
	if err != nil && tab != nil {
		tab.RenameChain("filter", "portal-old", "portal")
	}
	if err == nil && tab != nil {
		cleanup := func(error) error {
			tab.ClearChain("filter", "portal")
			tab.DeleteChain("filter", "portal")
			tab.RenameChain("filter", "portal-old", "portal")
			return fmt.Errorf("Failed to tab.Insert() - %s", err)
		}

		tab.NewChain("filter", "portal")
		tab.ClearChain("filter", "portal")
		tab.AppendUnique("filter", "portal", "-j", "RETURN")

		// rules for all services
		for i := range services {
			err := tab.Insert("filter", "portal", 1, "-p", services[i].Type, "-d", services[i].Host, "--dport", fmt.Sprintf("%d", services[i].Port), "-j", "ACCEPT")
			if err != nil {
				return cleanup(err)
			}
		}

		// Allow router through by default (ports 80/443)
		err = tab.Insert("filter", "portal", 1, "-p", "tcp", "--dport", "80", "-j", "ACCEPT")
		if err != nil {
			return cleanup(err)
		}
		err = tab.Insert("filter", "portal", 1, "-p", "tcp", "--dport", "443", "-j", "ACCEPT")
		if err != nil {
			return cleanup(err)
		}

		tab.AppendUnique("filter", "INPUT", "-j", "portal")
		tab.Delete("filter", "INPUT", "-j", "portal-old")
		tab.ClearChain("filter", "portal-old")
		tab.DeleteChain("filter", "portal-old")
	}

	return err
}

func SetService(service *core.Service) error {
	if Balancer == nil {
		return nil
	}

	err := Balancer.SetService(service)
	// update iptables rules
	if err == nil && tab != nil {
		errTab := tab.Insert("filter", "portal", 1, "-p", service.Type, "-d", service.Host, "--dport", fmt.Sprintf("%d", service.Port), "-j", "ACCEPT")
		if errTab != nil {
			return err
		}
	}
	return err
}

func DeleteService(id string) error {
	if Balancer == nil {
		return nil
	}
	service, err := parseSvc(id)
	if err != nil {
		return err
	}

	err = Balancer.DeleteService(id)
	if err == nil && tab != nil {
		// update iptables rules
		errTab := tab.Delete("filter", "portal", "-p", service.Type, "-d", service.Host, "--dport", fmt.Sprintf("%d", service.Port), "-j", "ACCEPT")
		if errTab != nil {
			return err
		}
	}

	return err
}

func SetServers(svcId string, servers []core.Server) error {
	if Balancer == nil {
		return nil
	}
	return Balancer.SetServers(svcId, servers)
}

func SetServer(svcId string, server *core.Server) error {
	if Balancer == nil {
		return nil
	}
	return Balancer.SetServer(svcId, server)
}

func DeleteServer(svcId, srvId string) error {
	if Balancer == nil {
		return nil
	}
	return Balancer.DeleteServer(svcId, srvId)
}

func GetServer(svcId, srvId string) (*core.Server, error) {
	if Balancer == nil {
		return nil, nil
	}
	return Balancer.GetServer(svcId, srvId)
}

func parseSvc(serviceId string) (*core.Service, error) {
	if Balancer == nil {
		return nil, nil
	}
	s := strings.Replace(serviceId, "_", ".", -1)
	svc := strings.Split(s, "-")
	if len(svc) != 3 {
		return nil, NoServiceError
	}
	p, _ := strconv.Atoi(svc[2])
	return &core.Service{Type: svc[0], Host: svc[1], Port: p}, nil
}

func parseSrv(serverId string) (*core.Server, error) {
	if Balancer == nil {
		return nil, nil
	}
	s := strings.Replace(serverId, "_", ".", -1)
	srv := strings.Split(s, "-")
	if len(srv) != 2 {
		return nil, NoServerError
	}
	p, _ := strconv.Atoi(srv[1])
	return &core.Server{Host: srv[0], Port: p}, nil
}
