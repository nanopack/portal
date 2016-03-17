package common

import (
	"fmt"
	"strings"

	"github.com/nanobox-io/nanobox-router"

	"github.com/nanopack/portal/balance"
	"github.com/nanopack/portal/core"
	"github.com/nanopack/portal/database"
	"github.com/nanopack/portal/routemgr"
)

func SetServices(services []core.Service) error {
	// in case of failure
	oldServices, err := database.GetServices()
	if err != nil {
		return err
	}

	// apply services to balancer
	err = balance.SetServices(services)
	if err != nil {
		return err
	}

	// save to backend
	err = database.SetServices(services)
	if err != nil {
		// undo balance action
		if uerr := balance.SetServices(oldServices); uerr != nil {
			err = fmt.Errorf("%v - %v", err.Error(), uerr.Error())
		}
		return err
	}
	return nil
}

func SetService(service *core.Service) error {
	// in case of failure
	oldServices, err := database.GetServices()
	if err != nil {
		return err
	}

	// apply to balancer
	err = balance.SetService(service)
	if err != nil {
		return err
	}

	// save to backend
	err = database.SetService(service)
	if err != nil {
		// undo balancer action
		if uerr := balance.SetServices(oldServices); uerr != nil {
			err = fmt.Errorf("%v - %v", err.Error(), uerr.Error())
		}
		return err
	}
	return nil
}

func DeleteService(svcId string) error {
	// in case of failure
	oldService, err := database.GetService(svcId)
	if err != nil {
		if !strings.Contains(err.Error(), "No Service Found") {
			return err
		}
	}

	// delete backend rule
	err = balance.DeleteService(svcId)
	if err != nil {
		return err
	}

	// remove from backend
	err = database.DeleteService(svcId)
	if err != nil {
		// undo balance action
		if oldService != nil {
			if uerr := balance.SetService(oldService); uerr != nil {
				err = fmt.Errorf("%v - %v", err.Error(), uerr.Error())
			}
		}
		return err
	}
	return nil
}

func SetServers(svcId string, servers []core.Server) error {
	// in case of failure
	oldService, err := database.GetService(svcId)
	if err != nil {
		return err
	}
	oldServers := oldService.Servers

	// implement in balancer
	err = balance.SetServers(svcId, servers)
	if err != nil {
		return err
	}

	// add to backend
	err = database.SetServers(svcId, servers)
	if err != nil {
		// undo balance action
		if uerr := balance.SetServers(svcId, oldServers); uerr != nil {
			err = fmt.Errorf("%v - %v", err.Error(), uerr.Error())
		}
		return err
	}
	return nil
}

func SetServer(svcId string, server *core.Server) error {
	// // idempotent additions (don't update server on post)
	// if srv, _ := database.GetServer(svcId, server.Id); srv != nil {
	//  return nil
	// }

	// apply to balancer
	err := balance.SetServer(svcId, server)
	if err != nil {
		return err
	}

	// save to backend
	err = database.SetServer(svcId, server)
	if err != nil {
		// undo balance action
		if uerr := balance.DeleteServer(svcId, server.Id); uerr != nil {
			err = fmt.Errorf("%v - %v", err.Error(), uerr.Error())
		}
		return err
	}
	return nil
}

func DeleteServer(svcId, srvId string) error {
	// in case of failure
	srv, err := database.GetServer(svcId, srvId)
	if err != nil {
		if !strings.Contains(err.Error(), "No Server Found") {
			return err
		}
	}

	// delete rule from balancer
	if err = balance.DeleteServer(svcId, srvId); err != nil {
		if !strings.Contains(err.Error(), "No Server Found") {
			return err
		}
	}

	// remove from backend
	if err = database.DeleteServer(svcId, srvId); err != nil && !strings.Contains(err.Error(), "No Server Found") {
		// undo balance action
		if uerr := balance.SetServer(svcId, srv); uerr != nil {
			err = fmt.Errorf("%v - %v", err.Error(), uerr.Error())
		}
		return err
	}
	return nil
}

func GetServices() ([]core.Service, error) {
	return database.GetServices()
}
func GetService(id string) (*core.Service, error) {
	return database.GetService(id)
}
func GetServer(svcId, srvId string) (*core.Server, error) {
	return database.GetServer(svcId, srvId)
}

func SetRoutes(routes []router.Route) error {
	// in case of failure
	oldRoutes, err := database.GetRoutes()
	if err != nil {
		return err
	}

	// apply routes to routemgr
	err = routemgr.SetRoutes(routes)
	if err != nil {
		return err
	}
	// save to backend
	err = database.SetRoutes(routes)
	if err != nil {
		// undo routemgr action
		if uerr := routemgr.SetRoutes(oldRoutes); uerr != nil {
			err = fmt.Errorf("%v - %v", err.Error(), uerr.Error())
		}
		return err
	}
	return nil
}

func SetRoute(route router.Route) error {
	// in case of failure
	oldRoutes, err := database.GetRoutes()
	if err != nil {
		return err
	}

	// apply to routemgr
	err = routemgr.SetRoute(route)
	if err != nil {
		return err
	}

	// save to backend
	err = database.SetRoute(route)
	if err != nil {
		// undo routemgr action
		if uerr := routemgr.SetRoutes(oldRoutes); uerr != nil {
			err = fmt.Errorf("%v - %v", err.Error(), uerr.Error())
		}
		return err
	}
	return nil
}

func DeleteRoute(route router.Route) error {
	// in case of failure
	oldRoutes, err := database.GetRoutes()
	if err != nil {
		return err
	}

	// apply to routemgr
	err = routemgr.DeleteRoute(route)
	if err != nil {
		return err
	}

	// save to backend
	err = database.DeleteRoute(route)
	if err != nil {
		// undo routemgr action
		if uerr := routemgr.SetRoutes(oldRoutes); uerr != nil {
			err = fmt.Errorf("%v - %v", err.Error(), uerr.Error())
		}
		return err
	}
	return nil
}

func GetRoutes() ([]router.Route, error) {
	return database.GetRoutes()
}
