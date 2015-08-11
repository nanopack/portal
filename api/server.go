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
	"net/http"
)

func init() {
	defaultApi.router.Post("/vips/{vip}/servers/{server}", traceRequest(serverEnable))
	defaultApi.router.Post("/vips/{vip}/servers", traceRequest(serverCreate))
	defaultApi.router.Get("/vips/{vip}/servers/{server}", traceRequest(serverGet))
	defaultApi.router.Delete("/vips/{vip}/servers/{server}", traceRequest(serverDelete))
	defaultApi.router.Get("/vips/{vip}/servers", traceRequest(serverList))
}

type (
	serversSlice struct {
		servers []ipvsadm.Server
	}
	createServerBody struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	}
	enableBody struct {
		enabled bool
	}
)

func (ss serversSlice) ToJson() ([]byte, error) {
	return json.Marshal(ss)
}
func (eb *enableBody) FromJson(bytes []byte) error {
	return json.Unmarshal(bytes, eb)
}
func (cs *createServerBody) FromJson(bytes []byte) error {
	return json.Unmarshal(bytes, cs)
}

func serverCreate(res http.ResponseWriter, req *http.Request) {
	var server *ipvsadm.Server
	vid := req.URL.Query().Get(":vip")
	opts := createServerBody{}
	err := parseBody(req, &opts)
	if err == nil {
		server, err = ipvsadm.AddServer(vid, opts.Host, opts.Port)
	}
	respond(201, err, server, res)
}

func serverList(res http.ResponseWriter, req *http.Request) {
	vid := req.URL.Query().Get(":vip")
	servers, err := ipvsadm.ListServers(vid)
	respond(200, err, serversSlice{servers}, res)
}

func serverGet(res http.ResponseWriter, req *http.Request) {
	sid := req.URL.Query().Get(":server")
	vid := req.URL.Query().Get(":vip")
	server, err := ipvsadm.GetServer(vid, sid)
	respond(200, err, server, res)
}

func serverEnable(res http.ResponseWriter, req *http.Request) {
	sid := req.URL.Query().Get(":server")
	vid := req.URL.Query().Get(":vip")
	opts := enableBody{}
	err := parseBody(req, &opts)
	if err == nil {
		err = ipvsadm.EnableServer(vid, sid, opts.enabled)
	}
	respond(200, err, nil, res)
}

func serverDelete(res http.ResponseWriter, req *http.Request) {
	sid := req.URL.Query().Get(":server")
	vid := req.URL.Query().Get(":vip")
	err := ipvsadm.DeleteServer(vid, sid)
	respond(200, err, nil, res)
}
