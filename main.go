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
package main

import (
	"fmt"
	"os"

	"github.com/pagodabox/na-router/api"
	"github.com/pagodabox/na-router/config"
	"github.com/pagodabox/na-router/ipvsadm"
)

func main() {
	if err := ipvsadm.Load(); err != nil {
		fmt.Printf("ipvsadm can not be used: %v\n", err)
		os.Exit(1)
	}
	if err := api.Start(config.ListenAddress); err != nil {
		fmt.Printf("api exited abnormally: %v\n", err)
	}
}
