// +build windows darwin
// this fake lvs enables portal to compile for darwin/windows

package vipmgr

import (
	"github.com/nanopack/portal/core"
	"github.com/nanopack/portal/config"
)

type ip struct{}

func (self ip) Init() error {
	// allow to start up
	return nil
}
func (self ip) SetVip(vip core.Vip) error {
	config.Log.Warn("VIP functionality not fully supported on darwin|windows. Continuing anyways")
	return nil
}
func (self ip) DeleteVip(vip core.Vip) error {
	config.Log.Warn("VIP functionality not fully supported on darwin|windows. Continuing anyways")
	return nil
}
func (self ip) SetVips(vips []core.Vip) error {
	config.Log.Warn("VIP functionality not fully supported on darwin|windows. Continuing anyways")
	return nil
}
func (self ip) GetVips() ([]core.Vip, error) {
	config.Log.Warn("VIP functionality not fully supported on darwin|windows. Continuing anyways")
	return nil, nil
}
