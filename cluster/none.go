package cluster

import (
	"github.com/nanopack/portal/core"
	"github.com/nanopack/portal/core/common"
	"github.com/nanopack/portal/database"
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

	// load vips
	vips, err := common.GetVips()
	if err != nil {
		return err
	}

	// apply vips
	err = common.SetVips(vips)
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
	err := common.SetServices(services)
	if err != nil {
		return err
	}
	if database.CentralStore {
		return database.SetServices(services)
	}
	return nil
}
func (n None) SetService(service *core.Service) error {
	err := common.SetService(service)
	if err != nil {
		return err
	}
	if database.CentralStore {
		return database.SetService(service)
	}
	return nil
}
func (n None) DeleteService(id string) error {
	err := common.DeleteService(id)
	if err != nil {
		return err
	}
	if database.CentralStore {
		return database.DeleteService(id)
	}
	return nil
}
func (n None) SetServers(svcId string, servers []core.Server) error {
	err := common.SetServers(svcId, servers)
	if err != nil {
		return err
	}
	if database.CentralStore {
		return database.SetServers(svcId, servers)
	}
	return nil
}
func (n None) SetServer(svcId string, server *core.Server) error {
	err := common.SetServer(svcId, server)
	if err != nil {
		return err
	}
	if database.CentralStore {
		return database.SetServer(svcId, server)
	}
	return nil
}
func (n None) DeleteServer(svcId, srvId string) error {
	err := common.DeleteServer(svcId, srvId)
	if err != nil {
		return err
	}
	if database.CentralStore {
		return database.DeleteServer(svcId, srvId)
	}
	return nil
}
func (n None) GetServer(svcId, srvId string) (*core.Server, error) {
	return common.GetServer(svcId, srvId)
}
func (n None) SetRoutes(routes []core.Route) error {
	err := common.SetRoutes(routes)
	if err != nil {
		return err
	}
	if database.CentralStore {
		return database.SetRoutes(routes)
	}
	return nil
}
func (n None) SetRoute(route core.Route) error {
	err := common.SetRoute(route)
	if err != nil {
		return err
	}
	if database.CentralStore {
		return database.SetRoute(route)
	}
	return nil
}
func (n None) DeleteRoute(route core.Route) error {
	err := common.DeleteRoute(route)
	if err != nil {
		return err
	}
	if database.CentralStore {
		return database.DeleteRoute(route)
	}
	return nil
}
func (n None) GetRoutes() ([]core.Route, error) {
	return common.GetRoutes()
}
func (n None) SetCerts(certs []core.CertBundle) error {
	err := common.SetCerts(certs)
	if err != nil {
		return err
	}
	if database.CentralStore {
		return database.SetCerts(certs)
	}
	return nil
}
func (n None) SetCert(cert core.CertBundle) error {
	err := common.SetCert(cert)
	if err != nil {
		return err
	}
	if database.CentralStore {
		return database.SetCert(cert)
	}
	return nil
}
func (n None) DeleteCert(cert core.CertBundle) error {
	err := common.DeleteCert(cert)
	if err != nil {
		return err
	}
	if database.CentralStore {
		return database.DeleteCert(cert)
	}
	return nil
}
func (n None) GetCerts() ([]core.CertBundle, error) {
	return common.GetCerts()
}
func (n None) SetVips(vips []core.Vip) error {
	err := common.SetVips(vips)
	if err != nil {
		return err
	}
	if database.CentralStore {
		return database.SetVips(vips)
	}
	return nil
}
func (n None) SetVip(vip core.Vip) error {
	err := common.SetVip(vip)
	if err != nil {
		return err
	}
	if database.CentralStore {
		return database.SetVip(vip)
	}
	return nil
}
func (n None) DeleteVip(vip core.Vip) error {
	err := common.DeleteVip(vip)
	if err != nil {
		return err
	}
	if database.CentralStore {
		return database.DeleteVip(vip)
	}
	return nil
}
func (n None) GetVips() ([]core.Vip, error) {
	return common.GetVips()
}
