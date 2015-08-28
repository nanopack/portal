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
	"github.com/jcelliott/lumber"
	"github.com/pagodabox/na-api"
	"github.com/pagodabox/na-router/ipvsadm"
	"github.com/pagodabox/na-router/routes"
	"github.com/pagodabox/nanobox-config"
	"os"
	"strings"
)

func main() {
	configFile := ""
	if len(os.Args) > 1 && !strings.HasPrefix(os.Args[1], "-") {
		configFile = os.Args[1]
	}

	defaults := map[string]string{
		"listenAddress": "127.0.0.1:1234",
		"logLevel":      "INFO",
	}

	config.Load(defaults, configFile)
	config := config.Config

	api.Name = "UNKNOWN"
	level := lumber.LvlInt(config["log_level"])
	api.Logger = lumber.NewConsoleLogger(level)

	if err := ipvsadm.Load(); err != nil {
		api.Logger.Fatal("ipvsadm can not be used: %v\n", err)
		os.Exit(1)
	}

	routes.Init()
	if err := api.Start(config["listenAddress"]); err != nil {
		api.Logger.Fatal("api exited abnormally: %v\n", err)
	}
}
