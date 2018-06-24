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
	CentralStore   bool
	Backend        Storable
	NoServiceError = errors.New("No Service Found")
	NoServerError  = errors.New("No Server Found")
)

type Storable interface {
	core.Backender
	core.Proxyable
	core.Vipable
}

func Init() error {
	var err error
	var u *url.URL
	u, err = url.Parse(config.DatabaseConnection)
	if err != nil {
		return fmt.Errorf("Failed to parse db connection - %s", err)
	}
	switch u.Scheme {
	case "bolt", "boltdb":
		CentralStore = false
		Backend = &BoltDb{}
	case "postgres", "postgresql":
		CentralStore = true
		Backend = &PostgresDb{}
	case "scribble":
		fallthrough
	default:
		CentralStore = false
		Backend = &ScribbleDatabase{}
	}
	err = Backend.Init()
	if err != nil {
		return fmt.Errorf("Failed to init database - %s", err)
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

func SetVips(vips []core.Vip) error {
	return Backend.SetVips(vips)
}

func SetVip(vip core.Vip) error {
	return Backend.SetVip(vip)
}

func DeleteVip(vip core.Vip) error {
	return Backend.DeleteVip(vip)
}

func GetVips() ([]core.Vip, error) {
	return Backend.GetVips()
}
