package database

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/nanobox-io/golang-scribble"
	"github.com/nanobox-io/nanobox-router"

	"github.com/nanopack/portal/config"
	"github.com/nanopack/portal/core"
)

type (
	ScribbleDatabase struct {
		scribbleDb *scribble.Driver
	}
)

func (s *ScribbleDatabase) Init() error {
	u, err := url.Parse(config.DatabaseConnection)
	if err != nil {
		return err
	}
	dir := u.Path
	db, err := scribble.New(dir, nil)
	if err != nil {
		return err
	}

	s.scribbleDb = db
	return nil
}

func (s ScribbleDatabase) GetServices() ([]core.Service, error) {
	services := make([]core.Service, 0, 0)
	values, err := s.scribbleDb.ReadAll("services")
	if err != nil {
		if strings.Contains(err.Error(), "no such file or directory") {
			// if error is about a missing db, return empty array
			return services, nil
		}
		return nil, err
	}
	for i := range values {
		var service core.Service
		if err = json.Unmarshal([]byte(values[i]), &service); err != nil {
			return nil, fmt.Errorf("Bad JSON syntax received in body")
		}
		services = append(services, service)
	}
	return services, nil
}

func (s ScribbleDatabase) GetService(id string) (*core.Service, error) {
	service := core.Service{}
	err := s.scribbleDb.Read("services", id, &service)
	if err != nil {
		if strings.Contains(err.Error(), "no such file or directory") {
			err = NoServiceError
		}
		return nil, err
	}
	return &service, nil
}

func (s ScribbleDatabase) SetServices(services []core.Service) error {
	s.scribbleDb.Delete("services", "")
	for i := range services {
		err := s.scribbleDb.Write("services", services[i].Id, services[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (s ScribbleDatabase) SetService(service *core.Service) error {
	return s.scribbleDb.Write("services", service.Id, *service)
}

func (s ScribbleDatabase) DeleteService(id string) error {
	err := s.scribbleDb.Delete("services", id)
	if err != nil {
		if strings.Contains(err.Error(), "Unable to find") {
			err = nil
		}
		return err
	}
	return nil
}

func (s ScribbleDatabase) SetServer(svcId string, server *core.Server) error {
	service, err := s.GetService(svcId)
	if err != nil {
		return err
	}
	for _, srv := range service.Servers {
		if srv.Id == server.Id {
			// if server already exists, don't duplicate it
			return nil
		}
	}
	service.Servers = append(service.Servers, *server)

	return s.scribbleDb.Write("services", service.Id, service)
}

func (s ScribbleDatabase) SetServers(svcId string, servers []core.Server) error {
	service, err := s.GetService(svcId)
	if err != nil {
		return err
	}

	// pretty simple, reset all servers
	service.Servers = servers

	return s.scribbleDb.Write("services", service.Id, service)
}

func (s ScribbleDatabase) DeleteServer(svcId, srvId string) error {
	service, err := s.GetService(svcId)
	if err != nil {
		if strings.Contains(err.Error(), "Unable to find") {
			err = nil
		}
		return err
	}
	config.Log.Trace("Deleting %v from %v", srvId, svcId)
checkRemove:
	for i, srv := range service.Servers {
		if srv.Id == srvId {
			service.Servers = append(service.Servers[:i], service.Servers[i+1:]...)
			goto checkRemove // prevents 'slice bounds out of range' panic
		}
	}

	return s.scribbleDb.Write("services", service.Id, service)
}

func (s ScribbleDatabase) GetServer(svcId, srvId string) (*core.Server, error) {
	service := core.Service{}
	err := s.scribbleDb.Read("services", svcId, &service)
	if err != nil {
		return nil, err
	}

	for _, srv := range service.Servers {
		if srv.Id == srvId {
			return &srv, nil
		}
	}

	return nil, NoServerError
}

////////////////////////////////////////////////////////////////////////////////
// ROUTES
////////////////////////////////////////////////////////////////////////////////

func (s ScribbleDatabase) GetRoutes() ([]router.Route, error) {
	routes := make([]router.Route, 0, 0)
	values, err := s.scribbleDb.ReadAll("routes")
	if err != nil {
		if strings.Contains(err.Error(), "no such file or directory") {
			// if error is about a missing db, return empty array
			return routes, nil
		}
		return nil, err
	}
	for i := range values {
		var route router.Route
		if err = json.Unmarshal([]byte(values[i]), &route); err != nil {
			return nil, fmt.Errorf("Bad JSON syntax stored in db")
		}
		routes = append(routes, route)
	}
	return routes, nil
}

func (s ScribbleDatabase) SetRoutes(routes []router.Route) error {
	s.scribbleDb.Delete("routes", "")
	for i := range routes {
		// unique (as much as what we keep) key to store route by
		ukey := fmt.Sprintf("%v-%v%v", strings.Replace(routes[i].SubDomain, ".", "-", -1), strings.Replace(routes[i].Domain, ".", "-", -1), strings.Replace(routes[i].Path, "/", "_", -1))
		err := s.scribbleDb.Write("routes", ukey, routes[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (s ScribbleDatabase) SetRoute(route router.Route) error {
	routes, err := s.GetRoutes()
	if err != nil {
		return err
	}
	// for idempotency
	for i := 0; i < len(routes); i++ {
		if routes[i].SubDomain == route.SubDomain && routes[i].Domain == route.Domain && routes[i].Path == route.Path {
			return nil
		}
	}

	routes = append(routes, route)
	return s.SetRoutes(routes)
}

func (s ScribbleDatabase) DeleteRoute(route router.Route) error {
	routes, err := s.GetRoutes()
	if err != nil {
		return err
	}
	for i := 0; i < len(routes); i++ {
		if routes[i].SubDomain == route.SubDomain && routes[i].Domain == route.Domain && routes[i].Path == route.Path {
			routes = append(routes[:i], routes[i+1:]...)
			break
		}
	}
	return s.SetRoutes(routes)
}
