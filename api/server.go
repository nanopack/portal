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
	"net/http"
	"github.com/pagodabox/na-router/ipvsadm"
)

func init() {
	defaultApi.router.Get("/servers", traceRequest(serverList))
	defaultApi.router.Post("/servers", traceRequest(serverCreate))
	defaultApi.router.Put("/servers/{server}", traceRequest(serverEnable))
	defaultApi.router.Get("/servers/{server}", traceRequest(serverGet))
	defaultApi.router.Delete("/servers/{server}", traceRequest(serverDelete))
}

type (
	serversSlice struct {
		servers []ipvsadm.Server
	}
)

func serverCreate(res http.ResponseWriter, req *http.Request) (routerResponse, error) {
	server := ipvsadm.Server{}
	err := parseBody(req, &server)
	if err == nil {
		err = ipvsadm.AddServer(server);
	}

	return nil, err
}

func serverList(res http.ResponseWriter, req *http.Request) (routerResponse, error) {
	servers, err := ipvsadm.ListServers()
	return serversSlice{servers}, err
}

func serverGet(res http.ResponseWriter, req *http.Request) (routerResponse, error) {
	server := ipvsadm.Server{req.URL.Query().Get(":server")}
	err := ipvsadm.GetServer(&server)
	return server, err
}

func serverEnable(res http.ResponseWriter, req *http.Request) (routerResponse, error) {
	server := ipvsadm.Server{req.URL.Query().Get(":server")}
	err := ipvsadm.EnableServer(server)
	return nil, err
}

func serverDelete(res http.ResponseWriter, req *http.Request) (routerResponse, error) {
	server := ipvsadm.Server{req.URL.Query().Get(":server")}
	err := ipvsadm.DeleteServer(server)
	return nil, err
}