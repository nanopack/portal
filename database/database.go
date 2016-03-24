// database handles portal's persistant storage.
package database

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/nanopack/portal/config"
	"github.com/nanopack/portal/core"
)

var (
	Backend        Storable
	NoServiceError = errors.New("No Service Found")
	NoServerError  = errors.New("No Server Found")
)

type Storable interface {
	core.Backender
	core.Proxyable
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

func SetRoutes(routes []core.Route) error {
	return Backend.SetRoutes(routes)
}

func SetRoute(route core.Route) error {
	return Backend.SetRoute(route)
}

func DeleteRoute(route core.Route) error {
	return Backend.DeleteRoute(route)
}

func GetRoutes() ([]core.Route, error) {
	return Backend.GetRoutes()
}

func SetCerts(certs []core.CertBundle) error {
	return Backend.SetCerts(certs)
}

func SetCert(cert core.CertBundle) error {
	return Backend.SetCert(cert)
}

func DeleteCert(cert core.CertBundle) error {
	return Backend.DeleteCert(cert)
}

func GetCerts() ([]core.CertBundle, error) {
	return Backend.GetCerts()
}
