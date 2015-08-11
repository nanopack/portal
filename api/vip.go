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
	"github.com/pagodabox/na-router/config"
	"github.com/pagodabox/na-router/ipvsadm"
	"net/http"
)

func init() {
	defaultApi.router.Post("/vips", traceRequest(vipCreate))
	defaultApi.router.Get("/vips/{vip}", traceRequest(vipGet))
	defaultApi.router.Delete("/vips/{vip}", traceRequest(vipDelete))
	defaultApi.router.Get("/vips", traceRequest(vipList))
}

type (
	vipSlice struct {
		vips []ipvsadm.Vip
	}

	createVipBody struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	}
)

func (vs vipSlice) ToJson() ([]byte, error) {
	return json.Marshal(vs)
}
func (cv *createVipBody) FromJson(bytes []byte) error {
	return json.Unmarshal(bytes, cv)
}

func vipCreate(res http.ResponseWriter, req *http.Request) {
	var vip *ipvsadm.Vip
	opts := createVipBody{}
	err := parseBody(req, &opts)
	if err == nil {
		vip, err = ipvsadm.AddVip(opts.Host, opts.Port)
	}

	config.Log.Info("[NA-ROUTER] add vip %v %v", vip, err)
	respond(201, err, vip, res)
}

func vipList(res http.ResponseWriter, req *http.Request) {
	vips, err := ipvsadm.ListVips()
	config.Log.Info("[NA-ROUTER] list vips %v %v", vips, err)
	respond(200, err, vipSlice{vips}, res)
}

func vipGet(res http.ResponseWriter, req *http.Request) {
	vid := req.URL.Query().Get(":vip")
	vip, err := ipvsadm.GetVip(vid)
	config.Log.Info("[NA-ROUTER] get vip %v %v", vip, err)
	respond(200, err, vip, res)
}

func vipDelete(res http.ResponseWriter, req *http.Request) {
	vid := req.URL.Query().Get(":vip")
	err := ipvsadm.DeleteVip(vid)
	config.Log.Info("[NA-ROUTER] delete vip %v %v", vid, err)
	respond(200, err, nil, res)
}
