package database

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/nanobox-io/golang-lvs"
	"github.com/nanobox-io/golang-scribble"

	"github.com/nanopack/portal/config"
)

type (
	ScribbleDatabase struct {
		scribbleDb *scribble.Driver
	}
)

func key(service lvs.Service) string {
	return fmt.Sprintf("%s-%s-%d", service.Type, service.Host, service.Port)
}

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

func (s ScribbleDatabase) GetServices() ([]lvs.Service, error) {
	services := make([]lvs.Service, 0)
	values, err := s.scribbleDb.ReadAll("services")
	if err != nil {
		return services, err
	}
	for i := range values {
		var service lvs.Service
		json.Unmarshal([]byte(values[i]), &service)
		services = append(services, service)
	}
	return services, nil
}

func (s ScribbleDatabase) GetService(service lvs.Service) (lvs.Service, error) {
	real_service := lvs.Service{}
	err := s.scribbleDb.Read("services", key(service), &real_service)
	if err != nil {
		return real_service, err
	}
	return real_service, nil
}

func (s ScribbleDatabase) SetServices(services []lvs.Service) error {
	s.scribbleDb.Delete("services", "")
	for i := range services {
		err := s.scribbleDb.Write("services", key(services[i]), services[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (s ScribbleDatabase) SetService(service lvs.Service) error {
	return s.scribbleDb.Write("services", key(service), service)
}

func (s ScribbleDatabase) DeleteService(service lvs.Service) error {
	return s.scribbleDb.Delete("services", key(service))
}
