// +build windows darwin
// this fake lvs enables portal to compile for darwin/windows

package balance

import "github.com/nanopack/portal/core"

type (
	Lvs struct {
	}
)

func (l Lvs) Init() error {
	return nil
}

func (l Lvs) GetServer(svcId, srvId string) (*core.Server, error) {
	return nil, nil
}

func (l Lvs) SetServer(svcId string, server *core.Server) error {
	return nil
}

func (l Lvs) DeleteServer(svcId, srvId string) error {
	return nil
}

func (l Lvs) SetServers(svcId string, servers []core.Server) error {
	return nil
}

func (l Lvs) GetService(id string) (*core.Service, error) {
	return nil, nil
}

func (l Lvs) SetService(service *core.Service) error {
	return nil
}

func (l Lvs) DeleteService(id string) error {
	return nil
}

func (l Lvs) GetServices() ([]core.Service, error) {
	return nil, nil
}

func (l Lvs) SetServices(services []core.Service) error {
	return nil
}
