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
