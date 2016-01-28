package database

import (
	"encoding/json"
	"fmt"
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

func key(service Service) string {
	return fmt.Sprintf("%v-%v-%d", service.Type, strings.Replace(service.Host, ".", "_", -1), service.Port)
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
	real_service := Service{}
	err := s.scribbleDb.Read("services", id, &real_service)
	if err != nil {
		return real_service, err
	}
	return real_service, nil
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

func (s ScribbleDatabase) SetService(service Service) error {
	return s.scribbleDb.Write("services", key(service), service)
}

func (s ScribbleDatabase) DeleteService(id string) error {
	return s.scribbleDb.Delete("services", id)
}
