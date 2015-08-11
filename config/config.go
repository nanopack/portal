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
package config

import (
	"github.com/jcelliott/lumber"
	"github.com/pagodabox/golang-hatchet"
)

var (
	ListenAddress string
	Log           hatchet.Logger
)

func init() {
	ListenAddress = "127.0.0.1:7750"
	Log = lumber.NewConsoleLogger(lumber.INFO)
}
