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
	"errors"
	"fmt"
	"io"
	"net"
	"os/exec"
)

var (
	Conflict       = errors.New("object already exists")
	NotFound       = errors.New("object was not found")
	DeleteFailed   = errors.New("object was not deleted")
	IpvsadmMissing = errors.New("unable to find the ipvsadm command on the system")

	// these are to allow a pluggable backend for testing, ipvsadm is
	// not needed to run the tests
	backend    = execute
	backendRun = run
)

func Load() error {
	if err := Check(); err != nil {
		return err
	}

	// NYI
	// populate the ipvsadm command with what was stored in the backup
	return nil
}
func Check() error {
	if err := backend("which", "ipvsadm"); err != nil {
		return IpvsadmMissing
	}
	return nil
}

func ListVips() ([]Vip, error) {
	return parse(parseAll, "ipvsadm", "-ln")
}

func AddVip(host string, port int) (*Vip, error) {
	id := fmt.Sprintf("%v:%v", host, port)
	// check if it already exists, this also validates the id
	vip, err := GetVip(id)
	if vip != nil {
		return vip, Conflict
	} else if err != NotFound {
		return nil, err
	}

	// create the vip
	if err := backend("ipvsadm", "-A", "-t", id, "-s", "wrr", "-p", "60"); err != nil {
		return nil, err // should be a custom error. this one may not make sense
	}

	backup()

	// double check that it was created
	return GetVip(id)
}

func GetVip(id string) (*Vip, error) {
	if err := validateId(id); err != nil {
		return nil, err
	}
	vips, err := ListVips()
	if err != nil {
		return nil, err
	}

	for _, vip := range vips {
		if vip.getId() == id {
			return &vip, nil
		}
	}
	return nil, NotFound
}

func DeleteVip(id string) error {
	_, err := GetVip(id)
	if err == NotFound {
		return NotFound
	} else if err != nil {
		return err
	}

	if err := backend("ipvsadm", "-D", "-t", id); err != nil {
		return err // I should return my own error here
	}

	_, err = GetVip(id)
	if err != NotFound {
		return err
	} else if err == nil {
		return DeleteFailed
	}

	return nil
}

func ListServers(vid string) ([]Server, error) {
	vip, err := GetVip(vid)
	if err != nil {
		return nil, err
	}
	return vip.Servers, nil
}

func AddServer(vid, host string, port int) (*Server, error) {
	id := fmt.Sprintf("%v:%v", host, port)
	server, err := GetServer(vid, id)
	if server != nil {
		return server, Conflict
	} else if err != NotFound {
		return nil, err
	}

	if err := backend("ipvsadm", "-a", "-t", vid, "-r", id, "-w", "0"); err != nil {
		return nil, err // I should return my own error here
	}

	backup()
	return GetServer(vid, id)
}

func GetServer(vid, id string) (*Server, error) {
	if err := validateId(id); err != nil {
		return nil, err
	}
	servers, err := ListServers(vid)
	if err != nil {
		return nil, err
	}
	for _, server := range servers {
		if server.getId() == id {
			return &server, nil
		}
	}
	return nil, NotFound
}

func EnableServer(vid, id string, enable bool) error {
	if _, err := GetServer(vid, id); err != nil {
		return err
	}

	var weight string
	if enable {
		weight = "100"
	} else {
		weight = "0"
	}

	if err := backend("ipvsadm", "-e", "-t", vid, "-r", id, "-w", weight); err != nil {
		return err // I should return my own error here
	}

	backup()
	return nil
}

func DeleteServer(vid, id string) error {
	if _, err := GetServer(vid, id); err != nil {
		return err
	}

	if err := backend("ipvsadm", "-d", "-t", vid, "-r", id); err != nil {
		return err // I should return my own error here
	}

	backup()
	return nil
}

func parse(fun func(*bufio.Scanner) ([]Vip, error), args ...string) ([]Vip, error) {
	pipe, err := backendRun(args)
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(pipe)
	scanner.Split(bufio.ScanWords)
	return fun(scanner)
}

func run(args []string) (io.ReadCloser, error) {
	cmd := exec.Command("ipvsadm", args...)
	pipe, err := cmd.StdoutPipe()
	cmd.Start()
	return pipe, err
}

func execute(exe string, args ...string) error {
	cmd := exec.Command(exe, args...)
	return cmd.Run()
}

func validateId(id string) error {
	_, _, err := net.SplitHostPort(id)
	return err
}

func backup() {
	//NYI
}
