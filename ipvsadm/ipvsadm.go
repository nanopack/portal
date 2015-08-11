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
	"bufio"
	"io"
	"errors"
	"os/exec"
	"github.com/pagodabox/na-router/config"
)

func Check() error {
	cmd := exec.Command("which", "ipvsadm")
	if err := cmd.Run(); err != nil {
		return errors.New("unable to find the ipvsadm command on the system")
	}
	return nil
}

func ListVips() ([]Vip, error) {
	return parse(parseAll, "ipvsadm", "-ln")
}

func AddVip(vip Vip) error {
	id := vip.getId()
	config.Log.Info("[NS_ROUTER] `ipvsadm -A -t %v -s wrr`", id)
	cmd := exec.Command("ipvsadm", "-A", "-t", id, "-s", "wrr")
	if err := cmd.Run(); err != nil {
		return err
	}
	config.Log.Info("what? %v",cmd)
	return nil
}

func GetVip(vip *Vip) error {
	return nil
}

func DeleteVip(vip Vip) error {
	return nil
}

func ListServers() ([]Server, error) {
	// var servers, err = parse(parseVip)
	return nil, nil
}

func AddServer(server Server) error {
	// vip := Vip{server.Vip, "", 0, nil}
	// if err := GetVip(&vip); err != nil {
	// 	return err
	// }
	// vipId := vip.getId()
	// serverId := server.getId()
	// cmd := exec.Command("ipvsadm", "-a", "-t", vipId, "-r", serverId, "-m", "-w", "0")
	
	// if err := cmd.Run(); err != nil {
	// 	return err
	// }
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

func parse(fun func(*bufio.Scanner) ([]Vip, error) , args... string) ([]Vip, error) {
	pipe, err := run(args)
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(pipe)
	scanner.Split(bufio.ScanWords)
	return fun(scanner)
}

func run(args []string) (io.ReadCloser, error) {
	cmd := exec.Command("ipvsadm", args...)
	return cmd.StdoutPipe()
}