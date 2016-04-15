// +build windows darwin
// this fake lvs enables portal to compile for darwin/windows

package vipmgr

import (
	"fmt"

	"github.com/nanopack/portal/core"
)
type ip struct{}

func (self ip) Init() error {
	// allow to start up
	return nil
}
func (self ip) SetVip(vip core.Vip) error {
	return fmt.Errorf("Functionality not supported on darwin|windows")
}
func (self ip) DeleteVip(vip core.Vip) error {
	return fmt.Errorf("Functionality not supported on darwin|windows")
}
func (self ip) SetVips(vips []core.Vip) error {
	return fmt.Errorf("Functionality not supported on darwin|windows")
}
func (self ip) GetVips() ([]core.Vip, error) {
	return nil, fmt.Errorf("Functionality not supported on darwin|windows")
}
