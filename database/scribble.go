package database

import (
	"encoding/json"
	"net/url"
	"strings"

	"github.com/nanobox-io/golang-scribble"

	"github.com/nanopack/portal/config"
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

func (s ScribbleDatabase) GetServices() ([]Service, error) {
	services := make([]Service, 0)
	values, err := s.scribbleDb.ReadAll("services")
	if err != nil {
		// don't return an error here if the stat fails
		if strings.Contains(err.Error(), "no such file or directory") {
			config.Log.Info("File 'services[.json]' could not be found, no services imported")
			err = nil
		}
		return services, err
	}
	for i := range values {
		var service Service
		json.Unmarshal([]byte(values[i]), &service)
		services = append(services, service)
	}
	return services, nil
}

func (s ScribbleDatabase) GetService(id string) (Service, error) {
	service := Service{}
	err := s.scribbleDb.Read("services", id, &service)
	if err != nil {
		// more generic error? no service error?
		return service, err
	}
	return service, nil
}

func (s ScribbleDatabase) SetServices(services []Service) error {
	s.scribbleDb.Delete("services", "")
	for i := range services {
		err := s.scribbleDb.Write("services", key(services[i]), services[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (s ScribbleDatabase) SetService(service *Service) error {
	return s.scribbleDb.Write("services", key(*service), *service)
}

func (s ScribbleDatabase) DeleteService(id string) error {
	return s.scribbleDb.Delete("services", id)
}

func (s ScribbleDatabase) SetServer(svcId string, server *Server) error {
	service, err := s.GetService(svcId)
	if err != nil {
		return err
	}
	service.Servers = append(service.Servers, *server)

	return s.scribbleDb.Write("services", key(service), service)
}

func (s ScribbleDatabase) SetServers(svcId string, servers []Server) error {
	service, err := s.GetService(svcId)
	if err != nil {
		return err
	}

	// pretty simple, reset all servers
	service.Servers = servers

	return s.scribbleDb.Write("services", key(service), service)
}

func (s ScribbleDatabase) DeleteServer(svcId, srvId string) error {
	service, err := s.GetService(svcId)
	if err != nil {
		// // if read was successful, but found no
		// if strings.Contains(err.Error(), "found") {
		// 	return nil
		// }
		return nil
	}
	for _, srv := range service.Servers {
		if srv.Id == srvId {
			// todo: empty or a = append(a[:i], a[i+1:]...)
			srv = Server{}
		}
	}

	return s.scribbleDb.Write("services", key(service), service)
}
