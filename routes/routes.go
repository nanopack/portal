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
package routes

import (
	"fmt"
	"github.com/pagodabox/na-api"
	"github.com/pagodabox/na-router/ipvsadm"
	"io/ioutil"
	"net/http"
)

func Init() {
	api.Router.Post("/vips", api.TraceRequest(vipCreate))
	api.Router.Get("/vips/{vip}", api.TraceRequest(vipGet))
	api.Router.Delete("/vips/{vip}", api.TraceRequest(vipDelete))
	api.Router.Get("/vips", api.TraceRequest(vipList))

	api.Router.Post("/vips/{vip}/servers/{server}", api.TraceRequest(serverEnable))
	api.Router.Post("/vips/{vip}/servers", api.TraceRequest(serverCreate))
	api.Router.Get("/vips/{vip}/servers/{server}", api.TraceRequest(serverGet))
	api.Router.Delete("/vips/{vip}/servers/{server}", api.TraceRequest(serverDelete))
	api.Router.Get("/vips/{vip}/servers", api.TraceRequest(serverList))
}

// read and parse the entire body
func parseBody(req *http.Request, output ipvsadm.FromJson) error {
	body, err := ioutil.ReadAll(req.Body)

	if err == nil {
		err = output.FromJson(body)
		req.Body.Close()
	}

	return err
}

// Send a response back to the client
func respond(code int, err error, body ipvsadm.ToJson, res http.ResponseWriter) {
	var bytes []byte
	if err == nil {
		if body == nil {
			bytes = []byte("{\"sucess\":true}")
		} else {
			bytes, err = body.ToJson()
		}
	}

	if err != nil {
		switch err {
		case ipvsadm.NotFound:
			res.WriteHeader(404)
		case ipvsadm.Conflict:
			res.WriteHeader(409)
		default:
			res.WriteHeader(500)
		}
		res.Write([]byte(fmt.Sprintf("{\"error\":\"%v\"}\n", err)))
		return
	}
	res.WriteHeader(code)
	res.Write(append(bytes, byte(15)))
}
