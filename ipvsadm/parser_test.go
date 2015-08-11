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
	"strings"
	"testing"
)

var (
	header = `IP Virtual Server version 1.0.10 (size=4096)
Prot LocalAddress:Port Scheduler Flags
  -> RemoteAddress:Port           Forward Weight ActiveConn InActConn
`
	cmd = `TCP  212.204.230.98:80 wrr persistent 360
  -> 127.0.0.1:80           Masq   200    25         44
  -> 127.0.0.2:80            Masq   200    12         27
TCP  212.204.230.98:443 wrr persistent 123
  -> 127.0.0.1:443         Local   100    0          0
  -> 127.0.0.2:443          Local   100    0          0
`
)

func TestParser(test *testing.T) {
	all := []string{header, cmd}
	reader := strings.NewReader(strings.Join(all, ""))
	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanWords)
	vips, err := parseAll(scanner)
	if err != nil {
		test.Fatal(err)
	}
	assert(test, len(vips) == 2, "should have 2 vips, only have %v", len(vips))

	assert(test, vips[0].Host == "212.204.230.98", "incorrect host for vip 0: %v", vips[0].Host)
	assert(test, vips[1].Host == "212.204.230.98", "incorrect host for vip 1: %v", vips[1].Host)

	assert(test, len(vips[0].Servers) == 2, "wrong number of servers for vip 0: %v", len(vips[0].Servers))
	assert(test, len(vips[1].Servers) == 2, "wrong number of servers for vip 1: %v", len(vips[1].Servers))

}

func assert(test *testing.T, check bool, fmt string, args ...interface{}) {
	if !check {
		test.Logf(fmt, args...)
		test.FailNow()
	}
}
