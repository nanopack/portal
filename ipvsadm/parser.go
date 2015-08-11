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
	"bytes"
	"errors"
	"strconv"
	"strings"
)

var (
	EOFError       = errors.New("ipvsadm terminated prematurely")
	UnexpecedToken = errors.New("Unexpected Token")
)

// parse a vip line `TCP 127.0.0.1:1234 wrr` the cursor comes pre
// advanced so that scan.Bytes() would return TCP.
func parseVip(scan *bufio.Scanner) (*Vip, error) {

	// skip TCP
	if !scan.Scan() {
		return nil, EOFError
	}

	// record id
	id := scan.Text()
	split := strings.Split(id, ":")
	host := split[0]
	port, err := strconv.Atoi(split[1])
	if err != nil {
		return nil, err
	}

	// advance to next token
	if !scan.Scan() {
		return nil, EOFError
	}
	schedular := scan.Text()

	// skip the string literal "persistant"
	if !scan.Scan() {
		return nil, EOFError
	}

	if !scan.Scan() {
		return nil, EOFError
	}
	persistance, err := strconv.Atoi(scan.Text())
	if err != nil {
		return nil, err
	}

	return &Vip{host, port, schedular, persistance, nil}, nil
}

// parse a list of vips and real servers, which is the default output
// of ipvsadm -ln
func parseVips(scan *bufio.Scanner) ([]Vip, error) {
	var currentVip *Vip = nil
	vips := make([]Vip, 0)
	for scan.Scan() {
		switch {
		case bytes.Equal(scan.Bytes(), []byte("TCP")):
			newVip, err := parseVip(scan)
			if err != nil {
				return nil, err
			}
			if currentVip != nil {
				vips = append(vips, *currentVip)
			}
			currentVip = newVip
		case bytes.Equal(scan.Bytes(), []byte("->")):
			server, err := parseRealServer(scan)
			if err != nil {
				return nil, err
			}
			currentVip.Servers = append(currentVip.Servers, *server)
		default:
			return nil, UnexpecedToken
		}
	}
	if currentVip != nil {
		vips = append(vips, *currentVip)
	}
	return vips, nil
}

// parse a real server declaration line `-> 10.0.0.1:1234 Masq 0 0 0`
// the cursor comes pre advanced so that scan.Bytes() would return ->.
func parseRealServer(scan *bufio.Scanner) (*Server, error) {
	// advance to 10.0.0.1:1234
	scan.Scan()
	id := scan.Text()
	split := strings.Split(id, ":")
	host := split[0]
	port, err := strconv.Atoi(split[1])
	if err != nil {
		return nil, err
	}
	// advance to Masq
	if !scan.Scan() {
		return nil, EOFError
	}
	forwarder := scan.Text()

	if !scan.Scan() {
		return nil, EOFError
	}
	weight, err := strconv.Atoi(scan.Text())
	if err != nil {
		return nil, err
	}

	if !scan.Scan() {
		return nil, EOFError
	}
	innactive, err := strconv.Atoi(scan.Text())
	if err != nil {
		return nil, err
	}

	if !scan.Scan() {
		return nil, EOFError
	}
	active, err := strconv.Atoi(scan.Text())
	if err != nil {
		return nil, err
	}
	// leave 0 on so that the parseVips loop removes it.
	return &Server{host, port, forwarder, weight, innactive, active}, nil
}

// discard the entire ipvsadm header, leaves one token on so that the
// parse Vips loop removes it correctly.
func discardHeader(scan *bufio.Scanner) error {
	for i := 0; i < 16; i++ {
		if !scan.Scan() {
			return EOFError
		}
	}
	return nil
}

func parseAll(scan *bufio.Scanner) ([]Vip, error) {
	if err := discardHeader(scan); err != nil {
		return nil, err
	}
	return parseVips(scan)
}
