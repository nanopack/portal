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
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/pat"
	"github.com/pagodabox/na-router/ipvsadm"
)

type (
	api struct {
		router *pat.Router
	}
	pong struct{}
)

func (p pong) ToJson() ([]byte, error) {
	return []byte("\"pong\""), nil
}

var (
	defaultApi = &api{pat.New()}
)

func init() {
	defaultApi.router.Get("/ping", traceRequest(pongRoute))
}

// pong to a ping.
func pongRoute(res http.ResponseWriter, req *http.Request) {
	respond(200, nil, pong{}, res)
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

// Start up the api and begin responding to requests. Blocking.
func Start(address string) error {
	return http.ListenAndServe(address, defaultApi.router)
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
		res.WriteHeader(500)
		res.Write([]byte(fmt.Sprintf("{\"error\":\"%v\"}", err)))
		return
	}
	res.WriteHeader(code)
	res.Write(bytes)
}

// Traces all routes going through the api.
func traceRequest(fn func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		// I need to do something to trace this request.
		fn(res, req)

	}
}
