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
	"github.com/pagodabox/na-router/ipvsadm"
	"net/http"
)

func init() {
	defaultApi.router.Get("/vips", traceRequest(vipList))
	defaultApi.router.Post("/vips", traceRequest(vipCreate))
	defaultApi.router.Get("/vips/{vip}", traceRequest(vipGet))
	defaultApi.router.Delete("/vips/{vip}", traceRequest(vipDelete))
}

type (
	vipSlice struct {
		vips []ipvsadm.Vip
	}
)

func vipCreate(res http.ResponseWriter, req *http.Request) (routerResponse, error) {
	vip := ipvsadm.Vip{}
	err := parseBody(req, &vip)
	if err == nil {
		err = ipvsadm.AddVip(vip)
	}

	return nil, err
}

func vipList(res http.ResponseWriter, req *http.Request) (routerResponse, error) {
	vips, err := ipvsadm.ListVips()
	return vipSlice{vips}, err
}

func vipGet(res http.ResponseWriter, req *http.Request) (routerResponse, error) {
	vip := ipvsadm.Vip{req.URL.Query().Get(":vip")}
	err := ipvsadm.GetVip(&vip)
	return vip, err
}

func vipDelete(res http.ResponseWriter, req *http.Request) (routerResponse, error) {
	vip := ipvsadm.Vip{req.URL.Query().Get(":vip")}
	err := ipvsadm.DeleteVip(vip)
	return nil, err
}
