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
	"bitbucket.org/nanobox/na-api"
	"bitbucket.org/nanobox/na-router/routes"
	"bitbucket.org/nanobox/nanobox-config"
	"github.com/jcelliott/lumber"
	"github.com/pagodabox/na-lvs"
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

	if err := lvs.Load(); err != nil {
		api.Logger.Fatal("ipvsadm can not be used: %v\n", err)
		os.Exit(1)
	}

	routes.Init()
	if err := api.Start(config["listenAddress"]); err != nil {
		api.Logger.Fatal("api exited abnormally: %v\n", err)
	}
}
