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
package api

import (
	"encoding/json"
	"github.com/pagodabox/na-router/ipvsadm"
	"github.com/pagodabox/na-router/config"
	"net/http"
)

func init() {
	defaultApi.router.Post("/servers", traceRequest(serverCreate))
	defaultApi.router.Put("/servers/{server}", traceRequest(serverEnable))
	defaultApi.router.Get("/servers/{server}", traceRequest(serverGet))
	defaultApi.router.Delete("/servers/{server}", traceRequest(serverDelete))
	defaultApi.router.Get("/servers", traceRequest(serverList))
}

type (
	serversSlice struct {
		servers []ipvsadm.Server
	}
)

func (ss serversSlice) ToJson() ([]byte, error) {
	return json.Marshal(ss)
}

func serverCreate(res http.ResponseWriter, req *http.Request) {
	server := ipvsadm.Server{}
	err := parseBody(req, &server)
	if err == nil {
		err = ipvsadm.AddServer(server)
	}
	respond(201, err, server, res)
}

func serverList(res http.ResponseWriter, req *http.Request) {
	servers, err := ipvsadm.ListServers()
	config.Log.Info("[NA-ROUTER] list servers %v %v", servers, err)
	respond(200, err, serversSlice{servers}, res)
}

func serverGet(res http.ResponseWriter, req *http.Request) {
	// server := ipvsadm.Server{req.URL.Query().Get(":server"), "", 0}
	// err := ipvsadm.GetServer(&server)
	// config.Log.Info("[NA-ROUTER] get server %v %v", server, err)
	// respond(200, err, server, res)
}

func serverEnable(res http.ResponseWriter, req *http.Request) {
	// server := ipvsadm.Server{req.URL.Query().Get(":server"), "", 0}
	// err := ipvsadm.EnableServer(server)
	// config.Log.Info("[NA-ROUTER] enable server %v %v", server, err)
	// respond(200, err, nil, res)
}

func serverDelete(res http.ResponseWriter, req *http.Request) {
	// server := ipvsadm.Server{req.URL.Query().Get(":server"), "", 0}
	// err := ipvsadm.DeleteServer(server)
	// config.Log.Info("[NA-ROUTER] delete server %v %v", server, err)
	// respond(200, err, nil, res)
}
