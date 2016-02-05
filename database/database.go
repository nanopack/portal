package database

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/nanopack/portal/config"
	"github.com/nanopack/portal/core"
)

var (
	Backend        core.Backender
	NoServiceError = errors.New("No Service Found")
	NoServerError  = errors.New("No Server Found")
)

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
