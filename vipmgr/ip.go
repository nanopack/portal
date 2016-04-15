// ip.go contains logic to use `ip` to add vips

package vipmgr

import (
	"fmt"
	"os/exec"
	"sync"

	"github.com/nanopack/portal/config"
	"github.com/nanopack/portal/core"
)

type ip struct{}

var (
	// virtIps stores all registered Vip objects
	virtIps = []core.Vip{}
	// mutex ensures updates to vips are atomic
	mutex = sync.Mutex{}
)

func (self ip) Init() error {
	return exec.Command("which", "ip").Run()
}

////////////////////////////////////////////////////////////////////////////////
// VIPS
////////////////////////////////////////////////////////////////////////////////

// SetVip adds a vip to the host and updates the cached vips
func (self ip) SetVip(vip core.Vip) error {
	// for idempotency
	for i := range virtIps {
		// if ip or alias is already added...
		if virtIps[i].Ip == vip.Ip || virtIps[i].Alias == vip.Alias {
			// check if the vip exists...
			if virtIps[i].Ip == vip.Ip && virtIps[i].Interface == vip.Interface && virtIps[i].Alias == vip.Alias {
				// and be idempotent...
				return nil
			}
			// otherwise, return an error...
			return fmt.Errorf("Ip or alias already in use")
		}
	}

	// add vip to host
	err := exec.Command("ip", "addr", "add", vip.Ip, "dev", vip.Interface, "label", vip.Alias).Run()
	if err != nil {
		return fmt.Errorf("Failed to add vip '%v' - %v", vip.Ip, err)
	}

	// update vip cache
	mutex.Lock()
	virtIps = append(virtIps, vip)
	mutex.Unlock()
	config.Log.Trace("Vip added")

	return nil
}

// DeleteVip removes a vip from the host and updates the cached vips
func (self ip) DeleteVip(vip core.Vip) error {
	vips := virtIps
	// don't delete ips that we didn't add and be idempotent
	for i := range vips {
		// check if the vip exists...
		if vips[i].Ip == vip.Ip && vips[i].Interface == vip.Interface {
			// and go delete it...
			break
		}
		// otherwise, be idempotent and report it was deleted...
		return nil
	}

	// remove vip from host
	err := exec.Command("ip", "addr", "del", vip.Ip, "dev", vip.Interface).Run()
	if err != nil {
		return fmt.Errorf("Failed to remove vip '%v' - %v", vip.Ip, err)
	}

	// remove from cache
	for i := range vips {
		if vips[i].Ip == vip.Ip && vips[i].Interface == vip.Interface {
			vips = append(vips[:i], vips[i+1:]...)
			break
		}
	}

	// update vip cache
	mutex.Lock()
	virtIps = vips
	mutex.Unlock()
	config.Log.Trace("Vip removed")

	return nil
}

// SetVips removes all vips from the host and re-adds the new ones. It then
// updates the cached vips
func (self ip) SetVips(vips []core.Vip) error {
	oldVips := virtIps
	// remove old vips from system and cache (upon success)
	for i := range oldVips {
		err := self.DeleteVip(oldVips[i])
		if err != nil {
			// try rolling back
			config.Log.Trace("Trying to roll back for old vip...")
			err2 := self.SetVip(oldVips[i])
			return fmt.Errorf("Failed to remove old vip - %v %v", err, err2)
		}
		config.Log.Trace("Removed old vip '%v'", oldVips[i].Ip)
	}

	for i := range vips {
		err := self.SetVip(vips[i])
		if err != nil {
			// rather than attempting to roll back, avoid loop issues and just return the error
			return fmt.Errorf("Failed to add new vip - %v", err)
		}
		config.Log.Trace("Added new vip '%v'", vips[i].Ip)
	}

	mutex.Lock()
	virtIps = vips
	mutex.Unlock()
	config.Log.Trace("Vips updated")
	return nil
}

// GetVips returns the cached vips
func (self ip) GetVips() ([]core.Vip, error) {
	return virtIps, nil
}
