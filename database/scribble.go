package database

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/nanobox-io/golang-scribble"

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
	config.Log.Trace("Got service %v", service.Id)
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
	config.Log.Trace("Deleting %v from %v", srvId, svcId)
	if err != nil {
		if strings.Contains(err.Error(), "Unable to find") {
			err = nil
		}
		return err
	}
	for i, srv := range service.Servers {
		if srv.Id == srvId {
			service.Servers = append(service.Servers[:i], service.Servers[i+1:]...)
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
