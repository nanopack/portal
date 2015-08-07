// -*- mode: go; tab-width: 2; indent-tabs-mode: 1; st-rulers: [70] -*-
// vim: ts=4 sw=4 ft=lua noet
//--------------------------------------------------------------------
// @author Daniel Barney <daniel@nanobox.io>
// @copyright 2015, Pagoda Box Inc.
// @doc
//
// @end
// Created :   7 August 2015 by Daniel Barney <daniel@nanobox.io>
//--------------------------------------------------------------------
package ipvsadm

import (
	"errors"
	"os/exec"
)

type (
	Server struct {
		Id string
	}
	Vip struct {
		Id string
	}
)

func Check() error {
	cmd := exec.Command("which", "ipvsadm")
	if err := cmd.Run(); err != nil {
		return errors.New("unable to find the ipvsadm command on the system")
	}
	return nil
}

func ListVips() ([]Vip, error) {
	var vips = make([]Vip, 0)
	return vips, nil
}

func AddVip(vip Vip) error {
	return nil
}

func GetVip(vip *Vip) error {
	return nil
}

func DeleteVip(vip Vip) error {
	return nil
}

func ListServers() ([]Server, error) {
	var servers = make([]Server, 0)
	return servers, nil
}

func AddServer(server Server) error {
	return nil
}

func GetServer(server *Server) error {
	return nil
}

func EnableServer(server Server) error {
	return nil
}

func DeleteServer(server Server) error {
	return nil
}
