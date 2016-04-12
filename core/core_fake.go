// +build windows darwin
// this fake core enables portal to compile for darwin/windows

package core

import (
	"fmt"
	"net"
	"strings"
)

func (s *Service) GenHost() error {
	iface, err := net.InterfaceByName(s.Interface)
	if err != nil {
		return fmt.Errorf("No interface found '%v' - %v", s.Interface, err)
	}
	addrs, err := iface.Addrs()
	if err != nil {
		return fmt.Errorf("Failed to get address for '%v' - %v", s.Interface, err)
	}
	if len(addrs) < 1 {
		s.Host = "127.0.0.1"
		return nil
	}
	s.Host = strings.Split(addrs[0].String(), "/")[0]
	return nil
}
