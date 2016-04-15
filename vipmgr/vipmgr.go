// vipmgr handles the adding of 'vips'
package vipmgr

import "github.com/nanopack/portal/core"

type vipable interface {
	Init() error
	core.Vipable
}

var (
	Vip vipable
)

func Init() error {
	Vip = &ip{}
	return Vip.Init()
}

func SetVip(vip core.Vip) error {
	return Vip.SetVip(vip)
}

func DeleteVip(vip core.Vip) error {
	return Vip.DeleteVip(vip)
}

func SetVips(vips []core.Vip) error {
	return Vip.SetVips(vips)
}

func GetVips() ([]core.Vip, error) {
	return Vip.GetVips()
}
