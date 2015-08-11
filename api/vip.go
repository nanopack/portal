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
	defaultApi.router.Post("/vips", traceRequest(vipCreate))
	defaultApi.router.Get("/vips/{vip}", traceRequest(vipGet))
	defaultApi.router.Delete("/vips/{vip}", traceRequest(vipDelete))
	defaultApi.router.Get("/vips", traceRequest(vipList))
}

type (
	vipSlice struct {
		vips []ipvsadm.Vip
	}
)

func (vs vipSlice) ToJson() ([]byte, error) {
	return json.Marshal(vs)
}

func vipCreate(res http.ResponseWriter, req *http.Request) {
	// vip := ipvsadm.Vip{}
	// err := parseBody(req, &vip)
	// if err == nil {
	// 	err = ipvsadm.AddVip(vip)
	// }
	// config.Log.Info("[NA-ROUTER] add vip %v %v", vip, err)
	// respond(201, err, vip, res)
}

func vipList(res http.ResponseWriter, req *http.Request) {
	vips, err := ipvsadm.ListVips()
	config.Log.Info("[NA-ROUTER] list vips %v %v", vips, err)
	respond(200, err, vipSlice{vips}, res)
}

func vipGet(res http.ResponseWriter, req *http.Request) {
	// vip := ipvsadm.Vip{req.URL.Query().Get(":vip"), "nil", 0, nil}
	// err := ipvsadm.GetVip(&vip)
	// config.Log.Info("[NA-ROUTER] get vip %v %v", vip, err)
	// respond(200, err, vip, res)
}

func vipDelete(res http.ResponseWriter, req *http.Request) {
	// vip := ipvsadm.Vip{req.URL.Query().Get(":vip"), "nil", 0, nil}
	// err := ipvsadm.DeleteVip(vip)
	// config.Log.Info("[NA-ROUTER] delete vip %v %v", vip, err)
	// respond(200, err, vip, res)
}
