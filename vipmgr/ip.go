// +build linux
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
	// todo: readlock
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
	out, err := exec.Command("ip", "addr", "add", vip.Ip, "dev", vip.Interface, "label", vip.Alias).CombinedOutput()
	if err != nil {
		// if portal exited uncleanly and interface is still up, "change" addr
		out2, err2 := exec.Command("ip", "addr", "change", vip.Ip, "dev", vip.Interface, "label", vip.Alias).CombinedOutput()
		if err2 != nil {
			return fmt.Errorf("Failed to add vip '%s' - %s - %s", vip.Ip, out, out2)
		}
	}

	// update vip cache
	mutex.Lock()
	virtIps = append(virtIps, vip)
	mutex.Unlock()
	config.Log.Trace("Vip '%s' added", vip.Ip)

	// arp vip to neighbors, takes time goroutine
	go func() {
		out, err = exec.Command("arping", "-A", "-c", "10", "-I", vip.Interface, vip.Ip).CombinedOutput()
		if err != nil {
			// log rather than return
			config.Log.Error("Failed to arp vip '%s' - %s", vip.Ip, out)
		}
		config.Log.Trace("Arped vip - '%s'", vip.Ip)
	}()

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
			config.Log.Trace("Vip '%s' found to remove", vip.Ip)
			goto deleteIt
		}
	}
	// otherwise, be idempotent and report it was deleted...
	config.Log.Trace("Vip '%s' not found, reporting success", vip.Ip)
	return nil

deleteIt:
	// remove vip from host
	err := exec.Command("ip", "addr", "del", vip.Ip+"/32", "dev", vip.Interface).Run()
	if err != nil {
		return fmt.Errorf("Failed to remove vip '%s' - %s", vip.Ip, err)
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
	config.Log.Trace("Vip '%s' removed", vip.Ip)

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
			config.Log.Trace("Trying to roll back for old vip - %s...", err)
			err2 := self.SetVip(oldVips[i])
			return fmt.Errorf("Failed to remove old vip - %s %s", err, err2)
		}
		config.Log.Trace("Removed old vip '%s'", oldVips[i].Ip)
	}

	for i := range vips {
		err := self.SetVip(vips[i])
		if err != nil {
			// rather than attempting to roll back, avoid loop issues and just return the error
			return fmt.Errorf("Failed to add new vip - %s", err)
		}
		config.Log.Trace("Added new vip '%s'", vips[i].Ip)
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
