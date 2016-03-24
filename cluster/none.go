package cluster

import (
	"github.com/nanopack/portal/core"
	"github.com/nanopack/portal/core/common"
)

type (
	None struct{}
)

func (n None) UnInit() error {
	return nil
}
func (n None) Init() error {
	// load services
	services, err := common.GetServices()
	if err != nil {
		return err
	}
	// apply services
	err = common.SetServices(services)
	if err != nil {
		return err
	}

	// load routes
	routes, err := common.GetRoutes()
	if err != nil {
		return err
	}
	// apply routes
	err = common.SetRoutes(routes)
	if err != nil {
		return err
	}

	// load certs
	certs, err := common.GetCerts()
	if err != nil {
		return err
	}
	// apply certs
	err = common.SetCerts(certs)
	if err != nil {
		return err
	}
	return nil
}
func (n None) GetServices() ([]core.Service, error) {
	return common.GetServices()
}
func (n None) GetService(id string) (*core.Service, error) {
	return common.GetService(id)
}
func (n None) SetServices(services []core.Service) error {
	return common.SetServices(services)
}
func (n None) SetService(service *core.Service) error {
	return common.SetService(service)
}
func (n None) DeleteService(id string) error {
	return common.DeleteService(id)
}
func (n None) SetServers(svcId string, servers []core.Server) error {
	return common.SetServers(svcId, servers)
}
func (n None) SetServer(svcId string, server *core.Server) error {
	return common.SetServer(svcId, server)
}
func (n None) DeleteServer(svcId, srvId string) error {
	return common.DeleteServer(svcId, srvId)
}
func (n None) GetServer(svcId, srvId string) (*core.Server, error) {
	return common.GetServer(svcId, srvId)
}
func (n None) SetRoutes(routes []core.Route) error {
	return common.SetRoutes(routes)
}
func (n None) SetRoute(route core.Route) error {
	return common.SetRoute(route)
}
func (n None) DeleteRoute(route core.Route) error {
	return common.DeleteRoute(route)
}
func (n None) GetRoutes() ([]core.Route, error) {
	return common.GetRoutes()
}
func (n None) SetCerts(certs []core.CertBundle) error {
	return common.SetCerts(certs)
}
func (n None) SetCert(cert core.CertBundle) error {
	return common.SetCert(cert)
}
func (n None) DeleteCert(cert core.CertBundle) error {
	return common.DeleteCert(cert)
}
func (n None) GetCerts() ([]core.CertBundle, error) {
	return common.GetCerts()
}
