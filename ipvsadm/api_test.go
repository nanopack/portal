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
	"fmt"
	"io"
	"strconv"
	"strings"
	"testing"
)

var (
	storage = map[string]Vip{}
)

type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error { return nil }

func listAll() string {
	output := make([]string, 0)
	output = append(output, header)
	for _, vip := range storage {
		vipString := fmt.Sprintf("TCP %v:%v %v persistent %v", vip.Host, vip.Port, vip.Schedular, vip.Persistance)
		output = append(output, vipString)
		for _, server := range vip.Servers {
			vipString := fmt.Sprintf(" -> %v:%v %v %v %v %v", server.Host, server.Port, server.Forwarder, server.Weight, server.InactiveConnections, server.ActiveConnections)
			output = append(output, vipString)
		}
	}
	return strings.Join(output, "\n")
}
func addVip(vid, scheduler, persistent string) error {
	split := strings.Split(vid, ":")
	host := split[0]
	port, err := strconv.Atoi(split[1])
	if err != nil {
		return err
	}
	persist, err := strconv.Atoi(split[1])
	if err != nil {
		return err
	}
	vip := Vip{host, port, scheduler, persist, nil}
	storage[vid] = vip
	return nil
}
func removeVip(vip string) error {
	delete(storage, vip)
	return nil
}
func addServer(vid, sid, w string) error {
	split := strings.Split(sid, ":")
	host := split[0]
	port, err := strconv.Atoi(split[1])
	if err != nil {
		return err
	}

	weight, err := strconv.Atoi(w)
	if err != nil {
		return err
	}

	vip := storage[vid]
	server := Server{host, port, "Masq", weight, 0, 0}
	vip.Servers = append(vip.Servers, server)
	storage[vid] = vip
	return nil
}
func removeServer(vid, sid string) error {
	vip := storage[vid]
	for idx, server := range vip.Servers {
		if server.getId() == sid {
			vip.Servers = append(vip.Servers[:idx], vip.Servers[idx+1:]...)
			storage[vid] = vip
			return nil
		}
	}
	return nil
}

func fake(exec string, args ...string) error {
	fmt.Printf("%v\n", args)
	switch exec {
	case "which":
		return nil
	case "ipvsadm":
		switch args[0] {
		case "-A":
			return addVip(args[2], args[4], args[5])
		case "-D":
			return removeVip(args[2])
		case "-a":
			return addServer(args[2], args[4], args[6])
		case "-d":
			return removeServer(args[2], args[4])
		default:
			return errors.New("invalid option")
		}
	default:
		return errors.New("unknown command")
	}
}

func fakeRun(args []string) (io.ReadCloser, error) {
	reader := strings.NewReader(listAll())
	return nopCloser{reader}, nil
}

func TestApi(test *testing.T) {
	backend = fake
	backendRun = fakeRun

	vips, err := ListVips()
	assert(test, err == nil, "unable to list vips: %v", err)
	assert(test, len(vips) == 0, "wrong number of vips")

	vip, err := AddVip("127.0.0.1", 1234)
	assert(test, err == nil, "unable to add vip: %v", err)
	vip, err = AddVip("127.0.0.1", 1234)
	assert(test, err == Conflict, "should have got a conflict: %v", err)

	vip, err = GetVip("127.0.0.1:1234")
	assert(test, err == nil, "unable to get vip: %v", err)
	vip, err = GetVip("127.0.0.1:6666")
	assert(test, err != nil, "should not have got a vip: %v", vip)

	vips, err = ListVips()
	assert(test, err == nil, "unable to list vips: %v", err)
	assert(test, len(vips) == 1, "wrong number of vips")

	servers, err := ListServers("127.0.0.1:1234")
	assert(test, err == nil, "unable to get server: %v", err)
	assert(test, len(servers) == 0, "wrong number of vips")

	server, err := AddServer("127.0.0.1:1234", "10.0.0.1", 1234)
	assert(test, err == nil, "unable to add server: %v", err)
	server, err = AddServer("127.0.0.1:1234", "10.0.0.1", 1234)
	assert(test, err == Conflict, "should have got a conflict: %v", server.Host)

	server, err = GetServer("127.0.0.1:1234", "10.0.0.1:1234")
	assert(test, err == nil, "unable to get server: %v", err)

	servers, err = ListServers("127.0.0.1:1234")
	assert(test, err == nil, "unable to get server: %v", err)
	assert(test, len(servers) == 1, "wrong number of vips")

	err = DeleteServer("127.0.0.1:1234", "10.0.0.1:1234")
	assert(test, err == nil, "unable to delete server: %v", err)
	err = DeleteServer("127.0.0.1:1234", "10.0.0.1:1234")
	assert(test, err == NotFound, "should have got a not found: %v", err)

	err = DeleteVip("127.0.0.1:1234")
	assert(test, err == nil, "unable to delete vip: %v", err)
	err = DeleteVip("127.0.0.1:1234")
	assert(test, err == NotFound, "should have got a not found: %v", err)

	vips, err = ListVips()
	assert(test, err == nil, "unable to list vips: %v", err)
	assert(test, len(vips) == 0, "wrong number of vips")
}
