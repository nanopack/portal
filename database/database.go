package database

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/nanobox-io/nanobox-router"

	"github.com/nanopack/portal/config"
	"github.com/nanopack/portal/core"
	"github.com/nanopack/portal/routemgr"
)

var (
	Backend        Storable
	NoServiceError = errors.New("No Service Found")
	NoServerError  = errors.New("No Server Found")
)

type Storable interface {
	core.Backender
	routemgr.Routable
}

func Init() error {
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
		Backend = &ScribbleDatabase{}
	}
	err = Backend.Init()
	if err != nil {
		Backend = nil
	}
	return nil
}

func GetServices() ([]core.Service, error) {
	return Backend.GetServices()
}

func GetService(id string) (*core.Service, error) {
	return Backend.GetService(id)
}

func SetServices(services []core.Service) error {
	return Backend.SetServices(services)
}

func SetService(service *core.Service) error {
	return Backend.SetService(service)
}

func DeleteService(id string) error {
	return Backend.DeleteService(id)
}

func SetServers(svcId string, servers []core.Server) error {
	return Backend.SetServers(svcId, servers)
}

func SetServer(svcId string, server *core.Server) error {
	return Backend.SetServer(svcId, server)
}

func DeleteServer(svcId, srvId string) error {
	return Backend.DeleteServer(svcId, srvId)
}

func GetServer(svcId, srvId string) (*core.Server, error) {
	return Backend.GetServer(svcId, srvId)
}

func SetRoutes(routes []router.Route) error {
	return Backend.SetRoutes(routes)
}

func SetRoute(route router.Route) error {
	return Backend.SetRoute(route)
}

func DeleteRoute(route router.Route) error {
	return Backend.DeleteRoute(route)
}

func GetRoutes() ([]router.Route, error) {
	return Backend.GetRoutes()
}
