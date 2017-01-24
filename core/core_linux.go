// +build linux

package core

import (
	"fmt"
	"os/exec"
	"regexp"
)

// Because it is more common to use an address alias vs an actual virtual
// interface, we need to use ifconfig and pass in the "interface" (or alias)
// to get the address
func (s *Service) GenHost() error {
	ifout, err := exec.Command("/sbin/ifconfig", s.Interface).Output()
	if err != nil {
		return fmt.Errorf("Failed to lookup address of '%s' - %s", s.Interface, err)
	}

	r, err := regexp.Compile(".*inet addr:([0-9]*.[0-9]*.[0-9]*.[0-9]*)")
	if err != nil {
		return fmt.Errorf("Regex compile fail - %s", err)
	}

	addr := r.FindStringSubmatch(fmt.Sprintf("%+v", string(ifout)))
	if len(addr) != 2 {
		return fmt.Errorf("No address found for '%s' - %s", s.Interface, err)
	}

	s.Host = addr[1]
	return nil
}
